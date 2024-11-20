package synthesis

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	apiv1 "github.com/Azure/eno/api/v1"
	"github.com/Azure/eno/internal/manager"
	"github.com/go-logr/logr"
)

type sliceController struct {
	client client.Client
}

// sliceController check if the resource slice is deleted but it is still present in the composition status.
// If yes, then it will update the composition status to trigger re-synthesis process.
func NewSliceController(mgr ctrl.Manager) error {
	c := &sliceController{
		client: mgr.GetClient(),
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named("synthesisSliceController").
		Watches(&apiv1.ResourceSlice{}, newSliceHandler()).
		WithLogConstructor(manager.NewLogConstructor(mgr, "sliceController")).
		Complete(c)
}

func (s *sliceController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logr.FromContextOrDiscard(ctx)

	comp := &apiv1.Composition{}
	err := s.client.Get(ctx, req.NamespacedName, comp)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("getting composition: %w", err)
	}

	syn := &apiv1.Synthesizer{}
	syn.Name = comp.Spec.Synthesizer.Name
	err = s.client.Get(ctx, client.ObjectKeyFromObject(syn), syn)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(fmt.Errorf("gettting synthesizer: %w", err))
	}

	logger = logger.WithValues("compositionGeneration", comp.Generation,
		"compositionName", comp.Name,
		"compositionNamespace", comp.Namespace,
		"synthesisID", comp.Status.GetCurrentSynthesisUUID())

	// skip if the composition is not eligible for resynthesis
	if comp.NotEligibleForResynthesis(syn) {
		return ctrl.Result{}, nil
	}

	for _, ref := range comp.Status.CurrentSynthesis.ResourceSlices {
		slice := &apiv1.ResourceSlice{}
		slice.Name = ref.Name
		slice.Namespace = comp.Namespace
		err := s.client.Get(ctx, client.ObjectKeyFromObject(slice), slice)
		if errors.IsNotFound(err) {
			// The resource slice should not be deleted if it is still referenced by the composition
			comp.Status.PendingResynthesis = ptr.To(metav1.Now())
			err = s.client.Status().Update(ctx, comp)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("swapping compisition state: %w", err)
			}
			return ctrl.Result{}, nil
		}

		if client.IgnoreNotFound(err) != nil {
			return ctrl.Result{}, client.IgnoreNotFound(fmt.Errorf("getting resource slice: %w", err))
		}
	}

	return ctrl.Result{}, nil
}

func newSliceHandler() handler.EventHandler {
	apply := func(rli workqueue.RateLimitingInterface, obj client.Object) {
		owner := metav1.GetControllerOf(obj)
		if owner == nil {
			// no need to check the deleted resource slice which doesn't have owner
			return
		}
		// pass the composition name to the request
		rli.Add(reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      owner.Name,
				Namespace: obj.GetNamespace(),
			},
		})
	}

	return &handler.Funcs{
		CreateFunc: func(ctx context.Context, ce event.CreateEvent, rli workqueue.RateLimitingInterface) {
			// No need to hanlde creation event
		},
		UpdateFunc: func(ctx context.Context, ue event.UpdateEvent, rli workqueue.RateLimitingInterface) {
			// No need to handle update event
		},
		DeleteFunc: func(ctx context.Context, de event.DeleteEvent, rli workqueue.RateLimitingInterface) {
			apply(rli, de.Object)
		},
	}
}
