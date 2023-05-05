package main

import (
	"fmt"
	"os"

	flag "github.com/spf13/pflag"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const (
	domain                = "k8s-gateway.nginx.org"
	gatewayClassNameUsage = `The name of the GatewayClass resource. ` +
		`Every NGINX Gateway must have a unique corresponding GatewayClass resource.`
	gatewayCtrlNameUsageFmt = `The name of the Gateway controller. ` +
		`The controller name must be of the form: DOMAIN/PATH. The controller's domain is '%s'`
)

var (
	// Command-line flags
	gatewayCtlrName = flag.String(
		"gateway-ctlr-name",
		"",
		fmt.Sprintf(gatewayCtrlNameUsageFmt, domain),
	)

	gatewayClassName = flag.String("gatewayclass", "", gatewayClassNameUsage)
)

func main() {
	flag.Parse()

	logger := zap.New()

	logger.Info("Starting NGINX Kubernetes Gateway Deployer")

	err := startManager(logger, *gatewayClassName)
	if err != nil {
		logger.Error(err, "Failed to start control loop")
		os.Exit(1)
	}
}
