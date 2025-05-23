package symphony

import (
	"reflect"
	"testing"
	"time"

	apiv1 "github.com/Azure/eno/api/v1"
	"github.com/Azure/eno/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestBasics(t *testing.T) {
	ctx := testutil.NewContext(t)
	mgr := testutil.NewManager(t)
	cli := mgr.GetClient()
	err := NewController(mgr.Manager)
	require.NoError(t, err)
	mgr.Start(t)

	// Create the symphony
	sym := &apiv1.Symphony{}
	sym.Name = "test-symphony"
	sym.Namespace = "default"
	sym.Spec.Bindings = []apiv1.Binding{
		{
			Key:      "foo",
			Resource: apiv1.ResourceBinding{Name: "test-resource-1"},
		},
		{
			Key:      "bar",
			Resource: apiv1.ResourceBinding{Name: "test-resource-2"},
		},
	}
	sym.Spec.Variations = []apiv1.Variation{
		{
			Synthesizer: apiv1.SynthesizerRef{Name: "foosynth"},
			Labels:      map[string]string{"foo": "bar"},
			Annotations: map[string]string{"foo": "bar"},
		},
		{
			Synthesizer: apiv1.SynthesizerRef{Name: "barsynth"},
			Labels:      map[string]string{"foo": "bar"},
			Annotations: map[string]string{"foo": "bar"},
		},
	}
	sym.Spec.SynthesisEnv = []apiv1.EnvVar{
		{
			Name:  "some_env",
			Value: "some-value",
		},
	}
	err = cli.Create(ctx, sym)
	require.NoError(t, err)

	// Exactly one composition should eventually be created for each synth
	testutil.Eventually(t, func() bool {
		comps := &apiv1.CompositionList{}
		err := cli.List(ctx, comps)
		if err != nil && len(comps.Items) < 2 {
			return false
		}
		synthsSeen := map[string]struct{}{}
		for _, comp := range comps.Items {
			comp := comp
			if !reflect.DeepEqual(sym.Spec.Bindings, comp.Spec.Bindings) ||
				!reflect.DeepEqual(comp.Annotations, map[string]string{"foo": "bar"}) ||
				!reflect.DeepEqual(comp.Labels, map[string]string{"foo": "bar"}) ||
				!reflect.DeepEqual(comp.Spec.SynthesisEnv, []apiv1.EnvVar{{Name: "some_env", Value: "some-value"}}) {
				t.Logf("composition %q was not replicated correctly", comp.Name)
				return false
			}
			synthsSeen[comp.Spec.Synthesizer.Name] = struct{}{}
		}
		if len(synthsSeen) > 2 {
			t.Logf("wrong number of synths seen: %d", len(synthsSeen))
			return false
		}
		for _, v := range sym.Spec.Variations {
			if _, ok := synthsSeen[v.Synthesizer.Name]; !ok {
				t.Logf("didn't see composition for synth %q", v.Synthesizer.Name)
				return false
			}
		}
		return true
	})

	// Mark each composition as reconciled
	comps := &apiv1.CompositionList{}
	err = cli.List(ctx, comps)
	require.NoError(t, err)
	for _, comp := range comps.Items {
		comp.Status.CurrentSynthesis = &apiv1.Synthesis{Reconciled: ptr.To(metav1.Now()), ObservedCompositionGeneration: comp.Generation}
		err = cli.Status().Update(ctx, &comp)
		require.NoError(t, err)
	}

	// The symphony should eventually be marked as reconciled
	testutil.Eventually(t, func() bool {
		err := cli.Get(ctx, client.ObjectKeyFromObject(sym), sym)
		return err == nil && sym.Status.Reconciled != nil
	})

	// Update the bindings and prove the new bindings are replicated to the compositions
	err = retry.RetryOnConflict(testutil.Backoff, func() error {
		cli.Get(ctx, client.ObjectKeyFromObject(sym), sym)
		sym.Spec.Bindings = []apiv1.Binding{{Key: "new-binding", Resource: apiv1.ResourceBinding{Name: "foo"}}}
		return cli.Update(ctx, sym)
	})
	require.NoError(t, err)

	testutil.Eventually(t, func() bool {
		comps := &apiv1.CompositionList{}
		err := cli.List(ctx, comps)
		if err != nil && len(comps.Items) < 2 {
			return false
		}
		for _, comp := range comps.Items {
			if !reflect.DeepEqual(sym.Spec.Bindings, comp.Spec.Bindings) {
				t.Logf("composition %q has incorrect bindings", comp.Name)
				return false
			}
		}
		return true
	})

	// Because the compositions have been updated, the symphony should be marked as not reconciled
	testutil.Eventually(t, func() bool {
		err := cli.Get(ctx, client.ObjectKeyFromObject(sym), sym)
		return err == nil && sym.Status.Reconciled == nil
	})

	// Update the labels and annotations and prove they're replicated to the compositions
	err = retry.RetryOnConflict(testutil.Backoff, func() error {
		cli.Get(ctx, client.ObjectKeyFromObject(sym), sym)
		for i := range sym.Spec.Variations {
			sym.Spec.Variations[i].Labels = map[string]string{"foo": "baz"}
			sym.Spec.Variations[i].Annotations = map[string]string{"foo": "baz"}
		}
		return cli.Update(ctx, sym)
	})
	require.NoError(t, err)

	testutil.Eventually(t, func() bool {
		comps := &apiv1.CompositionList{}
		err := cli.List(ctx, comps)
		if err != nil && len(comps.Items) < 2 {
			return false
		}
		for _, comp := range comps.Items {
			if comp.Labels == nil ||
				comp.Labels["foo"] != "baz" ||
				comp.Annotations == nil ||
				comp.Annotations["foo"] != "baz" {
				t.Logf("composition %q doesn't have the expected labels and annotations", comp.Name)
				return false
			}
		}
		return true
	})

	// Test deletion
	require.NoError(t, cli.Delete(ctx, sym))
	testutil.Eventually(t, func() bool {
		return errors.IsNotFound(cli.Get(ctx, client.ObjectKeyFromObject(sym), sym))
	})

	comps = &apiv1.CompositionList{}
	err = cli.List(ctx, comps)
	require.NoError(t, err)
	assert.Len(t, comps.Items, 0)
}

// TestSymphonyDuplicateCleanup proves that the newest compositions are deleted if multiple exist for a given synthesizer.
func TestSymphonyDuplicateCleanup(t *testing.T) {
	ctx := testutil.NewContext(t)
	cli := testutil.NewClient(t)
	s := &symphonyController{client: cli}

	sym := &apiv1.Symphony{}
	sym.Name = "test"
	sym.Namespace = "default"
	sym.Spec.Variations = []apiv1.Variation{{Synthesizer: apiv1.SynthesizerRef{Name: "foo"}}}
	require.NoError(t, cli.Create(ctx, sym))

	comp := apiv1.Composition{}
	now := metav1.Now()
	comp.CreationTimestamp = metav1.NewTime(now.Add(time.Second))
	comp.Name = "foo"
	comp.Spec.Synthesizer.Name = "foo"

	comp2 := apiv1.Composition{}
	comp2.CreationTimestamp = now
	comp2.Name = "bar"
	comp2.Spec.Synthesizer.Name = "foo"

	comps := &apiv1.CompositionList{Items: []apiv1.Composition{comp, comp2}}
	_, err := s.reconcileReverse(ctx, sym, comps)
	require.EqualError(t, err, `deleting duplicate composition: compositions.eno.azure.io "bar" not found`)
}

func TestBuildStatus(t *testing.T) {
	c := &symphonyController{}

	t.Run("empty", func(t *testing.T) {
		symph := &apiv1.Symphony{}
		symph.Generation = 123

		comps := &apiv1.CompositionList{}

		status := c.buildStatus(symph, comps)
		assert.Equal(t, apiv1.SymphonyStatus{
			ObservedGeneration: symph.Generation,
		}, status)
	})

	t.Run("one ready", func(t *testing.T) {
		readyTime := ptr.To(metav1.NewTime(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)))

		symph := &apiv1.Symphony{
			Spec: apiv1.SymphonySpec{
				Variations: []apiv1.Variation{
					{Synthesizer: apiv1.SynthesizerRef{Name: "foo"}},
				},
			},
		}
		comps := &apiv1.CompositionList{}
		comps.Items = []apiv1.Composition{{
			Spec: apiv1.CompositionSpec{
				Synthesizer: apiv1.SynthesizerRef{Name: "foo"},
			},
			Status: apiv1.CompositionStatus{
				CurrentSynthesis: &apiv1.Synthesis{
					Ready: readyTime,
				},
			},
		}}

		status := c.buildStatus(symph, comps)
		assert.Equal(t, apiv1.SymphonyStatus{
			Ready: readyTime,
		}, status)
	})

	t.Run("one ready, one not ready", func(t *testing.T) {
		readyTime := ptr.To(metav1.NewTime(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)))

		symph := &apiv1.Symphony{
			Spec: apiv1.SymphonySpec{
				Variations: []apiv1.Variation{
					{Synthesizer: apiv1.SynthesizerRef{Name: "foo"}},
					{Synthesizer: apiv1.SynthesizerRef{Name: "bar"}},
				},
			},
		}
		comps := &apiv1.CompositionList{}
		comps.Items = []apiv1.Composition{
			{
				Spec: apiv1.CompositionSpec{
					Synthesizer: apiv1.SynthesizerRef{Name: "foo"},
				},
				Status: apiv1.CompositionStatus{
					CurrentSynthesis: &apiv1.Synthesis{
						Ready: readyTime,
					},
				},
			},
			{
				Spec: apiv1.CompositionSpec{
					Synthesizer: apiv1.SynthesizerRef{Name: "bar"},
				},
				Status: apiv1.CompositionStatus{
					CurrentSynthesis: &apiv1.Synthesis{
						Ready: nil,
					},
				},
			},
		}

		status := c.buildStatus(symph, comps)
		assert.Equal(t, apiv1.SymphonyStatus{}, status)
	})

	t.Run("two ready", func(t *testing.T) {
		readyTime := ptr.To(metav1.NewTime(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)))

		symph := &apiv1.Symphony{
			Spec: apiv1.SymphonySpec{
				Variations: []apiv1.Variation{
					{Synthesizer: apiv1.SynthesizerRef{Name: "foo"}},
					{Synthesizer: apiv1.SynthesizerRef{Name: "bar"}},
				},
			},
		}
		comps := &apiv1.CompositionList{}
		comps.Items = []apiv1.Composition{
			{
				Spec: apiv1.CompositionSpec{
					Synthesizer: apiv1.SynthesizerRef{Name: "foo"},
				},
				Status: apiv1.CompositionStatus{
					CurrentSynthesis: &apiv1.Synthesis{
						Ready: readyTime,
					},
				},
			},
			{
				Spec: apiv1.CompositionSpec{
					Synthesizer: apiv1.SynthesizerRef{Name: "bar"},
				},
				Status: apiv1.CompositionStatus{
					CurrentSynthesis: &apiv1.Synthesis{
						Ready: ptr.To(metav1.NewTime(readyTime.Add(-time.Second))),
					},
				},
			},
		}

		status := c.buildStatus(symph, comps)
		assert.Equal(t, apiv1.SymphonyStatus{
			Ready: readyTime,
		}, status)
	})
}

func TestGetBindings(t *testing.T) {
	tcs := []struct {
		name             string
		symph            apiv1.Symphony
		variation        apiv1.Variation
		expectedBindings []apiv1.Binding
	}{
		{
			name: "just symphony bindings",
			symph: apiv1.Symphony{
				Spec: apiv1.SymphonySpec{
					Bindings: []apiv1.Binding{
						{Key: "bnd-1"},
					},
				},
			},
			expectedBindings: []apiv1.Binding{
				{Key: "bnd-1"},
			},
		},
		{
			name: "just variation bindings",
			variation: apiv1.Variation{
				Bindings: []apiv1.Binding{
					{Key: "bnd-1"},
				},
			},
			expectedBindings: []apiv1.Binding{
				{Key: "bnd-1"},
			},
		},
		{
			name: "symphony and variation bindings",
			variation: apiv1.Variation{
				Bindings: []apiv1.Binding{
					{Key: "bnd-1"},
				},
			},
			symph: apiv1.Symphony{
				Spec: apiv1.SymphonySpec{
					Bindings: []apiv1.Binding{
						{Key: "bnd-2"},
					},
				},
			},
			expectedBindings: []apiv1.Binding{
				{Key: "bnd-1"},
				{Key: "bnd-2"},
			},
		},
		{
			name: "symphony and variation bindings with dups",
			variation: apiv1.Variation{
				Bindings: []apiv1.Binding{
					{Key: "bnd-1"},
					{Key: "bnd-1"},
				},
			},
			symph: apiv1.Symphony{
				Spec: apiv1.SymphonySpec{
					Bindings: []apiv1.Binding{
						{Key: "bnd-2"},
						{Key: "bnd-2"},
					},
				},
			},
			expectedBindings: []apiv1.Binding{
				{Key: "bnd-1"},
				{Key: "bnd-2"},
			},
		},
		{
			name: "variation takes precedence over symphony",
			variation: apiv1.Variation{
				Bindings: []apiv1.Binding{
					{Key: "bnd-1", Resource: apiv1.ResourceBinding{Name: "from-variation"}},
				},
			},
			symph: apiv1.Symphony{
				Spec: apiv1.SymphonySpec{
					Bindings: []apiv1.Binding{
						{Key: "bnd-1", Resource: apiv1.ResourceBinding{Name: "from-symphony"}},
					},
				},
			},
			expectedBindings: []apiv1.Binding{
				{Key: "bnd-1", Resource: apiv1.ResourceBinding{Name: "from-variation"}},
			},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			actualBindings := getBindings(&tc.symph, &tc.variation)
			require.ElementsMatch(t, tc.expectedBindings, actualBindings)
		})
	}
}

func TestGetSynthesisEnv(t *testing.T) {
	tcs := []struct {
		name        string
		symph       apiv1.Symphony
		variation   apiv1.Variation
		expectedEnv []apiv1.EnvVar
	}{
		{
			name: "just symphony env",
			symph: apiv1.Symphony{
				Spec: apiv1.SymphonySpec{
					SynthesisEnv: []apiv1.EnvVar{{Name: "var1", Value: "val1"}},
				},
			},
			expectedEnv: []apiv1.EnvVar{{Name: "var1", Value: "val1"}},
		},
		{
			name: "just variation env",
			variation: apiv1.Variation{
				SynthesisEnv: []apiv1.EnvVar{{Name: "var1", Value: "val1"}},
			},
			expectedEnv: []apiv1.EnvVar{{Name: "var1", Value: "val1"}},
		},
		{
			name: "symphony and variation env",
			variation: apiv1.Variation{
				SynthesisEnv: []apiv1.EnvVar{{Name: "var1", Value: "val1"}},
			},
			symph: apiv1.Symphony{
				Spec: apiv1.SymphonySpec{
					SynthesisEnv: []apiv1.EnvVar{{Name: "var2", Value: "val2"}},
				},
			},
			expectedEnv: []apiv1.EnvVar{{Name: "var1", Value: "val1"}, {Name: "var2", Value: "val2"}},
		},
		{
			name: "symphony and variation env with dups",
			variation: apiv1.Variation{
				SynthesisEnv: []apiv1.EnvVar{{Name: "var1", Value: "val1"}, {Name: "val2", Value: "var2"}},
			},
			symph: apiv1.Symphony{
				Spec: apiv1.SymphonySpec{
					SynthesisEnv: []apiv1.EnvVar{{Name: "var1", Value: "val1"}, {Name: "val2", Value: "var2"}},
				},
			},
			expectedEnv: []apiv1.EnvVar{{Name: "var1", Value: "val1"}, {Name: "val2", Value: "var2"}},
		},
		{
			name: "variation takes precedence over symphony",
			variation: apiv1.Variation{
				SynthesisEnv: []apiv1.EnvVar{{Name: "var1", Value: "comp-override"}, {Name: "val2", Value: "var2"}},
			},
			symph: apiv1.Symphony{
				Spec: apiv1.SymphonySpec{
					SynthesisEnv: []apiv1.EnvVar{{Name: "var1", Value: "val1"}, {Name: "val2", Value: "var2"}},
				},
			},
			expectedEnv: []apiv1.EnvVar{{Name: "var1", Value: "comp-override"}, {Name: "val2", Value: "var2"}},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			actualEnv := getSynthesisEnv(&tc.symph, &tc.variation)
			require.ElementsMatch(t, tc.expectedEnv, actualEnv)
		})
	}
}

func TestCoalesceMetadata(t *testing.T) {
	tests := []struct {
		name           string
		variation      *apiv1.Variation
		existing       *apiv1.Composition
		expectedLabels map[string]string
		expectedAnnos  map[string]string
		expectedChange bool
	}{
		{
			name: "no labels or annotations - no change",
			variation: &apiv1.Variation{
				Labels:      map[string]string{},
				Annotations: map[string]string{},
			},
			existing: &apiv1.Composition{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      map[string]string{},
					Annotations: map[string]string{},
				},
			},
			expectedLabels: map[string]string{},
			expectedAnnos:  map[string]string{},
			expectedChange: false,
		},
		{
			name: "new label added",
			variation: &apiv1.Variation{
				Labels: map[string]string{
					"label1": "value1",
				},
				Annotations: map[string]string{},
			},
			existing: &apiv1.Composition{
				ObjectMeta: metav1.ObjectMeta{},
			},
			expectedLabels: map[string]string{
				"label1": "value1",
			},
			expectedAnnos:  map[string]string{},
			expectedChange: true,
		},
		{
			name: "existing label modified",
			variation: &apiv1.Variation{
				Labels: map[string]string{
					"label1": "newValue",
				},
				Annotations: map[string]string{},
			},
			existing: &apiv1.Composition{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"label1": "oldValue",
					},
					Annotations: map[string]string{},
				},
			},
			expectedLabels: map[string]string{
				"label1": "newValue",
			},
			expectedAnnos:  map[string]string{},
			expectedChange: true,
		},
		{
			name: "new annotation added",
			variation: &apiv1.Variation{
				Labels: map[string]string{},
				Annotations: map[string]string{
					"anno1": "value1",
				},
			},
			existing: &apiv1.Composition{
				ObjectMeta: metav1.ObjectMeta{},
			},
			expectedLabels: map[string]string{},
			expectedAnnos: map[string]string{
				"anno1": "value1",
			},
			expectedChange: true,
		},
		{
			name: "label and annotation modified",
			variation: &apiv1.Variation{
				Labels: map[string]string{
					"label1": "newValue",
				},
				Annotations: map[string]string{
					"anno1": "newValue",
				},
			},
			existing: &apiv1.Composition{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"label1": "oldValue",
					},
					Annotations: map[string]string{
						"anno1": "oldValue",
					},
				},
			},
			expectedLabels: map[string]string{
				"label1": "newValue",
			},
			expectedAnnos: map[string]string{
				"anno1": "newValue",
			},
			expectedChange: true,
		},
		{
			name: "no change in labels and annotations",
			variation: &apiv1.Variation{
				Labels: map[string]string{
					"label1": "value1",
				},
				Annotations: map[string]string{
					"anno1": "value1",
				},
			},
			existing: &apiv1.Composition{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"label1": "value1",
					},
					Annotations: map[string]string{
						"anno1": "value1",
					},
				},
			},
			expectedLabels: map[string]string{
				"label1": "value1",
			},
			expectedAnnos: map[string]string{
				"anno1": "value1",
			},
			expectedChange: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changed := coalesceMetadata(tt.variation, tt.existing)

			assert.Equal(t, tt.expectedChange, changed)
			assert.Equal(t, tt.expectedLabels, tt.existing.Labels)
			assert.Equal(t, tt.expectedAnnos, tt.existing.Annotations)
		})
	}
}
