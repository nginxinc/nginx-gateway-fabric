package main

import (
	"fmt"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctlr "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/events"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/manager/filter"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/status"
)

var scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(gatewayv1beta1.AddToScheme(scheme))
	utilruntime.Must(apiv1.AddToScheme(scheme))
	utilruntime.Must(v1.AddToScheme(scheme))
}

func startManager(logger logr.Logger, gcName string) error {
	clusterCfg := ctlr.GetConfigOrDie()

	options := manager.Options{
		Scheme: scheme,
		Logger: logger,
	}

	mgr, err := manager.New(clusterCfg, options)
	if err != nil {
		return fmt.Errorf("cannot build runtime manager: %w", err)
	}

	controllerRegCfgs := []struct {
		objectType client.Object
		options    []controllerOption
	}{
		{
			objectType: &gatewayv1beta1.GatewayClass{},
			options: []controllerOption{
				withNamespacedNameFilter(filter.CreateFilterForGatewayClass(gcName)),
			},
		},
		{
			objectType: &gatewayv1beta1.Gateway{},
		},
	}

	ctx := ctlr.SetupSignalHandler()

	eventCh := make(chan interface{})

	for _, regCfg := range controllerRegCfgs {
		err := registerController(ctx, regCfg.objectType, mgr, eventCh, regCfg.options...)
		if err != nil {
			return fmt.Errorf("cannot register controller for %T: %w", regCfg.objectType, err)
		}
	}

	firstBatchPreparer := events.NewFirstEventBatchPreparerImpl(
		mgr.GetCache(),
		[]client.Object{
			&gatewayv1beta1.GatewayClass{ObjectMeta: metav1.ObjectMeta{Name: gcName}},
		},
		[]client.ObjectList{
			&gatewayv1beta1.GatewayList{},
		},
	)

	statusUpdater := status.NewUpdater(
		status.UpdaterConfig{
			Client:           mgr.GetClient(),
			Clock:            status.NewRealClock(),
			Logger:           logger.WithName("statusUpdater"),
			GatewayClassName: gcName,
		},
	)

	handler := newEventHandler(
		gcName,
		statusUpdater,
		mgr.GetClient(),
		logger.WithName("eventHandler"),
	)

	eventLoop := events.NewEventLoop(
		eventCh,
		logger.WithName("eventLoop"),
		handler,
		firstBatchPreparer)

	err = mgr.Add(eventLoop)
	if err != nil {
		return fmt.Errorf("cannot register event loop: %w", err)
	}

	logger.Info("Starting manager")
	return mgr.Start(ctx)
}
