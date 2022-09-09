package manager

import (
	"fmt"
	"time"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctlr "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/config"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/events"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/implementations"
	ngxcfg "github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/config"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/file"
	ngxruntime "github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/runtime"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/status"
	"github.com/nginxinc/nginx-kubernetes-gateway/pkg/sdk"
)

const (
	// clusterTimeout is a timeout for connections to the Kubernetes API
	clusterTimeout = 10 * time.Second
	// secretsFolder is the folder that holds all the secrets for NGINX servers.
	// nolint:gosec
	secretsFolder = "/etc/nginx/secrets"
)

var scheme = runtime.NewScheme()

func init() {
	// FIXME(pleshakov): handle errors returned by the calls bellow
	_ = gatewayv1beta1.AddToScheme(scheme)
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

	// Register GatewayClass implementation
	err = sdk.RegisterController[*gatewayv1beta1.GatewayClass](mgr,
		implementations.NewImplementationWithFilter[*gatewayv1beta1.GatewayClass](cfg.Logger, eventCh, func(nsname types.NamespacedName) (bool, string) {
			if nsname.Name == cfg.GatewayClassName {
				return false, fmt.Sprintf("GatewayClass was upserted but ignored because this controller only supports the GatewayClass %s", cfg.GatewayClassName)
			}
			return true, ""
		}),
	)
	if err != nil {
		return fmt.Errorf("cannot register GatewayClass implementation: %w", err)
	}

	// Register Gateway implementation
	err = sdk.RegisterController[*gatewayv1beta1.Gateway](mgr,
		implementations.NewImplementation[*gatewayv1beta1.Gateway](cfg.Logger, eventCh),
	)
	if err != nil {
		return fmt.Errorf("cannot register Gateway implementation: %w", err)
	}

	// Register HTTPRoute implementation
	err = sdk.RegisterController[*gatewayv1beta1.HTTPRoute](mgr,
		implementations.NewImplementation[*gatewayv1beta1.HTTPRoute](cfg.Logger, eventCh),
	)
	if err != nil {
		return fmt.Errorf("cannot register HTTPRoute implementation: %w", err)
	}

	// Register Service implementation
	err = sdk.RegisterController[*apiv1.Service](mgr,
		implementations.NewImplementation[*apiv1.Service](cfg.Logger, eventCh),
	)
	if err != nil {
		return fmt.Errorf("cannot register Service implementation: %w", err)
	}

	// Register Secret implementation
	err = sdk.RegisterController[*apiv1.Secret](mgr,
		implementations.NewImplementation[*apiv1.Secret](cfg.Logger, eventCh),
	)
	if err != nil {
		return fmt.Errorf("cannot register Secret implementation: %w", err)
	}

	secretStore := state.NewSecretStore()
	secretMemoryMgr := state.NewSecretDiskMemoryManager(secretsFolder, secretStore)

	processor := state.NewChangeProcessorImpl(state.ChangeProcessorConfig{
		GatewayCtlrName:     cfg.GatewayCtlrName,
		GatewayClassName:    cfg.GatewayClassName,
		SecretMemoryManager: secretMemoryMgr,
	})

	serviceStore := state.NewServiceStore()
	configGenerator := ngxcfg.NewGeneratorImpl(serviceStore)
	nginxFileMgr := file.NewManagerImpl()
	nginxRuntimeMgr := ngxruntime.NewManagerImpl()
	statusUpdater := status.NewUpdater(status.UpdaterConfig{
		GatewayCtlrName:  cfg.GatewayCtlrName,
		GatewayClassName: cfg.GatewayClassName,
		Client:           mgr.GetClient(),
		// FIXME(pleshakov) Make sure each component:
		// (1) Has a dedicated named logger.
		// (2) Get it from the Manager (the WithName is done here for all components).
		Logger: cfg.Logger.WithName("statusUpdater"),
		Clock:  status.NewRealClock(),
	})

	eventHandler := events.NewEventHandlerImpl(events.EventHandlerConfig{
		Processor:           processor,
		ServiceStore:        serviceStore,
		SecretStore:         secretStore,
		SecretMemoryManager: secretMemoryMgr,
		Generator:           configGenerator,
		Logger:              cfg.Logger.WithName("eventHandler"),
		NginxFileMgr:        nginxFileMgr,
		NginxRuntimeMgr:     nginxRuntimeMgr,
		StatusUpdater:       statusUpdater,
	})

	firstBatchPreparer := events.NewFirstEventBatchPreparerImpl(
		mgr.GetCache(),
		[]client.Object{
			&gatewayv1beta1.GatewayClass{ObjectMeta: metav1.ObjectMeta{Name: cfg.GatewayClassName}},
		},
		[]client.ObjectList{
			&apiv1.ServiceList{},
			&apiv1.SecretList{},
			&gatewayv1beta1.GatewayList{},
			&gatewayv1beta1.HTTPRouteList{},
		},
	)

	eventLoop := events.NewEventLoop(
		eventCh,
		cfg.Logger.WithName("eventLoop"),
		eventHandler,
		firstBatchPreparer)

	err = mgr.Add(eventLoop)
	if err != nil {
		return fmt.Errorf("cannot register event loop: %w", err)
	}

	ctx := ctlr.SetupSignalHandler()

	logger.Info("Starting manager")
	return mgr.Start(ctx)
}
