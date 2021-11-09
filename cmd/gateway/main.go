package main

import (
	"os"

	"github.com/nginxinc/nginx-gateway-kubernetes/internal/config"
	"github.com/nginxinc/nginx-gateway-kubernetes/internal/controller"

	flag "github.com/spf13/pflag"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	// Set during go build
	version string
	commit  string
	date    string

	// Command-line flags
	gatewayCtlrName = flag.String("gateway-ctlr-name", "", "The name of the Gateway controller")
)

func main() {
	flag.Parse()

	if *gatewayCtlrName == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	logger := zap.New()
	conf := config.Config{
		GatewayCtlrName: *gatewayCtlrName,
		Logger:          logger,
	}

	logger.Info("Starting NGINX Gateway",
		"version", version,
		"commit", commit,
		"date", date)

	err := controller.Start(conf)
	if err != nil {
		logger.Error(err, "Failed to start control loop")
		os.Exit(1)
	}
}
