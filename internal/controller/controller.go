package controller

import (
	"fmt"

	"github.com/nginxinc/nginx-gateway-kubernetes/internal/config"
	gw "github.com/nginxinc/nginx-gateway-kubernetes/internal/implementations/gateway"
	gc "github.com/nginxinc/nginx-gateway-kubernetes/internal/implementations/gatewayclass"
	gcfg "github.com/nginxinc/nginx-gateway-kubernetes/internal/implementations/gatewayconfig"
	nginxgwv1alpha1 "github.com/nginxinc/nginx-gateway-kubernetes/pkg/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-kubernetes/pkg/sdk"

	"k8s.io/apimachinery/pkg/runtime"
	ctlr "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	gatewayv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = gatewayv1alpha2.AddToScheme(scheme)
	_ = nginxgwv1alpha1.AddToScheme(scheme)
}

func Start(conf config.Config) error {
	logger := conf.Logger

	options := manager.Options{
		Scheme: scheme,
	}

	mgr, err := manager.New(ctlr.GetConfigOrDie(), options)
	if err != nil {
		return fmt.Errorf("cannot build runtime manager: %w", err)
	}

	err = sdk.RegisterGatewayController(mgr, gw.NewGatewayImplementation(conf))
	if err != nil {
		return fmt.Errorf("cannot register gateway implementation: %w", err)
	}
	err = sdk.RegisterGatewayClassController(mgr, gc.NewGatewayClassImplementation(conf))
	if err != nil {
		return fmt.Errorf("cannot register gatewayclass implementation: %w", err)
	}
	err = sdk.RegisterGatewayConfigController(mgr, gcfg.NewGatewayConfigImplementation(conf))
	if err != nil {
		return fmt.Errorf("cannot register gatewayconfig implementation: %w", err)
	}

	ctx := ctlr.SetupSignalHandler()

	logger.Info("Starting manager")
	return mgr.Start(ctx)
}
