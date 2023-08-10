package static

import (
	"encoding/json"
	"fmt"

	"github.com/go-logr/logr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/client-go/tools/record"

	nkgAPI "github.com/nginxinc/nginx-kubernetes-gateway/apis/v1alpha1"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/framework/helpers"
)

// ZapLogLevelSetter defines an interface for setting the logging level of a zap logger.
type ZapLogLevelSetter interface {
	SetLevel(string) error
	Enabled(zapcore.Level) bool
}

type zapSetterImpl struct {
	atomicLevel zap.AtomicLevel
}

func newZapLogLevelSetter(atomicLevel zap.AtomicLevel) zapSetterImpl {
	return zapSetterImpl{
		atomicLevel: atomicLevel,
	}
}

// SetLevel sets the logging level for the zap logger.
func (z zapSetterImpl) SetLevel(level string) error {
	parsedLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		fieldErr := field.NotSupported(
			field.NewPath("logging.level"),
			level,
			[]string{
				string(nkgAPI.ControllerLogLevelInfo),
				string(nkgAPI.ControllerLogLevelDebug),
				string(nkgAPI.ControllerLogLevelError),
			})
		return fieldErr
	}
	z.atomicLevel.SetLevel(parsedLevel)

	return nil
}

// Enabled returns true if the given level is at or above the current level.
func (z zapSetterImpl) Enabled(level zapcore.Level) bool {
	return z.atomicLevel.Enabled(level)
}

// updateControlPlane updates the control plane configuration with the given user spec.
// If any fields are not set within the user spec, the default configuration values are used.
func updateControlPlane(
	cfg *nkgAPI.NginxGateway,
	logger logr.Logger,
	eventRecorder record.EventRecorder,
	configNSName types.NamespacedName,
	logLevelSetter ZapLogLevelSetter,
) error {
	// build up default configuration
	controlConfig := nkgAPI.NginxGatewaySpec{
		Logging: &nkgAPI.Logging{
			Level: helpers.GetPointer(nkgAPI.ControllerLogLevelInfo),
		},
	}

	// by marshaling the user config and then unmarshaling on top of the default config,
	// we ensure that any unset user values are set with the default values
	if cfg != nil {
		cfgBytes, err := json.Marshal(cfg.Spec)
		if err != nil {
			return fmt.Errorf("error marshaling control config: %w", err)
		}

		if err := json.Unmarshal(cfgBytes, &controlConfig); err != nil {
			return fmt.Errorf("error unmarshaling control config: %w", err)
		}
	} else {
		msg := "NginxGateway configuration was deleted; using defaults"
		logger.Error(nil, msg)
		eventRecorder.Event(
			&nkgAPI.NginxGateway{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: configNSName.Namespace,
					Name:      configNSName.Name,
				},
			},
			apiv1.EventTypeWarning,
			"ResourceDeleted",
			msg,
		)
	}

	// set the log level
	return logLevelSetter.SetLevel(string(*controlConfig.Logging.Level))
}
