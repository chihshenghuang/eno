package resource

import (
	"context"
	"os"
	"sort"
	"testing"
	"time"

	apiv1 "github.com/Azure/eno/api/v1"
	openapi_v2 "github.com/google/gnostic-models/openapiv2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/kube-openapi/pkg/schemaconv"
	"k8s.io/kube-openapi/pkg/util/proto"
	smdschema "sigs.k8s.io/structured-merge-diff/v4/schema"
)

var newResourceTests = []struct {
	Name     string
	Manifest string
	Assert   func(*testing.T, *Resource)
}{
	{
		Name: "configmap",
		Manifest: `{
			"apiVersion": "v1",
			"kind": "ConfigMap",
			"metadata": {
				"name": "foo",
				"annotations": {
					"foo": "bar",
					"eno.azure.io/reconcile-interval": "10s",
					"eno.azure.io/readiness-group": "250",
					"eno.azure.io/readiness": "true",
					"eno.azure.io/readiness-test": "false",
					"eno.azure.io/disable-updates": "true"
				}
			}
		}`,
		Assert: func(t *testing.T, r *Resource) {
			assert.Equal(t, schema.GroupVersionKind{Version: "v1", Kind: "ConfigMap"}, r.GVK)
			assert.Len(t, r.ReadinessChecks, 2)
			assert.Equal(t, time.Second*10, r.ReconcileInterval.Duration)
			assert.Equal(t, Ref{
				Name:      "foo",
				Namespace: "",
				Group:     "",
				Kind:      "ConfigMap",
			}, r.Ref)
			assert.True(t, r.DisableUpdates)
			assert.Equal(t, int(250), r.ReadinessGroup)
		},
	},
	{
		Name: "zero-readiness-group",
		Manifest: `{
			"apiVersion": "v1",
			"kind": "ConfigMap",
			"metadata": {
				"name": "foo",
				"annotations": {
					"eno.azure.io/readiness-group": "0"
				}
			}
		}`,
		Assert: func(t *testing.T, r *Resource) {
			assert.Equal(t, int(0), r.ReadinessGroup)
		},
	},
	{
		Name: "negative-readiness-group",
		Manifest: `{
			"apiVersion": "v1",
			"kind": "ConfigMap",
			"metadata": {
				"name": "foo",
				"annotations": {
					"eno.azure.io/readiness-group": "-10"
				}
			}
		}`,
		Assert: func(t *testing.T, r *Resource) {
			assert.Equal(t, int(-10), r.ReadinessGroup)
		},
	},
	{
		Name: "deployment",
		Manifest: `{
			"apiVersion": "apps/v1",
			"kind": "Deployment",
			"metadata": {
				"name": "foo",
				"namespace": "bar"
			}
		}`,
		Assert: func(t *testing.T, r *Resource) {
			assert.Equal(t, schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}, r.GVK)
			assert.Len(t, r.ReadinessChecks, 0)
			assert.Nil(t, r.ReconcileInterval)
			assert.Equal(t, Ref{
				Name:      "foo",
				Namespace: "bar",
				Group:     "apps",
				Kind:      "Deployment",
			}, r.Ref)
		},
	},
	{
		Name: "patch",
		Manifest: `{
			"apiVersion": "eno.azure.io/v1",
			"kind": "Patch",
			"metadata": {
				"name": "foo",
				"namespace": "bar"
			},
			"patch": {
				"apiVersion": "v1",
				"kind": "ConfigMap",
				"ops": [
					{ "op": "add", "path": "/data/foo", "value": "bar" }
				]
			}
		}`,
		Assert: func(t *testing.T, r *Resource) {
			assert.Equal(t, schema.GroupVersionKind{Version: "v1", Kind: "ConfigMap"}, r.GVK)
			assert.Len(t, r.Patch, 1)
			assert.False(t, r.patchSetsDeletionTimestamp())
		},
	},
	{
		Name: "deletionPatch",
		Manifest: `{
			"apiVersion": "eno.azure.io/v1",
			"kind": "Patch",
			"metadata": {
				"name": "foo",
				"namespace": "bar"
			},
			"patch": {
				"apiVersion": "v1",
				"kind": "ConfigMap",
				"ops": [
					{"op": "add", "path": "/metadata/deletionTimestamp", "value": "anything"}
				]
			}
		}`,
		Assert: func(t *testing.T, r *Resource) {
			assert.Equal(t, schema.GroupVersionKind{Version: "v1", Kind: "ConfigMap"}, r.GVK)
			assert.Len(t, r.Patch, 1)
			assert.True(t, r.patchSetsDeletionTimestamp())
		},
	},
	{
		Name: "crd",
		Manifest: `{
			"apiVersion": "apiextensions.k8s.io/v1",
			"kind": "CustomResourceDefinition",
			"metadata": {
				"name": "foo"
			},
			"spec": {
				"group": "test-group",
				"names": {
					"kind": "TestKind"
				}
			}
		}`,
		Assert: func(t *testing.T, r *Resource) {
			assert.Equal(t, schema.GroupVersionKind{Group: "apiextensions.k8s.io", Version: "v1", Kind: "CustomResourceDefinition"}, r.GVK)
			assert.Equal(t, &schema.GroupKind{Group: "test-group", Kind: "TestKind"}, r.DefinedGroupKind)
		},
	},
	{
		Name: "empty-crd",
		Manifest: `{
			"apiVersion": "apiextensions.k8s.io/v1",
			"kind": "CustomResourceDefinition",
			"metadata": {
				"name": "foo"
			}
		}`,
		Assert: func(t *testing.T, r *Resource) {
			assert.Equal(t, schema.GroupVersionKind{Group: "apiextensions.k8s.io", Version: "v1", Kind: "CustomResourceDefinition"}, r.GVK)
			assert.Equal(t, &schema.GroupKind{Group: "", Kind: ""}, r.DefinedGroupKind)
		},
	},
	{
		Name: "extra-metadata",
		Manifest: `{
			"apiVersion": "v1",
			"kind": "ConfigMap",
			"metadata": {
				"name": "foo",
				"labels": {
					"test-label": "should not be pruned",
					"eno.azure.io/extra-label": "should be pruned"
				},
				"annotations": {
					"test-annotation": "should not be pruned",
					"eno.azure.io/extra-annotation": "should be pruned",
					"eno.azure.io/reconcile-interval": "10s"
				}
			}
		}`,
		Assert: func(t *testing.T, r *Resource) {
			assert.Equal(t, time.Second*10, r.ReconcileInterval.Duration)
			assert.Equal(t, &unstructured.Unstructured{
				Object: map[string]any{
					"apiVersion": "v1",
					"kind":       "ConfigMap",
					"metadata": map[string]any{
						"name":        "foo",
						"annotations": map[string]any{"test-annotation": "should not be pruned"},
						"labels":      map[string]any{"test-label": "should not be pruned"},
					},
				},
			}, r.Unstructured())
		},
	},
	{
		Name: "empty-metadata",
		Manifest: `{
			"apiVersion": "v1",
			"kind": "ConfigMap",
			"metadata": {
				"name": "foo",
				"labels": {
					"eno.azure.io/extra-label": "should be pruned"
				},
				"annotations": {
					"eno.azure.io/extra-annotation": "should be pruned"
				}
			}
		}`,
		Assert: func(t *testing.T, r *Resource) {
			assert.Equal(t, &unstructured.Unstructured{
				Object: map[string]any{
					"apiVersion": "v1",
					"kind":       "ConfigMap",
					"metadata": map[string]any{
						"name": "foo",
					},
				},
			}, r.Unstructured())
		},
	},
	{
		Name: "labels",
		Manifest: `{
			"apiVersion": "v1",
			"kind": "ConfigMap",
			"metadata": {
				"name": "foo",
				"labels": {
					"test-label": "label-value",
					"eno.azure.io/extra-label": "should be pruned from resource"
				}
			}
		}`,
		Assert: func(t *testing.T, r *Resource) {
			assert.Equal(t, &unstructured.Unstructured{
				Object: map[string]any{
					"apiVersion": "v1",
					"kind":       "ConfigMap",
					"metadata": map[string]any{
						"name": "foo",
						"labels": map[string]any{
							"test-label": "label-value",
						},
					},
				},
			}, r.Unstructured())

			assert.Equal(t, map[string]string{
				"test-label":               "label-value",
				"eno.azure.io/extra-label": "should be pruned from resource",
			}, r.Labels)
		},
	},
}

func TestNewResource(t *testing.T) {
	ctx := context.Background()
	for _, tc := range newResourceTests {
		t.Run(tc.Name, func(t *testing.T) {
			r, err := NewResource(ctx, &apiv1.ResourceSlice{
				Spec: apiv1.ResourceSliceSpec{
					Resources: []apiv1.Manifest{{Manifest: tc.Manifest}},
				},
			}, 0)
			require.NoError(t, err)
			tc.Assert(t, r)
		})
	}
}

func TestMergeBasics(t *testing.T) {
	testMergeBasics(t, "io.k8s.api.apps.v1.Deployment")
}

func TestMergeBasicsNoSchema(t *testing.T) {
	testMergeBasics(t, "")
}

func testMergeBasics(t *testing.T, schemaName string) {
	t.Helper()
	ctx := context.Background()

	sg := newTestSchemaGetter(t, schemaName)

	newSlice := &apiv1.ResourceSlice{
		Spec: apiv1.ResourceSliceSpec{
			Resources: []apiv1.Manifest{{
				Manifest: `{
				  "apiVersion": "apps/v1",
				  "kind": "Deployment",
				  "metadata": {
				    "name": "foo"
				  },
				  "spec": {
				    "replicas": 2,
					"template": {
					  "spec": {
					    "serviceAccountName": "updated"
					  }
				    }
				  }
				}`,
			}},
		},
	}
	newState, err := NewResource(ctx, newSlice, 0)
	require.NoError(t, err)

	oldSlice := &apiv1.ResourceSlice{
		Spec: apiv1.ResourceSliceSpec{
			Resources: []apiv1.Manifest{{
				Manifest: `{
				  "apiVersion": "apps/v1",
				  "kind": "Deployment",
				  "metadata": {
				    "name": "foo"
				  },
				  "spec": {
				    "strategy": {
						"type": "RollingUpdate"
				    },
					"template": {
					  "spec": {
					    "serviceAccountName": "original"
					  }
				    }
				  }
				}`,
			}},
		},
	}
	oldState, err := NewResource(ctx, oldSlice, 0)
	require.NoError(t, err)

	current := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata":   map[string]any{"name": "foo", "resourceVersion": "1"},
		"spec": map[string]any{
			"selector": map[string]any{
				"matchLabels": map[string]any{"foo": "bar"},
			},
			"strategy": map[string]any{"type": "RollingUpdate"},
			"template": map[string]any{
				"spec": map[string]any{
					"serviceAccountName": "original",
				},
			},
		},
	}}

	expected := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata":   map[string]any{"name": "foo", "resourceVersion": "1"},
		"spec": map[string]any{
			"replicas": int64(2),
			"selector": map[string]any{
				"matchLabels": map[string]any{"foo": "bar"},
			},
			"template": map[string]any{
				"spec": map[string]any{
					"serviceAccountName": "updated",
				},
			},
		},
	}}

	// Apply changes
	merged, typed, err := newState.Merge(ctx, oldState, current, sg)
	require.NoError(t, err)
	assert.Equal(t, schemaName != "", typed)
	require.Equal(t, expected, merged)

	expectedWithoutOldState := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata":   map[string]any{"name": "foo", "resourceVersion": "1"},
		"spec": map[string]any{
			"replicas": int64(2),
			"strategy": map[string]any{
				"type": "RollingUpdate",
			},
			"selector": map[string]any{
				"matchLabels": map[string]any{"foo": "bar"},
			},
			"template": map[string]any{
				"spec": map[string]any{
					"serviceAccountName": "updated",
				},
			},
		},
	}}

	// Supports nil oldState
	merged, typed, err = newState.Merge(ctx, nil, current, sg)
	require.NoError(t, err)
	assert.Equal(t, schemaName != "", typed)
	require.Equal(t, expectedWithoutOldState, merged)

	// Check idempotence
	expected.SetResourceVersion("2")                                            // ignore resource version change
	expected.Object["status"] = map[string]any{"availableReplicas": float64(2)} // ignore status change
	merged, typed, err = newState.Merge(ctx, oldState, expected, sg)
	require.NoError(t, err)
	assert.Equal(t, schemaName != "", typed)

	if schemaName == "" {
		assert.NotNil(t, merged)
	} else {
		assert.Nil(t, merged)
	}
}

func TestResourceOrdering(t *testing.T) {
	resources := []*Resource{
		{ManifestHash: []byte("a")},
		{},
		{ManifestHash: []byte("b")},
		{},
		{ManifestHash: []byte("c")},
	}
	sort.Slice(resources, func(i, j int) bool {
		return resources[i].Less(resources[j])
	})

	assert.Equal(t, []*Resource{
		{},
		{},
		{ManifestHash: []byte("a")},
		{ManifestHash: []byte("b")},
		{ManifestHash: []byte("c")},
	}, resources)
}

type testSchemaGetter struct {
	name   string
	schema *smdschema.Schema
}

func (t *testSchemaGetter) Get(ctx context.Context, gvk schema.GroupVersionKind) (typeref *smdschema.TypeRef, schem *smdschema.Schema, err error) {
	if t.name == "" {
		return nil, nil, nil
	}
	return &smdschema.TypeRef{NamedType: &t.name}, t.schema, nil
}

func newTestSchemaGetter(t *testing.T, name string) *testSchemaGetter {
	oapiJS, err := os.ReadFile("fixtures/openapi.json")
	require.NoError(t, err)

	doc := &openapi_v2.Document{}
	err = protojson.Unmarshal(oapiJS, doc)
	require.NoError(t, err)

	models, err := proto.NewOpenAPIData(doc)
	require.NoError(t, err)

	schem, err := schemaconv.ToSchema(models)
	require.NoError(t, err)

	return &testSchemaGetter{schema: schem, name: name}
}
