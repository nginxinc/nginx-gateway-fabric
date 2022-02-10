package manager

import (
	"fmt"

	"github.com/nginxinc/nginx-gateway-kubernetes/internal/config"
	"github.com/nginxinc/nginx-gateway-kubernetes/internal/events"
	gw "github.com/nginxinc/nginx-gateway-kubernetes/internal/implementations/gateway"
	gc "github.com/nginxinc/nginx-gateway-kubernetes/internal/implementations/gatewayclass"
	gcfg "github.com/nginxinc/nginx-gateway-kubernetes/internal/implementations/gatewayconfig"
	hr "github.com/nginxinc/nginx-gateway-kubernetes/internal/implementations/httproute"
	"github.com/nginxinc/nginx-gateway-kubernetes/internal/state"
	nginxgwv1alpha1 "github.com/nginxinc/nginx-gateway-kubernetes/pkg/apis/gateway/v1alpha1"
	"github.com/nginxinc/nginx-gateway-kubernetes/pkg/sdk"

	"k8s.io/apimachinery/pkg/runtime"
	ctlr "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	gatewayv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

var scheme = runtime.NewScheme()

func init() {
	_ = gatewayv1alpha2.AddToScheme(scheme)
	_ = nginxgwv1alpha1.AddToScheme(scheme)
}

func Start(cfg config.Config) error {
	logger := cfg.Logger

	options := manager.Options{
		Scheme: scheme,
	}

	eventCh := make(chan interface{})

	mgr, err := manager.New(ctlr.GetConfigOrDie(), options)
	if err != nil {
		return fmt.Errorf("cannot build runtime manager: %w", err)
	}

	err = sdk.RegisterGatewayController(mgr, gw.NewGatewayImplementation(cfg))
	if err != nil {
		return fmt.Errorf("cannot register gateway implementation: %w", err)
	}
	err = sdk.RegisterGatewayClassController(mgr, gc.NewGatewayClassImplementation(cfg))
	if err != nil {
		return fmt.Errorf("cannot register gatewayclass implementation: %w", err)
	}
	err = sdk.RegisterGatewayConfigController(mgr, gcfg.NewGatewayConfigImplementation(cfg))
	if err != nil {
		return fmt.Errorf("cannot register gatewayconfig implementation: %w", err)
	}
	err = sdk.RegisterHTTPRouteController(mgr, hr.NewHTTPRouteImplementation(cfg, eventCh))
	if err != nil {
		return fmt.Errorf("cannot register httproute implementation: %w", err)
	}

	conf := state.NewConfiguration(cfg.GatewayCtlrName, state.NewRealClock())
	mainCtrl := events.NewEventLoop(conf, eventCh, cfg.Logger)

	err = mgr.Add(mainCtrl)
	if err != nil {
		return fmt.Errorf("cannot register main controller")
	}

	ctx := ctlr.SetupSignalHandler()

	logger.Info("Starting manager")
	return mgr.Start(ctx)
}
