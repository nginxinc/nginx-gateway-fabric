package static

import (
	"fmt"
	"time"

	apiv1 "k8s.io/api/core/v1"
	discoveryV1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctlr "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	k8spredicate "sigs.k8s.io/controller-runtime/pkg/predicate"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/framework/config"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/framework/controller"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/framework/events"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/framework/manager/filter"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/framework/manager/index"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/framework/manager/predicate"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/framework/status"
	ngxcfg "github.com/nginxinc/nginx-kubernetes-gateway/internal/mode/static/nginx/config"
	ngxvalidation "github.com/nginxinc/nginx-kubernetes-gateway/internal/mode/static/nginx/config/validation"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/mode/static/nginx/file"
	ngxruntime "github.com/nginxinc/nginx-kubernetes-gateway/internal/mode/static/nginx/runtime"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/mode/static/state"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/mode/static/state/relationship"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/mode/static/state/resolver"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/mode/static/state/validation"
)

const (
	// clusterTimeout is a timeout for connections to the Kubernetes API
	clusterTimeout = 10 * time.Second
)

var scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(gatewayv1beta1.AddToScheme(scheme))
	utilruntime.Must(apiv1.AddToScheme(scheme))
	utilruntime.Must(discoveryV1.AddToScheme(scheme))
}

func StartManager(cfg config.Config) error {
	logger := cfg.Logger

	options := manager.Options{
		Scheme: scheme,
		Logger: logger,
		// We disable the metrics server because we reserve all ports (1-65535) for the data plane.
		// Once we add support for Prometheus, we can make this port configurable by the user.
		MetricsBindAddress: "0",
	}

	eventCh := make(chan interface{})

	clusterCfg := ctlr.GetConfigOrDie()
	clusterCfg.Timeout = clusterTimeout

	mgr, err := manager.New(clusterCfg, options)
	if err != nil {
		return fmt.Errorf("cannot build runtime manager: %w", err)
	}

	// Note: for any new object type or a change to the existing one,
	// make sure to also update prepareFirstEventBatchPreparerArgs()
	controllerRegCfgs := []struct {
		objectType client.Object
		options    []controller.Option
	}{
		{
			objectType: &gatewayv1beta1.GatewayClass{},
			options: []controller.Option{
				controller.WithK8sPredicate(predicate.GatewayClassPredicate{ControllerName: cfg.GatewayCtlrName}),
			},
		},
		{
			objectType: &gatewayv1beta1.Gateway{},
			options: func() []controller.Option {
				if cfg.GatewayNsName != nil {
					return []controller.Option{
						controller.WithNamespacedNameFilter(filter.CreateSingleResourceFilter(*cfg.GatewayNsName)),
					}
				}
				return nil
			}(),
		},
		{
			objectType: &gatewayv1beta1.HTTPRoute{},
		},
		{
			objectType: &apiv1.Service{},
			options: []controller.Option{
				controller.WithK8sPredicate(predicate.ServicePortsChangedPredicate{}),
			},
		},
		{
			objectType: &apiv1.Secret{},
		},
		{
			objectType: &discoveryV1.EndpointSlice{},
			options: []controller.Option{
				controller.WithK8sPredicate(k8spredicate.GenerationChangedPredicate{}),
				controller.WithFieldIndices(index.CreateEndpointSliceFieldIndices()),
			},
		},
		{
			objectType: &apiv1.Namespace{},
			options: []controller.Option{
				controller.WithK8sPredicate(k8spredicate.LabelChangedPredicate{}),
			},
		},
		{
			objectType: &gatewayv1beta1.ReferenceGrant{},
		},
	}

	ctx := ctlr.SetupSignalHandler()

	for _, regCfg := range controllerRegCfgs {
		err := controller.Register(ctx, regCfg.objectType, mgr, eventCh, regCfg.options...)
		if err != nil {
			return fmt.Errorf("cannot register controller for %T: %w", regCfg.objectType, err)
		}
	}

	recorderName := fmt.Sprintf("nginx-kubernetes-gateway-%s", cfg.GatewayClassName)
	recorder := mgr.GetEventRecorderFor(recorderName)

	processor := state.NewChangeProcessorImpl(state.ChangeProcessorConfig{
		GatewayCtlrName:      cfg.GatewayCtlrName,
		GatewayClassName:     cfg.GatewayClassName,
		RelationshipCapturer: relationship.NewCapturerImpl(),
		Logger:               cfg.Logger.WithName("changeProcessor"),
		Validators: validation.Validators{
			HTTPFieldsValidator: ngxvalidation.HTTPValidator{},
		},
		EventRecorder: recorder,
		Scheme:        scheme,
	})

	configGenerator := ngxcfg.NewGeneratorImpl()

	// Clear the configuration folders to ensure that no files are left over in case the control plane was restarted
	// (this assumes the folders are in a shared volume).
	removedPaths, err := file.ClearFolders(file.NewStdLibOSFileManager(), ngxcfg.ConfigFolders)
	for _, path := range removedPaths {
		logger.Info("removed configuration file", "path", path)
	}
	if err != nil {
		return fmt.Errorf("cannot clear NGINX configuration folders: %w", err)
	}

	nginxFileMgr := file.NewManagerImpl(logger.WithName("nginxFileManager"), file.NewStdLibOSFileManager())
	nginxRuntimeMgr := ngxruntime.NewManagerImpl()
	statusUpdater := status.NewUpdater(status.UpdaterConfig{
		GatewayCtlrName:          cfg.GatewayCtlrName,
		GatewayClassName:         cfg.GatewayClassName,
		Client:                   mgr.GetClient(),
		PodIP:                    cfg.PodIP,
		Logger:                   cfg.Logger.WithName("statusUpdater"),
		Clock:                    status.NewRealClock(),
		UpdateGatewayClassStatus: cfg.UpdateGatewayClassStatus,
	})

	eventHandler := NewEventHandlerImpl(EventHandlerConfig{
		Processor:       processor,
		ServiceResolver: resolver.NewServiceResolverImpl(mgr.GetClient()),
		Generator:       configGenerator,
		Logger:          cfg.Logger.WithName("eventHandler"),
		NginxFileMgr:    nginxFileMgr,
		NginxRuntimeMgr: nginxRuntimeMgr,
		StatusUpdater:   statusUpdater,
	})

	objects, objectLists := prepareFirstEventBatchPreparerArgs(cfg.GatewayClassName, cfg.GatewayNsName)
	firstBatchPreparer := events.NewFirstEventBatchPreparerImpl(mgr.GetCache(), objects, objectLists)

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

func prepareFirstEventBatchPreparerArgs(
	gcName string,
	gwNsName *types.NamespacedName,
) ([]client.Object, []client.ObjectList) {
	objects := []client.Object{
		&gatewayv1beta1.GatewayClass{ObjectMeta: metav1.ObjectMeta{Name: gcName}},
	}
	objectLists := []client.ObjectList{
		&apiv1.ServiceList{},
		&apiv1.SecretList{},
		&apiv1.NamespaceList{},
		&discoveryV1.EndpointSliceList{},
		&gatewayv1beta1.HTTPRouteList{},
		&gatewayv1beta1.ReferenceGrantList{},
	}

	if gwNsName == nil {
		objectLists = append(objectLists, &gatewayv1beta1.GatewayList{})
	} else {
		objects = append(
			objects,
			&gatewayv1beta1.Gateway{ObjectMeta: metav1.ObjectMeta{Name: gwNsName.Name, Namespace: gwNsName.Namespace}},
		)
	}

	return objects, objectLists
}
