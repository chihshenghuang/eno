package synthesis

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiv1 "github.com/Azure/eno/api/v1"
	"github.com/Azure/eno/internal/testutil"
)

func TestSliceRecreation(t *testing.T) {
	ctx := testutil.NewContext(t)
	mgr := testutil.NewManager(t)
	require.NoError(t, NewSliceController(mgr.Manager))
	mgr.Start(t)

	// Create resource slice
	readyTime := metav1.Now()
	slice := &apiv1.ResourceSlice{}
	slice.Name = "test-slice"
	slice.Namespace = "default"
	slice.Spec.Resources = []apiv1.Manifest{{Manifest: "{}"}}
	slice.Status.Resources = []apiv1.ResourceState{{Ready: &readyTime, Reconciled: true}}
	require.NoError(t, mgr.GetClient().Create(ctx, slice))
	require.NoError(t, mgr.GetClient().Status().Update(ctx, slice))

	// Create composition
	comp := &apiv1.Composition{}
	comp.Name = "test-comp"
	comp.Namespace = "default"
	require.NoError(t, mgr.GetClient().Create(ctx, comp))

	// Synthesis has completed with no error
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		err := mgr.GetClient().Get(ctx, client.ObjectKeyFromObject(comp), comp)
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		comp.Status.CurrentSynthesis = &apiv1.Synthesis{
			Synthesized:                   ptr.To(metav1.Now()),
			ObservedCompositionGeneration: comp.Generation,
			ResourceSlices:                []*apiv1.ResourceSliceRef{{Name: "test-slice"}},
		}
		return mgr.GetClient().Status().Update(ctx, comp)
	})
	require.NoError(t, err)

	// Check resource slice is existed
	require.NoError(t, mgr.GetClient().Get(ctx, client.ObjectKeyFromObject(slice), slice))

	// Delete the resource slice
	require.NoError(t, mgr.GetClient().Delete(ctx, slice))
	// Check the resource slice is deleted
	testutil.Eventually(t, func() bool {
		err := mgr.GetClient().Get(ctx, client.ObjectKeyFromObject(slice), slice)
		if errors.IsNotFound(err) {
			return true
		}
		return false
	})
	// s := &sliceController{client: mgr.GetClient()}
	// req := ctrl.Request{NamespacedName: types.NamespacedName{Namespace: comp.Namespace, Name: comp.Name}}
	// _, err = s.Reconcile(ctx, req)
	require.NoError(t, err)

	testutil.Eventually(t, func() bool {
		err := mgr.GetClient().Get(ctx, client.ObjectKeyFromObject(slice), slice)
		if err != nil {
			return false
		}
		return true
	})
}
