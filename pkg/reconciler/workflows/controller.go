package workflows

import (
	"context"

	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"

	workflowsinformer "github.com/tektoncd/experimental/workflows/pkg/client/injection/informers/workflows/v1alpha1/workflow"
	workflowsreconciler "github.com/tektoncd/experimental/workflows/pkg/client/injection/reconciler/workflows/v1alpha1/workflow"
)

// NewController creates a Reconciler and returns the result of NewImpl.
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {
	// Obtain an informer to both the main and child resources. These will be started by
	// the injection framework automatically. They'll keep a cached representation of the
	// cluster's state of the respective resource at all times.
	workflowsInformer := workflowsinformer.Get(ctx)
	// podInformer := podinformer.Get(ctx)

	r := &Reconciler{
		// The client will be needed to create/delete Pods via the API.
		// kubeclient: kubeclient.Get(ctx),
		// A lister allows read-only access to the informer's cache, allowing us to cheaply
		// read pod data.
		// podLister: podInformer.Lister(),
	}
	impl := workflowsreconciler.NewImpl(ctx, r)

	// Listen for events on the main resource and enqueue themselves.
	workflowsInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	// Listen for events on the child resources and enqueue the owner of them.
	// podInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
	// 	// FilterFunc: controller.FilterController(&v1alpha1.Workflow{}),
	// 	Handler: controller.HandleAll(impl.EnqueueControllerOf),
	// })

	return impl
}
