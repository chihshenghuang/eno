package aggregation

import (
	"context"
	"fmt"

	apiv1 "github.com/Azure/eno/api/v1"
	"github.com/Azure/eno/internal/manager"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/equality"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type symphonyController struct {
	client client.Client
}

func NewSymphonyController(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1.Symphony{}).
		Owns(&apiv1.Composition{}).
		WithLogConstructor(manager.NewLogConstructor(mgr, "symphonyAggregationController")).
		Complete(&symphonyController{
			client: mgr.GetClient(),
		})
}

func (c *symphonyController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logr.FromContextOrDiscard(ctx)

	symph := &apiv1.Symphony{}
	err := c.client.Get(ctx, req.NamespacedName, symph)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	logger = logger.WithValues("symphonyName", symph.Name, "symphonyNamespace", symph.Namespace)

	existing := &apiv1.CompositionList{}
	err = c.client.List(ctx, existing, client.InNamespace(symph.Namespace), client.MatchingFields{
		manager.IdxCompositionsBySymphony: symph.Name,
	})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("listing existing compositions: %w", err)
	}

	newStatus, ok := c.buildStatus(symph, existing)
	if !ok {
		return ctrl.Result{}, nil
	}

	copy := symph.DeepCopy()
	copy.Status = newStatus
	if err := c.client.Status().Patch(ctx, copy, client.MergeFrom(symph)); err != nil {
		return ctrl.Result{}, fmt.Errorf("updating status: %w", err)
	}

	logger.V(1).Info("aggregated composition status into symphony")
	return ctrl.Result{}, nil
}

func (c *symphonyController) buildStatus(symph *apiv1.Symphony, comps *apiv1.CompositionList) (apiv1.SymphonyStatus, bool) {
	newStatus := apiv1.SymphonyStatus{ObservedGeneration: symph.Generation, Synthesizers: symph.Status.Synthesizers}

	synthMap := map[string]struct{}{}
	// Find the max values
	for _, comp := range comps.Items {
		if comp.Status.CurrentSynthesis == nil {
			continue
		}
		if newStatus.Ready.Before(comp.Status.CurrentSynthesis.Ready) || newStatus.Ready == nil {
			newStatus.Ready = comp.Status.CurrentSynthesis.Ready
		}
		if newStatus.Reconciled.Before(comp.Status.CurrentSynthesis.Reconciled) || newStatus.Reconciled == nil {
			newStatus.Reconciled = comp.Status.CurrentSynthesis.Reconciled
		}
		if newStatus.Synthesized.Before(comp.Status.CurrentSynthesis.Synthesized) || newStatus.Synthesized == nil {
			newStatus.Synthesized = comp.Status.CurrentSynthesis.Synthesized
		}
	}

	// Filter any values where one or more composition hasn't reached the corresponding state
	for _, comp := range comps.Items {
		if comp.Status.CurrentSynthesis == nil || comp.Status.CurrentSynthesis.ObservedCompositionGeneration != comp.Generation || comp.DeletionTimestamp != nil {
			newStatus.Ready = nil
			newStatus.Reconciled = nil
			newStatus.Synthesized = nil
			return newStatus, false
		}
		if comp.Status.CurrentSynthesis.Ready == nil {
			newStatus.Ready = nil
		}
		if comp.Status.CurrentSynthesis.Reconciled == nil {
			newStatus.Reconciled = nil
		}
		if comp.Status.CurrentSynthesis.Synthesized == nil {
			newStatus.Synthesized = nil
		}

		synthMap[comp.Spec.Synthesizer.Name] = struct{}{}
	}

	// It isn't safe to sync until we've seen a composition for every synthesizer in the symphony.
	// Otherwise the status might be incorrect until the next tick of the loop.
	//
	// Technically it can still be incorrect in the case of duplicates, but this is very unlikely
	// since the duplicate would have to live long enough to be synthesized.
	for _, v := range symph.Spec.Variations {
		if _, ok := synthMap[v.Synthesizer.Name]; !ok {
			return newStatus, false
		}
	}

	return newStatus, !equality.Semantic.DeepEqual(newStatus, symph.Status)
}
