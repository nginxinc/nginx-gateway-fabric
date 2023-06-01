package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/config"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/manager"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/provisioner"
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
// it implements the pflag.Value interface.
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

// namespacedNameValue is a string flag value that represents a namespaced name.
// it implements the pflag.Value interface.
type namespacedNameValue struct {
	value types.NamespacedName
}

func (v *namespacedNameValue) String() string {
	if (v.value == types.NamespacedName{}) {
		// if we don't do that, the default value in the help message will be printed as "/"
		return ""
	}
	return v.value.String()
}

func (v *namespacedNameValue) Set(param string) error {
	nsname, err := parseNamespacedResourceName(param)
	if err != nil {
		return err
	}

	v.value = nsname
	return nil
}

func (v *namespacedNameValue) Type() string {
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
	const gatewayFlag = "gateway"

	// flag values
	gateway := namespacedNameValue{}
	var updateGCStatus bool

	cmd := &cobra.Command{
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

			var gwNsName *types.NamespacedName
			if cmd.Flags().Changed(gatewayFlag) {
				gwNsName = &gateway.value
			}

			conf := config.Config{
				GatewayCtlrName:          gatewayCtlrName.value,
				Logger:                   logger,
				GatewayClassName:         gatewayClassName.value,
				PodIP:                    podIP,
				GatewayNsName:            gwNsName,
				UpdateGatewayClassStatus: updateGCStatus,
			}

			if err := manager.Start(conf); err != nil {
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

	cmd.Flags().BoolVar(
		&updateGCStatus,
		"update-gatewayclass-status",
		true,
		"Update the status of the GatewayClass resource.",
	)

	return cmd
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

			return provisioner.StartManager(provisioner.Config{
				Logger:           logger,
				GatewayClassName: gatewayClassName.value,
			})
		},
	}
}
