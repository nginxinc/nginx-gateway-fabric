package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/gateway-api/pkg/client/clientset/gateway/versioned"
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

	gatewayClient, err := versioned.NewForConfig(config)
	if err != nil {
		sugar.Fatalw("Failed to create a client for Gateway APIs",
			"error", err)
	}

	gc, err := gatewayClient.GatewayV1alpha2().GatewayClasses().Get(context.TODO(), *gatewayClass, meta_v1.GetOptions{})
	if err != nil {
		sugar.Fatalw("Failed to get the GatewayClass",
			"name", *gatewayClass,
			"error", err)
	}

	if gc.Spec.ControllerName != "k8s-gateway.nginx.org/gateway" {
		sugar.Fatalw("Wrong ControllerName in the GatewayClass resource",
			"expected", "k8s-gateway.nginx.org/gateway",
			"got", "gc.Spec.ControllerName")
	}

	sugar.Infow("Gateway class info",
		"name", gc.Name,
		"creation timestamp", gc.CreationTimestamp)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		sugar.Infow("Terminating because of the signal",
			"signal", <-signalChan)

		os.Exit(0)
	}()

	for {
		sugar.Infow("Gateway is running")
		time.Sleep(30 * time.Second)
	}
}
