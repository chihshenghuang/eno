package resource

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"maps"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	apiv1 "github.com/Azure/eno/api/v1"
	"github.com/Azure/eno/internal/readiness"
	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/jsonmergepatch"
	"sigs.k8s.io/controller-runtime/pkg/client"
	smdschema "sigs.k8s.io/structured-merge-diff/v4/schema"
	"sigs.k8s.io/structured-merge-diff/v4/typed"
	"sigs.k8s.io/structured-merge-diff/v4/value"
)

var patchGVK = schema.GroupVersionKind{
	Group:   "eno.azure.io",
	Version: "v1",
	Kind:    "Patch",
}

// Ref refers to a specific synthesized resource.
type Ref struct {
	Name, Namespace, Group, Kind string
}

func (r *Ref) String() string {
	return fmt.Sprintf("(%s.%s)/%s/%s", r.Group, r.Kind, r.Namespace, r.Name)
}

// ManifestRef references a particular resource manifest within a resource slice.
type ManifestRef struct {
	Slice types.NamespacedName
	Index int // position of this manifest within the slice
}

// Resource is the controller's internal representation of a single resource out of a ResourceSlice.
type Resource struct {
	Ref               Ref
	ManifestDeleted   bool
	ManifestRef       ManifestRef
	ManifestHash      []byte
	ReconcileInterval *metav1.Duration
	GVK               schema.GroupVersionKind
	ReadinessChecks   readiness.Checks
	Patch             jsonpatch.Patch
	DisableUpdates    bool
	ReadinessGroup    int
	Labels            map[string]string

	// DefinedGroupKind is set on CRDs to represent the resource type they define.
	DefinedGroupKind *schema.GroupKind

	value            value.Value
	latestKnownState atomic.Pointer[apiv1.ResourceState]
}

func (r *Resource) Deleted(comp *apiv1.Composition) bool {
	return (comp.DeletionTimestamp != nil && !comp.ShouldOrphanResources()) || r.ManifestDeleted || (r.Patch != nil && r.patchSetsDeletionTimestamp())
}

func (r *Resource) Unstructured() *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: r.value.Unstructured().(map[string]any)}
}

func (r *Resource) State() *apiv1.ResourceState { return r.latestKnownState.Load() }

func (r *Resource) NeedsToBePatched(current *unstructured.Unstructured) bool {
	if r.Patch == nil || current == nil {
		return false
	}

	curjson, err := current.MarshalJSON()
	if err != nil {
		return false
	}

	patchedjson, err := r.Patch.Apply(curjson)
	if err != nil {
		return false
	}

	patched := &unstructured.Unstructured{}
	err = patched.UnmarshalJSON(patchedjson)
	if err != nil {
		return false
	}

	return !equality.Semantic.DeepEqual(current, patched)
}

func (r *Resource) patchSetsDeletionTimestamp() bool {
	if r.Patch == nil {
		return false
	}

	// Apply the patch to a minimally-viable unstructured resource.
	// This is needed to satisfy the validation logic of the unstructured json parser, which requires a kind/apiVersion.
	patchedjson, err := r.Patch.Apply([]byte(`{"apiVersion": "eno.azure.io/v1", "kind":"PatchPlaceholder", "metadata":{}}`))
	if err != nil {
		return false
	}

	patched := map[string]any{}
	err = json.Unmarshal(patchedjson, &patched)
	if err != nil {
		return false
	}

	dt, _, _ := unstructured.NestedString(patched, "metadata", "deletionTimestamp")
	return dt != ""
}

type SchemaGetter interface {
	Get(ctx context.Context, gvk schema.GroupVersionKind) (typeref *smdschema.TypeRef, schem *smdschema.Schema, err error)
}

// Merge performs a three-way merge between the resource, it's old/previous Resource, and the current state.
// Falls back to a non-structured three-way merge if the SchemaGetter returns a nil TypeRef.
func (r *Resource) Merge(ctx context.Context, old *Resource, current *unstructured.Unstructured, sg SchemaGetter) (*unstructured.Unstructured, bool /* typed */, error) {
	typeref, schem, err := sg.Get(ctx, r.GVK)
	if err != nil {
		return nil, false, fmt.Errorf("looking up schema: %w", err)
	}

	// Naive three-way merge for unknown types
	if typeref == nil {
		currentJS, err := current.MarshalJSON()
		if err != nil {
			return nil, false, fmt.Errorf("encoding current state: %w", err)
		}

		var prevJS []byte
		if old != nil {
			prevJS, err = old.Unstructured().MarshalJSON()
			if err != nil {
				return nil, false, fmt.Errorf("encoding old state: %w", err)
			}
		}

		expectedJS, err := r.Unstructured().MarshalJSON()
		if err != nil {
			return nil, false, fmt.Errorf("encoding expected state: %w", err)
		}

		patch, err := jsonmergepatch.CreateThreeWayJSONMergePatch(prevJS, expectedJS, currentJS)
		if err != nil {
			return nil, false, fmt.Errorf("building merge patch: %w", err)
		}
		patchedJSON, err := jsonpatch.MergePatch(currentJS, patch)
		if err != nil {
			return nil, false, fmt.Errorf("applying merge patch: %w", err)
		}

		patched := &unstructured.Unstructured{}
		err = patched.UnmarshalJSON(patchedJSON)
		if err != nil {
			return nil, false, fmt.Errorf("parsing patched resource: %w", err)
		}

		if equality.Semantic.DeepEqual(current, patched) {
			return nil, false, nil
		}
		return patched, false, nil
	}

	// Convert to SMD values
	currentVal := value.NewValueInterface(current.Object)
	typedNew := typed.AsTypedUnvalidated(r.value, schem, *typeref)
	typedCurrent := typed.AsTypedUnvalidated(currentVal, schem, *typeref)

	// Merge properties that are set in the new state onto the current state
	merged, err := typedCurrent.Merge(typedNew)
	if err != nil {
		return nil, false, fmt.Errorf("merging new state into current: %w", err)
	}

	// Prune properties that were present in the old state but not the new
	if old != nil {
		typedOld, err := typed.AsTyped(old.value, schem, *typeref)
		if err != nil {
			return nil, false, fmt.Errorf("converting old version to typed: %w", err)
		}
		toOld, err := typedOld.Compare(typedNew)
		if err != nil {
			return nil, false, fmt.Errorf("comparing new and old states: %w", err)
		}
		merged = merged.RemoveItems(toOld.Removed)
	}

	// Bail out if no changes are required
	cmp, err := merged.Compare(typedCurrent)
	if err == nil && cmp.IsSame() {
		return nil, true, nil // no changes
	}

	copy := &unstructured.Unstructured{Object: merged.AsValue().Unstructured().(map[string]any)}
	return copy, true, nil
}

func NewResource(ctx context.Context, slice *apiv1.ResourceSlice, index int) (*Resource, error) {
	logger := logr.FromContextOrDiscard(ctx)
	resource := slice.Spec.Resources[index]
	res := &Resource{
		ManifestDeleted: resource.Deleted,
		ManifestRef: ManifestRef{
			Slice: types.NamespacedName{
				Namespace: slice.Namespace,
				Name:      slice.Name,
			},
			Index: index,
		},
	}

	hash := fnv.New64()
	hash.Write([]byte(resource.Manifest))
	res.ManifestHash = hash.Sum(nil)

	parsed := &unstructured.Unstructured{}
	err := parsed.UnmarshalJSON([]byte(resource.Manifest))
	if err != nil {
		return nil, fmt.Errorf("invalid json: %w", err)
	}

	// Prune out the status/creation time.
	// This is a pragmatic choice to make Eno behave in expected ways for synthesizers written using client-go structs,
	// which set metadata.creationTime=null and status={}.
	if parsed.Object != nil {
		delete(parsed.Object, "status")
		parsed.SetCreationTimestamp(metav1.Time{})
	}

	res.value = value.NewValueInterface(parsed.Object)
	gvk := parsed.GroupVersionKind()
	res.GVK = gvk
	res.Ref.Name = parsed.GetName()
	res.Ref.Namespace = parsed.GetNamespace()
	res.Ref.Group = parsed.GroupVersionKind().Group
	res.Ref.Kind = parsed.GetKind()
	logger = logger.WithValues("resourceKind", parsed.GetKind(), "resourceName", parsed.GetName(), "resourceNamespace", parsed.GetNamespace())

	if res.Ref.Name == "" || res.Ref.Kind == "" || parsed.GetAPIVersion() == "" {
		return nil, fmt.Errorf("missing name, kind, or apiVersion")
	}

	if res.GVK == patchGVK {
		obj := struct {
			Patch patchMeta `json:"patch"`
		}{}
		err = json.Unmarshal([]byte(resource.Manifest), &obj)
		if err != nil {
			return nil, fmt.Errorf("parsing patch json: %w", err)
		}
		gv, err := schema.ParseGroupVersion(obj.Patch.APIVersion)
		if err != nil {
			return nil, fmt.Errorf("parsing patch apiVersion: %w", err)
		}
		res.GVK.Group = gv.Group
		res.GVK.Version = gv.Version
		res.GVK.Kind = obj.Patch.Kind
		res.Patch = obj.Patch.Ops
	}

	if res.GVK.Group == "apiextensions.k8s.io" && res.GVK.Kind == "CustomResourceDefinition" {
		res.DefinedGroupKind = &schema.GroupKind{}
		res.DefinedGroupKind.Group, _, _ = unstructured.NestedString(parsed.Object, "spec", "group")
		res.DefinedGroupKind.Kind, _, _ = unstructured.NestedString(parsed.Object, "spec", "names", "kind")
	}

	res.Labels = maps.Clone(parsed.GetLabels())
	anno := parsed.GetAnnotations()
	if anno == nil {
		anno = map[string]string{}
	}

	const reconcileIntervalKey = "eno.azure.io/reconcile-interval"
	if str, ok := anno[reconcileIntervalKey]; ok {
		reconcileInterval, err := time.ParseDuration(str)
		if anno[reconcileIntervalKey] != "" && err != nil {
			logger.V(0).Info("invalid reconcile interval - ignoring")
		}
		res.ReconcileInterval = &metav1.Duration{Duration: reconcileInterval}
	}

	const disableUpdatesKey = "eno.azure.io/disable-updates"
	res.DisableUpdates = anno[disableUpdatesKey] == "true"

	const readinessGroupKey = "eno.azure.io/readiness-group"
	if str, ok := anno[readinessGroupKey]; ok {
		rg, err := strconv.Atoi(str)
		if err != nil {
			logger.V(0).Info("invalid readiness group - ignoring")
		} else {
			res.ReadinessGroup = rg
		}
	}

	for key, value := range anno {
		if !strings.HasPrefix(key, "eno.azure.io/readiness") || key == readinessGroupKey {
			continue
		}

		name := strings.TrimPrefix(key, "eno.azure.io/readiness-")
		if name == "eno.azure.io/readiness" {
			name = "default"
		}

		check, err := readiness.ParseCheck(value)
		if err != nil {
			logger.Error(err, "invalid cel expression")
			continue
		}
		check.Name = name
		res.ReadinessChecks = append(res.ReadinessChecks, check)
	}
	sort.Slice(res.ReadinessChecks, func(i, j int) bool { return res.ReadinessChecks[i].Name < res.ReadinessChecks[j].Name })

	parsed.SetAnnotations(pruneMetadata(parsed.GetAnnotations()))
	parsed.SetLabels(pruneMetadata(parsed.GetLabels()))

	return res, nil
}

func pruneMetadata(m map[string]string) map[string]string {
	maps.DeleteFunc(m, func(key string, value string) bool {
		return strings.HasPrefix(key, "eno.azure.io/")
	})
	if len(m) == 0 {
		m = nil
	}
	return m
}

// Less returns true when r < than.
// Used to establish determinstic ordering for conflicting resources.
func (r *Resource) Less(than *Resource) bool {
	return bytes.Compare(r.ManifestHash, than.ManifestHash) < 0
}

type patchMeta struct {
	APIVersion string          `json:"apiVersion"`
	Kind       string          `json:"kind"`
	Ops        jsonpatch.Patch `json:"ops"`
}

func NewInputRevisions(obj client.Object, refKey string) *apiv1.InputRevisions {
	ir := apiv1.InputRevisions{
		Key:             refKey,
		ResourceVersion: obj.GetResourceVersion(),
	}
	if rev, _ := strconv.Atoi(obj.GetAnnotations()["eno.azure.io/revision"]); rev != 0 {
		ir.Revision = &rev
	}
	if rev, _ := strconv.ParseInt(obj.GetAnnotations()["eno.azure.io/synthesizer-generation"], 10, 64); rev != 0 {
		ir.SynthesizerGeneration = &rev
	}
	return &ir
}
