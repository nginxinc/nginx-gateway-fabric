package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
	ctlrZap "sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/mode/provisioner"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/mode/static"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/mode/static/config"
)

const (
	domain                = "gateway.nginx.org"
	gatewayClassFlag      = "gatewayclass"
	gatewayClassNameUsage = `The name of the GatewayClass resource. ` +
		`Every NGINX Gateway must have a unique corresponding GatewayClass resource.`
	gatewayCtrlNameFlag     = "gateway-ctlr-name"
	gatewayCtrlNameUsageFmt = `The name of the Gateway controller. ` +
		`The controller name must be of the form: DOMAIN/PATH. The controller's domain is '%s'`
	gatewayFlag = "gateway"
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

	// Backing values for static subcommand cli flags.
	updateGCStatus bool
	disableMetrics bool
	metricsSecure  bool
	disableHealth  bool

	metricsListenPort = intValidatingValue{
		validator: validatePort,
		value:     9113,
	}
	healthListenPort = intValidatingValue{
		validator: validatePort,
		value:     8081,
	}
	gateway    = namespacedNameValue{}
	configName = stringValidatingValue{
		validator: validateResourceName,
	}
)

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
	cmd := &cobra.Command{
		Use:   "static-mode",
		Short: "Configure NGINX in the scope of a single Gateway resource",
		RunE: func(cmd *cobra.Command, args []string) error {
			atom := zap.NewAtomicLevel()

			logger := ctlrZap.New(ctlrZap.Level(atom))
			logger.Info(
				"Starting NGINX Kubernetes Gateway in static mode",
				"version", version,
				"commit", commit,
				"date", date,
			)
			log.SetLogger(logger)

			if err := ensureNoPortCollisions(metricsListenPort.value, healthListenPort.value); err != nil {
				return fmt.Errorf("error validating ports: %w", err)
			}

			podIP := os.Getenv("POD_IP")
			if err := validateIP(podIP); err != nil {
				return fmt.Errorf("error validating POD_IP environment variable: %w", err)
			}

			namespace := os.Getenv("MY_NAMESPACE")
			if namespace == "" {
				return errors.New("MY_NAMESPACE environment variable must be set")
			}

			var gwNsName *types.NamespacedName
			if cmd.Flags().Changed(gatewayFlag) {
				gwNsName = &gateway.value
			}

			metricsConfig := config.MetricsConfig{}
			if !disableMetrics {
				metricsConfig.Enabled = true
				metricsConfig.Port = metricsListenPort.value
				metricsConfig.Secure = metricsSecure
			}

			conf := config.Config{
				GatewayCtlrName:          gatewayCtlrName.value,
				ConfigName:               configName.String(),
				Logger:                   logger,
				AtomicLevel:              atom,
				GatewayClassName:         gatewayClassName.value,
				PodIP:                    podIP,
				Namespace:                namespace,
				GatewayNsName:            gwNsName,
				UpdateGatewayClassStatus: updateGCStatus,
				MetricsConfig:            metricsConfig,
				HealthConfig: config.HealthConfig{
					Enabled: !disableHealth,
					Port:    healthListenPort.value,
				},
			}

			if err := static.StartManager(conf); err != nil {
				return fmt.Errorf("failed to start control loop: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().Var(
		&gateway,
		gatewayFlag,
		"The namespaced name of the Gateway resource to use. "+
			"Must be of the form: NAMESPACE/NAME. "+
			"If not specified, the control plane will process all Gateways for the configured GatewayClass. "+
			"However, among them, it will choose the oldest resource by creation timestamp. If the timestamps are "+
			"equal, it will choose the resource that appears first in alphabetical order by {namespace}/{name}.",
	)

	cmd.Flags().VarP(
		&configName,
		"config",
		"c",
		`The name of the NginxGateway resource to be used for this controller's dynamic configuration.`+
			` Lives in the same Namespace as the controller.`,
	)

	cmd.Flags().BoolVar(
		&updateGCStatus,
		"update-gatewayclass-status",
		true,
		"Update the status of the GatewayClass resource.",
	)

	cmd.Flags().BoolVar(
		&disableMetrics,
		"metrics-disable",
		false,
		"Disable exposing metrics in the Prometheus format.",
	)

	cmd.Flags().Var(
		&metricsListenPort,
		"metrics-port",
		"Set the port where the metrics are exposed. Format: [1024 - 65535]",
	)

	cmd.Flags().BoolVar(
		&metricsSecure,
		"metrics-secure-serving",
		false,
		"Enable serving metrics via https. By default metrics are served via http."+
			" Please note that this endpoint will be secured with a self-signed certificate.",
	)

	cmd.Flags().BoolVar(
		&disableHealth,
		"health-disable",
		false,
		"Disable running the health probe server.",
	)

	cmd.Flags().Var(
		&healthListenPort,
		"health-port",
		"Set the port where the health probe server is exposed. Format: [1024 - 65535]",
	)

	return cmd
}

func createProvisionerModeCommand() *cobra.Command {
	return &cobra.Command{
		Use:    "provisioner-mode",
		Short:  "Provision a static-mode NGINX Gateway Deployment per Gateway resource",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := ctlrZap.New()
			logger.Info(
				"Starting NGINX Kubernetes Gateway Provisioner",
				"version", version,
				"commit", commit,
				"date", date,
			)

			return provisioner.StartManager(provisioner.Config{
				Logger:           logger,
				GatewayClassName: gatewayClassName.value,
				GatewayCtlrName:  gatewayCtlrName.value,
			})
		},
	}
}
