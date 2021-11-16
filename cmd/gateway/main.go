package main

import (
	"os"

	"github.com/nginxinc/nginx-gateway-kubernetes/internal/config"
	"github.com/nginxinc/nginx-gateway-kubernetes/internal/controller"

	flag "github.com/spf13/pflag"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const (
	domain string = "k8s-gateway.nginx.org"
)

var (
	// Set during go build
	version string
	commit  string
	date    string

	// Command-line flags
	gatewayCtlrName = flag.String("gateway-ctlr-name", "", "The name of the Gateway controller. The controller name must be of the form: DOMAIN/NAMESPACE/NAME. The controller's domain is 'k8s-gateway.nginx.org'.")
)

func main() {
	flag.Parse()

	logger := zap.New()
	conf := config.Config{
		GatewayCtlrName: *gatewayCtlrName,
		Logger:          logger,
	}

	valid := ValidateArguments(
		logger,
		GatewayControllerParam(true, domain, "nginx-gateway" /* TODO dynamically set */, *gatewayCtlrName),
	)
	if !valid {
		flag.PrintDefaults()
		os.Exit(1)
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
