package manager

import (
	"fmt"
	"time"

	apiv1 "k8s.io/api/core/v1"
	discoveryV1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctlr "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	k8spredicate "sigs.k8s.io/controller-runtime/pkg/predicate"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/config"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/events"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/manager/filter"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/manager/index"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/manager/predicate"
	ngxcfg "github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/config"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/file"
	ngxruntime "github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/runtime"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/relationship"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/resolver"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/status"
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
	utilruntime.Must(gatewayv1beta1.AddToScheme(scheme))
	utilruntime.Must(apiv1.AddToScheme(scheme))
	utilruntime.Must(discoveryV1.AddToScheme(scheme))
}

func Start(cfg config.Config) error {
	logger := cfg.Logger

	options := manager.Options{
		Scheme: scheme,
		Logger: logger,
	}

	eventCh := make(chan interface{})

	clusterCfg := ctlr.GetConfigOrDie()
	clusterCfg.Timeout = clusterTimeout

	mgr, err := manager.New(clusterCfg, options)
	if err != nil {
		return fmt.Errorf("cannot build runtime manager: %w", err)
	}

	controllerConfigs := []controllerConfig{
		{
			objectType:           &gatewayv1beta1.GatewayClass{},
			namespacedNameFilter: filter.CreateFilterForGatewayClass(cfg.GatewayClassName),
		},
		{
			objectType: &gatewayv1beta1.Gateway{},
		},
		{
			objectType: &gatewayv1beta1.HTTPRoute{},
		},
		{
			objectType:     &apiv1.Service{},
			k8sEventFilter: predicate.ServicePortsChangedPredicate{},
		},
		{
			objectType: &apiv1.Secret{},
		},
		{
			objectType:     &discoveryV1.EndpointSlice{},
			k8sEventFilter: k8spredicate.GenerationChangedPredicate{},
			fieldIndexes: map[string]client.IndexerFunc{
				index.KubernetesServiceNameIndexField: index.ServiceNameIndexFunc,
			},
		},
	}

	ctx := ctlr.SetupSignalHandler()

	for _, cfg := range controllerConfigs {
		err := registerController(ctx, mgr, eventCh, cfg)
		if err != nil {
			return fmt.Errorf("cannot register controller for %T: %w", cfg.objectType, err)
		}
	}

	secretStore := state.NewSecretStore()
	secretMemoryMgr := state.NewSecretDiskMemoryManager(secretsFolder, secretStore)

	processor := state.NewChangeProcessorImpl(state.ChangeProcessorConfig{
		GatewayCtlrName:      cfg.GatewayCtlrName,
		GatewayClassName:     cfg.GatewayClassName,
		SecretMemoryManager:  secretMemoryMgr,
		ServiceResolver:      resolver.NewServiceResolverImpl(mgr.GetClient()),
		RelationshipCapturer: relationship.NewCapturerImpl(),
		Logger:               cfg.Logger.WithName("changeProcessor"),
	})

	configGenerator := ngxcfg.NewGeneratorImpl()
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
			&discoveryV1.EndpointSliceList{},
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

	logger.Info("Starting manager")
	return mgr.Start(ctx)
}
