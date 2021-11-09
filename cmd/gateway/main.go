package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/nginxinc/nginx-gateway-kubernetes/internal/implementation"
	"github.com/nginxinc/nginx-gateway-kubernetes/pkg/sdk"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
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

	logger := zap.New()

	logger.Info("Starting NGINX Gateway",
		"version", version,
		"commit", commit,
		"date", date)

	config, err := rest.InClusterConfig()
	if err != nil {
		logger.Error(err, "Failed to create InClusterConfig")
		os.Exit(1)
	}

	mgr, err := manager.New(config, manager.Options{
		Logger: logger,
	})
	if err != nil {
		logger.Error(err, "Failed to create Manager")
		os.Exit(1)
	}

	err = v1alpha2.AddToScheme(mgr.GetScheme())
	if err != nil {
		logger.Error(err, "Failed to add Gateway API scheme")
		os.Exit(1)
	}

	err = sdk.RegisterGatewayClassController(mgr, implementation.NewGatewayClassImplementation(logger))
	if err != nil {
		logger.Error(err, "Failed to register GatewayClassController")
		os.Exit(1)
	}

	logger.Info("Starting manager")

	err = mgr.Start(signals.SetupSignalHandler())
	if err != nil {
		logger.Error(err, "Failed to start Manager")
		os.Exit(1)
	}
}
