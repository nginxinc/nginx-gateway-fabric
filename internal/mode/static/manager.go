package static

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	ngxclient "github.com/nginxinc/nginx-plus-go-client/client"
	"github.com/prometheus/client_golang/prometheus"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	discoveryV1 "k8s.io/api/discovery/v1"
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/record"
	ctlr "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlcfg "sigs.k8s.io/controller-runtime/pkg/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	k8spredicate "sigs.k8s.io/controller-runtime/pkg/predicate"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/controller"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/controller/filter"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/controller/index"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/controller/predicate"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/events"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/gatewayclass"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/runnables"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/status"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/config"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/metrics/collectors"
	ngxcfg "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/config"
	ngxvalidation "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/config/validation"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/file"
	ngxruntime "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/runtime"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/resolver"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/validation"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/telemetry"
)

const (
	// clusterTimeout is a timeout for connections to the Kubernetes API
	clusterTimeout = 10 * time.Second
)

var scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(gatewayv1beta1.AddToScheme(scheme))
	utilruntime.Must(gatewayv1.AddToScheme(scheme))
	utilruntime.Must(gatewayv1alpha2.AddToScheme(scheme))
	utilruntime.Must(apiv1.AddToScheme(scheme))
	utilruntime.Must(discoveryV1.AddToScheme(scheme))
	utilruntime.Must(ngfAPI.AddToScheme(scheme))
	utilruntime.Must(apiext.AddToScheme(scheme))
	utilruntime.Must(appsv1.AddToScheme(scheme))
}

// nolint:gocyclo
func StartManager(cfg config.Config) error {
	nginxChecker := newNginxConfiguredOnStartChecker()
	mgr, err := createManager(cfg, nginxChecker)
	if err != nil {
		return fmt.Errorf("cannot build runtime manager: %w", err)
	}

	recorderName := fmt.Sprintf("nginx-gateway-fabric-%s", cfg.GatewayClassName)
	recorder := mgr.GetEventRecorderFor(recorderName)

	promLogger, err := newLeveledPrometheusLogger()
	if err != nil {
		return fmt.Errorf("error creating leveled prometheus logger: %w", err)
	}

	logLevelSetter := newMultiLogLevelSetter(newZapLogLevelSetter(cfg.AtomicLevel), newPromLogLevelSetter(promLogger))

	ctx := ctlr.SetupSignalHandler()

	eventCh := make(chan interface{})
	controlConfigNSName := types.NamespacedName{
		Namespace: cfg.GatewayPodConfig.Namespace,
		Name:      cfg.ConfigName,
	}
	if err := registerControllers(ctx, cfg, mgr, recorder, logLevelSetter, eventCh, controlConfigNSName); err != nil {
		return err
	}

	// protectedPorts is the map of ports that may not be configured by a listener, and the name of what it is used for
	protectedPorts := map[int32]string{
		int32(cfg.MetricsConfig.Port): "MetricsPort",
		int32(cfg.HealthConfig.Port):  "HealthPort",
	}

	processor := state.NewChangeProcessorImpl(state.ChangeProcessorConfig{
		GatewayCtlrName:  cfg.GatewayCtlrName,
		GatewayClassName: cfg.GatewayClassName,
		Logger:           cfg.Logger.WithName("changeProcessor"),
		Validators: validation.Validators{
			HTTPFieldsValidator: ngxvalidation.HTTPValidator{},
		},
		EventRecorder:  recorder,
		Scheme:         scheme,
		ProtectedPorts: protectedPorts,
	})

	// Clear the configuration folders to ensure that no files are left over in case the control plane was restarted
	// (this assumes the folders are in a shared volume).
	removedPaths, err := file.ClearFolders(file.NewStdLibOSFileManager(), ngxcfg.ConfigFolders)
	for _, path := range removedPaths {
		cfg.Logger.Info("removed configuration file", "path", path)
	}
	if err != nil {
		return fmt.Errorf("cannot clear NGINX configuration folders: %w", err)
	}

	// Ensure NGINX is running before registering metrics & starting the manager.
	if err := ngxruntime.EnsureNginxRunning(ctx); err != nil {
		return fmt.Errorf("NGINX is not running: %w", err)
	}

	var (
		ngxruntimeCollector ngxruntime.MetricsCollector = collectors.NewManagerNoopCollector()
		handlerCollector    handlerMetricsCollector     = collectors.NewControllerNoopCollector()
	)

	var ngxPlusClient *ngxclient.NginxClient
	if cfg.Plus {
		ngxPlusClient, err = ngxruntime.CreatePlusClient()
		if err != nil {
			return fmt.Errorf("error creating NGINX plus client: %w", err)
		}
	}

	if cfg.MetricsConfig.Enabled {
		constLabels := map[string]string{"class": cfg.GatewayClassName}
		var ngxCollector prometheus.Collector
		if cfg.Plus {
			ngxCollector, err = collectors.NewNginxPlusMetricsCollector(ngxPlusClient, constLabels, promLogger)
		} else {
			ngxCollector = collectors.NewNginxMetricsCollector(constLabels, promLogger)
		}
		if err != nil {
			return fmt.Errorf("cannot create nginx metrics collector: %w", err)
		}

		ngxruntimeCollector = collectors.NewManagerMetricsCollector(constLabels)
		handlerCollector = collectors.NewControllerCollector(constLabels)
		metrics.Registry.MustRegister(
			ngxCollector,
			ngxruntimeCollector.(prometheus.Collector),
			handlerCollector.(prometheus.Collector),
		)
	}

	statusUpdater := status.NewUpdater(status.UpdaterConfig{
		GatewayCtlrName:          cfg.GatewayCtlrName,
		GatewayClassName:         cfg.GatewayClassName,
		Client:                   mgr.GetClient(),
		Logger:                   cfg.Logger.WithName("statusUpdater"),
		Clock:                    status.NewRealClock(),
		UpdateGatewayClassStatus: cfg.UpdateGatewayClassStatus,
		LeaderElectionEnabled:    cfg.LeaderElection.Enabled,
	})

	eventHandler := newEventHandlerImpl(eventHandlerConfig{
		k8sClient:       mgr.GetClient(),
		processor:       processor,
		serviceResolver: resolver.NewServiceResolverImpl(mgr.GetClient()),
		generator:       ngxcfg.NewGeneratorImpl(cfg.Plus),
		logLevelSetter:  logLevelSetter,
		nginxFileMgr: file.NewManagerImpl(
			cfg.Logger.WithName("nginxFileManager"),
			file.NewStdLibOSFileManager(),
		),
		nginxRuntimeMgr: ngxruntime.NewManagerImpl(
			ngxPlusClient,
			ngxruntimeCollector,
			cfg.Logger.WithName("nginxRuntimeManager"),
		),
		statusUpdater:                 statusUpdater,
		eventRecorder:                 recorder,
		nginxConfiguredOnStartChecker: nginxChecker,
		controlConfigNSName:           controlConfigNSName,
		gatewayPodConfig:              cfg.GatewayPodConfig,
		metricsCollector:              handlerCollector,
	})

	objects, objectLists := prepareFirstEventBatchPreparerArgs(
		cfg.GatewayClassName,
		cfg.GatewayNsName,
		cfg.EnableExperimentalFeatures,
	)
	firstBatchPreparer := events.NewFirstEventBatchPreparerImpl(mgr.GetCache(), objects, objectLists)
	eventLoop := events.NewEventLoop(
		eventCh,
		cfg.Logger.WithName("eventLoop"),
		eventHandler,
		firstBatchPreparer,
	)

	if err = mgr.Add(&runnables.LeaderOrNonLeader{Runnable: eventLoop}); err != nil {
		return fmt.Errorf("cannot register event loop: %w", err)
	}

	if err = mgr.Add(runnables.NewEnableAfterBecameLeader(statusUpdater.Enable)); err != nil {
		return fmt.Errorf("cannot register status updater: %w", err)
	}

	dataCollector := telemetry.NewDataCollectorImpl(telemetry.DataCollectorConfig{
		K8sClientReader:     mgr.GetAPIReader(),
		GraphGetter:         processor,
		ConfigurationGetter: eventHandler,
		Version:             cfg.Version,
		PodNSName: types.NamespacedName{
			Namespace: cfg.GatewayPodConfig.Namespace,
			Name:      cfg.GatewayPodConfig.Name,
		},
	})
	if err = mgr.Add(createTelemetryJob(cfg, dataCollector, nginxChecker.getReadyCh())); err != nil {
		return fmt.Errorf("cannot register telemetry job: %w", err)
	}

	cfg.Logger.Info("Starting manager")
	return mgr.Start(ctx)
}

func createManager(cfg config.Config, nginxChecker *nginxConfiguredOnStartChecker) (manager.Manager, error) {
	options := manager.Options{
		Scheme:  scheme,
		Logger:  cfg.Logger,
		Metrics: getMetricsOptions(cfg.MetricsConfig),
		// Note: when the leadership is lost, the manager will return an error in the Start() method.
		// However, it will not wait for any Runnable it starts to finish, meaning any in-progress operations
		// might get terminated half-way.
		LeaderElection:          true,
		LeaderElectionNamespace: cfg.GatewayPodConfig.Namespace,
		LeaderElectionID:        cfg.LeaderElection.LockName,
		// We're not enabling LeaderElectionReleaseOnCancel because when the Manager stops gracefully, it waits
		// for all started Runnables (including Leader-only ones) to finish. Otherwise, the new leader might start
		// running Leader-only Runnables before the old leader has finished running them.
		// See the doc comment for the LeaderElectionReleaseOnCancel for more details.
		LeaderElectionReleaseOnCancel: false,
		Controller: ctrlcfg.Controller{
			// All of our controllers still need to work in case of non-leader pods
			NeedLeaderElection: helpers.GetPointer(false),
		},
	}

	if cfg.HealthConfig.Enabled {
		options.HealthProbeBindAddress = fmt.Sprintf(":%d", cfg.HealthConfig.Port)
	}

	clusterCfg := ctlr.GetConfigOrDie()
	clusterCfg.Timeout = clusterTimeout

	mgr, err := manager.New(clusterCfg, options)
	if err != nil {
		return nil, err
	}

	if cfg.HealthConfig.Enabled {
		if err := mgr.AddReadyzCheck("readyz", nginxChecker.readyCheck); err != nil {
			return nil, fmt.Errorf("error adding ready check: %w", err)
		}
	}

	return mgr, nil
}

func registerControllers(
	ctx context.Context,
	cfg config.Config,
	mgr manager.Manager,
	recorder record.EventRecorder,
	logLevelSetter logLevelSetter,
	eventCh chan interface{},
	controlConfigNSName types.NamespacedName,
) error {
	type ctlrCfg struct {
		objectType client.Object
		options    []controller.Option
	}

	crdWithGVK := apiext.CustomResourceDefinition{}
	crdWithGVK.SetGroupVersionKind(
		schema.GroupVersionKind{Group: apiext.GroupName, Version: "v1", Kind: "CustomResourceDefinition"},
	)

	// Note: for any new object type or a change to the existing one,
	// make sure to also update prepareFirstEventBatchPreparerArgs()
	controllerRegCfgs := []ctlrCfg{
		{
			objectType: &gatewayv1.GatewayClass{},
			options: []controller.Option{
				controller.WithK8sPredicate(k8spredicate.And(
					k8spredicate.GenerationChangedPredicate{},
					predicate.GatewayClassPredicate{ControllerName: cfg.GatewayCtlrName})),
			},
		},
		{
			objectType: &gatewayv1.Gateway{},
			options: func() []controller.Option {
				options := []controller.Option{
					controller.WithK8sPredicate(k8spredicate.GenerationChangedPredicate{}),
				}
				if cfg.GatewayNsName != nil {
					options = append(
						options,
						controller.WithNamespacedNameFilter(filter.CreateSingleResourceFilter(*cfg.GatewayNsName)),
					)
				}
				return options
			}(),
		},
		{
			objectType: &gatewayv1.HTTPRoute{},
			options: []controller.Option{
				controller.WithK8sPredicate(k8spredicate.GenerationChangedPredicate{}),
			},
		},
		{
			objectType: &apiv1.Service{},
			options: []controller.Option{
				controller.WithK8sPredicate(predicate.ServicePortsChangedPredicate{}),
			},
		},
		{
			objectType: &apiv1.Service{},
			options: func() []controller.Option {
				svcNSName := types.NamespacedName{
					Namespace: cfg.GatewayPodConfig.Namespace,
					Name:      cfg.GatewayPodConfig.ServiceName,
				}
				return []controller.Option{
					controller.WithK8sPredicate(predicate.GatewayServicePredicate{NSName: svcNSName}),
				}
			}(),
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
			options: []controller.Option{
				controller.WithK8sPredicate(k8spredicate.GenerationChangedPredicate{}),
			},
		},
		{
			objectType: &crdWithGVK,
			options: []controller.Option{
				controller.WithOnlyMetadata(),
				controller.WithK8sPredicate(
					predicate.AnnotationPredicate{Annotation: gatewayclass.BundleVersionAnnotation},
				),
			},
		},
	}

	if cfg.EnableExperimentalFeatures {
		backendTLSObjs := []ctlrCfg{
			{
				objectType: &gatewayv1alpha2.BackendTLSPolicy{},
			},
			{
				objectType: &apiv1.ConfigMap{},
			},
		}
		controllerRegCfgs = append(controllerRegCfgs, backendTLSObjs...)
	}

	if cfg.ConfigName != "" {
		controllerRegCfgs = append(controllerRegCfgs,
			ctlrCfg{
				objectType: &ngfAPI.NginxGateway{},
				options: []controller.Option{
					controller.WithNamespacedNameFilter(filter.CreateSingleResourceFilter(controlConfigNSName)),
				},
			})
		if err := setInitialConfig(
			mgr.GetAPIReader(),
			cfg.Logger,
			recorder,
			logLevelSetter,
			controlConfigNSName,
		); err != nil {
			return fmt.Errorf("error setting initial control plane configuration: %w", err)
		}
	}

	for _, regCfg := range controllerRegCfgs {
		if err := controller.Register(
			ctx,
			regCfg.objectType,
			mgr,
			eventCh,
			regCfg.options...,
		); err != nil {
			return fmt.Errorf("cannot register controller for %T: %w", regCfg.objectType, err)
		}
	}
	return nil
}

func createTelemetryJob(
	cfg config.Config,
	dataCollector telemetry.DataCollector,
	readyCh <-chan struct{},
) *runnables.Leader {
	logger := cfg.Logger.WithName("telemetryJob")
	exporter := telemetry.NewLoggingExporter(cfg.Logger.WithName("telemetryExporter").V(1 /* debug */))

	worker := telemetry.CreateTelemetryJobWorker(logger, exporter, dataCollector)

	// 10 min jitter is enough per telemetry destination recommendation
	// For the default period of 24 hours, jitter will be 10min /(24*60)min  = 0.0069
	jitterFactor := 10.0 / (24 * 60) // added jitter is bound by jitterFactor * period

	return &runnables.Leader{
		Runnable: runnables.NewCronJob(
			runnables.CronJobConfig{
				Worker:       worker,
				Logger:       logger,
				Period:       cfg.TelemetryReportPeriod,
				JitterFactor: jitterFactor,
				ReadyCh:      readyCh,
			},
		),
	}
}

func prepareFirstEventBatchPreparerArgs(
	gcName string,
	gwNsName *types.NamespacedName,
	enableExperimentalFeatures bool,
) ([]client.Object, []client.ObjectList) {
	objects := []client.Object{
		&gatewayv1.GatewayClass{ObjectMeta: metav1.ObjectMeta{Name: gcName}},
	}

	partialObjectMetadataList := &metav1.PartialObjectMetadataList{}
	partialObjectMetadataList.SetGroupVersionKind(
		schema.GroupVersionKind{
			Group:   apiext.GroupName,
			Version: "v1",
			Kind:    "CustomResourceDefinition",
		},
	)

	objectLists := []client.ObjectList{
		&apiv1.ServiceList{},
		&apiv1.SecretList{},
		&apiv1.NamespaceList{},
		&discoveryV1.EndpointSliceList{},
		&gatewayv1.HTTPRouteList{},
		&gatewayv1beta1.ReferenceGrantList{},
		partialObjectMetadataList,
	}

	if enableExperimentalFeatures {
		objectLists = append(objectLists, &gatewayv1alpha2.BackendTLSPolicyList{}, &apiv1.ConfigMapList{})
	}

	if gwNsName == nil {
		objectLists = append(objectLists, &gatewayv1.GatewayList{})
	} else {
		objects = append(
			objects,
			&gatewayv1.Gateway{ObjectMeta: metav1.ObjectMeta{Name: gwNsName.Name, Namespace: gwNsName.Namespace}},
		)
	}

	return objects, objectLists
}

func setInitialConfig(
	reader client.Reader,
	logger logr.Logger,
	eventRecorder record.EventRecorder,
	logLevelSetter logLevelSetter,
	configName types.NamespacedName,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var config ngfAPI.NginxGateway
	// Polling to wait for CRD to exist if the Deployment is created first.
	if err := wait.PollUntilContextCancel(
		ctx,
		500*time.Millisecond,
		true, /* poll immediately */
		func(ctx context.Context) (bool, error) {
			if err := reader.Get(ctx, configName, &config); err != nil {
				if !apierrors.IsNotFound(err) {
					return false, err
				}
				return false, nil
			}
			return true, nil
		},
	); err != nil {
		return fmt.Errorf("NginxGateway %s not found: %w", configName, err)
	}

	// status is not updated until the status updater's cache is started and the
	// resource is processed by the controller
	return updateControlPlane(&config, logger, eventRecorder, configName, logLevelSetter)
}

func getMetricsOptions(cfg config.MetricsConfig) metricsserver.Options {
	metricsOptions := metricsserver.Options{BindAddress: "0"}

	if cfg.Enabled {
		if cfg.Secure {
			metricsOptions.SecureServing = true
		}
		metricsOptions.BindAddress = fmt.Sprintf(":%v", cfg.Port)
	}

	return metricsOptions
}
