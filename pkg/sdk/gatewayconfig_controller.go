package sdk

import (
	"context"

	nginxgwv1alpha1 "github.com/nginxinc/nginx-gateway-kubernetes/pkg/apis/gateway/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctlr "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type gatewayConfigReconciler struct {
	client.Client
	scheme *runtime.Scheme
	impl   GatewayConfigImpl
}

func RegisterGatewayConfigController(mgr manager.Manager, impl GatewayConfigImpl) error {
	r := &gatewayConfigReconciler{
		Client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
		impl:   impl,
	}

	return ctlr.NewControllerManagedBy(mgr).
		For(&nginxgwv1alpha1.GatewayConfig{}).
		Complete(r)
}

func (r *gatewayConfigReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	log := log.FromContext(ctx).WithValues("gatewayconfig", req.Name)
	log.V(3).Info("Reconciling GatewayConfig")

	found := true
	var gcfg nginxgwv1alpha1.GatewayConfig
	err := r.Get(ctx, req.NamespacedName, &gcfg)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			log.Error(err, "Failed to get GatewayConfig")
			return reconcile.Result{}, err
		}
		found = false
	}

	if !found {
		r.impl.Remove(req.Name)
		return reconcile.Result{}, nil
	}

	r.impl.Upsert(&gcfg)
	return reconcile.Result{}, nil
}
