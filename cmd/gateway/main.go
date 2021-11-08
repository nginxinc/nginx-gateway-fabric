package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/nginxinc/nginx-gateway-kubernetes/internal/implementation"
	"github.com/nginxinc/nginx-gateway-kubernetes/internal/sdk"
	"go.uber.org/zap"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/log"
	rzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
)

var (
	// Set during go build
	version string
	commit  string
	date    string

	// Command-line flags
	gatewayClass = flag.String("gatewayclass", "", "Tha name of the GatewayClass resource")
)

func main() {
	flag.Parse()

	if *gatewayClass == "" {
		fmt.Fprintln(os.Stderr, "-gatewayclass argument must be set")
		os.Exit(1)
	}

	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logging: %v\n", err)
		os.Exit(1)
	}
	sugar := logger.Sugar()

	sugar.Infow("Starting NGINX Gateway",
		"version", version,
		"commit", commit,
		"date", date)

	config, err := rest.InClusterConfig()
	if err != nil {
		sugar.Fatalw("Failed to create InClusterConfig",
			"error", err)
	}

	log.SetLogger(rzap.New())

	mgr, err := manager.New(config, manager.Options{})
	if err != nil {
		sugar.Fatalw("Failed to create Manager",
			"error", err)
	}

	err = v1alpha2.AddToScheme(mgr.GetScheme())
	if err != nil {
		if err != nil {
			sugar.Fatalw("Failed to add Gateway API scheme",
				"error", err)
		}
	}

	err = sdk.RegisterGatewayClassController(mgr, implementation.NewGatewayClassImplementation(sugar))
	if err != nil {
		sugar.Fatalw("Failed to register GatewayClassController",
			"error", err)
	}

	sugar.Infow("Starting manager")

	ctx, cancel := context.WithCancel(context.Background())
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		sugar.Infow("Terminating because of the signal",
			"signal", <-signalChan)

		cancel()
		os.Exit(0)
	}()

	err = mgr.Start(ctx)
	if err != nil {
		sugar.Fatalw("Failed to start Manager",
			"error", err)
	}
}
