package manager

import (
	"fmt"
	"time"

	"github.com/nginxinc/nginx-gateway-kubernetes/internal/config"
	"github.com/nginxinc/nginx-gateway-kubernetes/internal/events"
	gw "github.com/nginxinc/nginx-gateway-kubernetes/internal/implementations/gateway"
	gc "github.com/nginxinc/nginx-gateway-kubernetes/internal/implementations/gatewayclass"
	gcfg "github.com/nginxinc/nginx-gateway-kubernetes/internal/implementations/gatewayconfig"
	hr "github.com/nginxinc/nginx-gateway-kubernetes/internal/implementations/httproute"
	svc "github.com/nginxinc/nginx-gateway-kubernetes/internal/implementations/service"
	"github.com/nginxinc/nginx-gateway-kubernetes/internal/state"
	"github.com/nginxinc/nginx-gateway-kubernetes/internal/status"
	nginxgwv1alpha1 "github.com/nginxinc/nginx-gateway-kubernetes/pkg/apis/gateway/v1alpha1"
	"github.com/nginxinc/nginx-gateway-kubernetes/pkg/sdk"
	apiv1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/runtime"
	ctlr "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	gatewayv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// clusterTimeout is a timeout for connections to the Kubernetes API
const clusterTimeout = 10 * time.Second

var scheme = runtime.NewScheme()

func init() {
	// TO-DO: handle errors returned by the calls bellow
	_ = gatewayv1alpha2.AddToScheme(scheme)
	_ = nginxgwv1alpha1.AddToScheme(scheme)
	_ = apiv1.AddToScheme(scheme)
}

func Start(cfg config.Config) error {
	logger := cfg.Logger

	options := manager.Options{
		Scheme: scheme,
	}

	eventCh := make(chan interface{})

	clusterCfg := ctlr.GetConfigOrDie()
	clusterCfg.Timeout = clusterTimeout

	mgr, err := manager.New(clusterCfg, options)
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
	err = sdk.RegisterServiceController(mgr, svc.NewServiceImplementation(cfg, eventCh))
	if err != nil {
		return fmt.Errorf("cannot register service implementation: %w", err)
	}

	conf := state.NewConfiguration(cfg.GatewayCtlrName, state.NewRealClock())
	serviceStore := state.NewServiceStore()
	reporter := status.NewUpdater(mgr.GetClient(), cfg.Logger)
	eventLoop := events.NewEventLoop(conf, serviceStore, eventCh, reporter, cfg.Logger)

	err = mgr.Add(eventLoop)
	if err != nil {
		return fmt.Errorf("cannot register event loop: %w", err)
	}

	ctx := ctlr.SetupSignalHandler()

	logger.Info("Starting manager")
	return mgr.Start(ctx)
}
