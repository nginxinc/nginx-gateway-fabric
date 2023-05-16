package main

import (
	"errors"
	"fmt"
	"net"
	"os"

	flag "github.com/spf13/pflag"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/config"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/manager"
)

const (
	domain                = "k8s-gateway.nginx.org"
	gatewayClassNameUsage = `The name of the GatewayClass resource. ` +
		`Every NGINX Gateway must have a unique corresponding GatewayClass resource.`
	gatewayCtrlNameUsageFmt = `The name of the Gateway controller. ` +
		`The controller name must be of the form: DOMAIN/PATH. The controller's domain is '%s'`
)

var (
	// Set during go build
	version string
	commit  string
	date    string

	// Command-line flags
	gatewayCtlrName = flag.String(
		"gateway-ctlr-name",
		"",
		fmt.Sprintf(gatewayCtrlNameUsageFmt, domain),
	)

	gatewayClassName = flag.String("gatewayclass", "", gatewayClassNameUsage)

	// Environment variables
	podIP = os.Getenv("POD_IP")
)

func validateIP(podIP string) error {
	if podIP == "" {
		return errors.New("IP address must be set")
	}
	if net.ParseIP(podIP) == nil {
		return fmt.Errorf("%q must be a valid IP address", podIP)
	}

	return nil
}

func main() {
	flag.Parse()

	MustValidateArguments(
		flag.CommandLine,
		GatewayControllerParam(domain),
		GatewayClassParam(),
	)

	if err := validateIP(podIP); err != nil {
		fmt.Printf("error validating POD_IP environment variable: %v\n", err)
		os.Exit(1)
	}

	logger := zap.New()
	conf := config.Config{
		GatewayCtlrName:  *gatewayCtlrName,
		Logger:           logger,
		GatewayClassName: *gatewayClassName,
		PodIP:            podIP,
	}

	logger.Info("Starting NGINX Kubernetes Gateway",
		"version", version,
		"commit", commit,
		"date", date,
	)

	err := manager.Start(conf)
	if err != nil {
		logger.Error(err, "Failed to start control loop")
		os.Exit(1)
	}
}
