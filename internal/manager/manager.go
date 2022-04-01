package manager

import (
	"fmt"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctlr "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	gatewayv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/config"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/events"
	gw "github.com/nginxinc/nginx-kubernetes-gateway/internal/implementations/gateway"
	gc "github.com/nginxinc/nginx-kubernetes-gateway/internal/implementations/gatewayclass"
	gcfg "github.com/nginxinc/nginx-kubernetes-gateway/internal/implementations/gatewayconfig"
	hr "github.com/nginxinc/nginx-kubernetes-gateway/internal/implementations/httproute"
	svc "github.com/nginxinc/nginx-kubernetes-gateway/internal/implementations/service"
	ngxcfg "github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/config"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/file"
	ngxruntime "github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/runtime"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/status"
	nginxgwv1alpha1 "github.com/nginxinc/nginx-kubernetes-gateway/pkg/apis/gateway/v1alpha1"
	"github.com/nginxinc/nginx-kubernetes-gateway/pkg/sdk"
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
	configGenerator := ngxcfg.NewGeneratorImpl(serviceStore)
	nginxFileMgr := file.NewManagerImpl()
	nginxRuntimeMgr := ngxruntime.NewManagerImpl()
	eventLoop := events.NewEventLoop(conf, serviceStore, configGenerator, eventCh, reporter, cfg.Logger, nginxFileMgr, nginxRuntimeMgr)

	err = mgr.Add(eventLoop)
	if err != nil {
		return fmt.Errorf("cannot register event loop: %w", err)
	}

	ctx := ctlr.SetupSignalHandler()

	logger.Info("Starting manager")
	return mgr.Start(ctx)
}
