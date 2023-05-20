package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/config"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/manager"
)

const (
	domain                = "k8s-gateway.nginx.org"
	gatewayClassFlag      = "gatewayclass"
	gatewayClassNameUsage = `The name of the GatewayClass resource. ` +
		`Every NGINX Gateway must have a unique corresponding GatewayClass resource.`
	gatewayCtrlNameFlag     = "gateway-ctlr-name"
	gatewayCtrlNameUsageFmt = `The name of the Gateway controller. ` +
		`The controller name must be of the form: DOMAIN/PATH. The controller's domain is '%s'`
)

var (
	// Backing values for common cli flags shared among all subcommands
	// The values are managed by the Root command.
	gatewayCtlrName = stringValidatingValue{
		validator: validateGatewayControllerName,
	}

	gatewayClassName = stringValidatingValue{
		validator: validateResourceName,
	}
)

// stringValidatingValue is a string flag value with custom validation logic.
// stringValidatingValue implements the pflag.Value interface.
type stringValidatingValue struct {
	validator func(v string) error
	value     string
}

func (v *stringValidatingValue) String() string {
	return v.value
}

func (v *stringValidatingValue) Set(param string) error {
	if err := v.validator(param); err != nil {
		return err
	}
	v.value = param
	return nil
}

func (v *stringValidatingValue) Type() string {
	return "string"
}

func createRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use: "gateway",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	rootCmd.PersistentFlags().Var(
		&gatewayCtlrName,
		gatewayCtrlNameFlag,
		fmt.Sprintf(gatewayCtrlNameUsageFmt, domain),
	)
	utilruntime.Must(rootCmd.MarkPersistentFlagRequired(gatewayCtrlNameFlag))

	rootCmd.PersistentFlags().Var(
		&gatewayClassName,
		gatewayClassFlag,
		gatewayClassNameUsage,
	)
	utilruntime.Must(rootCmd.MarkPersistentFlagRequired(gatewayClassFlag))

	return rootCmd
}

func createStaticModeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "static-mode",
		Short: "Configure NGINX in the scope of a single Gateway resource",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := zap.New()
			logger.Info("Starting NGINX Kubernetes Gateway in static mode",
				"version", version,
				"commit", commit,
				"date", date,
			)

			podIP := os.Getenv("POD_IP")
			if err := validateIP(podIP); err != nil {
				return fmt.Errorf("error validating POD_IP environment variable: %w", err)
			}

			conf := config.Config{
				GatewayCtlrName:  gatewayCtlrName.value,
				Logger:           logger,
				GatewayClassName: gatewayClassName.value,
				PodIP:            podIP,
			}

			if err := manager.Start(conf); err != nil {
				return fmt.Errorf("failed to start control loop: %w", err)
			}

			return nil
		},
	}
}

func createProvisionerModeCommand() *cobra.Command {
	return &cobra.Command{
		Use:    "provisioner-mode",
		Short:  "Provision a static-mode NGINX Gateway Deployment per Gateway resource",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := zap.New()
			logger.Info("Starting NGINX Kubernetes Gateway Provisioner",
				"version", version,
				"commit", commit,
				"date", date,
			)

			return errors.New("not implemented yet")
		},
	}
}
