package static

import (
	"encoding/json"
	"fmt"

	"github.com/go-kit/log"
	"github.com/go-logr/logr"
	"github.com/prometheus/common/promlog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/client-go/tools/record"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
)

// logLevelSetter defines an interface for setting the logging level of a logger.
type logLevelSetter interface {
	SetLevel(string) error
}

// multiLogLevelSetter sets the log level for multiple logLevelSetters.
type multiLogLevelSetter struct {
	setters []logLevelSetter
}

func newMultiLogLevelSetter(setters ...logLevelSetter) multiLogLevelSetter {
	return multiLogLevelSetter{setters: setters}
}

func (m multiLogLevelSetter) SetLevel(level string) error {
	for _, s := range m.setters {
		if err := s.SetLevel(level); err != nil {
			return err
		}
	}

	return nil
}

// zapLogLevelSetter sets the level for a zap logger.
type zapLogLevelSetter struct {
	atomicLevel zap.AtomicLevel
}

func newZapLogLevelSetter(atomicLevel zap.AtomicLevel) zapLogLevelSetter {
	return zapLogLevelSetter{
		atomicLevel: atomicLevel,
	}
}

// SetLevel sets the logging level for the zap logger.
func (z zapLogLevelSetter) SetLevel(level string) error {
	parsedLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		fieldErr := field.NotSupported(
			field.NewPath("logging.level"),
			level,
			[]string{
				string(ngfAPI.ControllerLogLevelInfo),
				string(ngfAPI.ControllerLogLevelDebug),
				string(ngfAPI.ControllerLogLevelError),
			})
		return fieldErr
	}
	z.atomicLevel.SetLevel(parsedLevel)

	return nil
}

// Enabled returns true if the given level is at or above the current level.
func (z zapLogLevelSetter) Enabled(level zapcore.Level) bool {
	return z.atomicLevel.Enabled(level)
}

// leveledPrometheusLogger is a leveled prometheus logger.
// This interface is required because the promlog.NewDynamic returns an unexported type *logger.
type leveledPrometheusLogger interface {
	log.Logger
	SetLevel(level *promlog.AllowedLevel)
}

type promLogLevelSetter struct {
	logger leveledPrometheusLogger
}

func newPromLogLevelSetter(logger leveledPrometheusLogger) promLogLevelSetter {
	return promLogLevelSetter{logger: logger}
}

func newLeveledPrometheusLogger() (leveledPrometheusLogger, error) {
	logFormat := &promlog.AllowedFormat{}

	if err := logFormat.Set("json"); err != nil {
		return nil, err
	}

	logConfig := &promlog.Config{Format: logFormat}
	logger := promlog.NewDynamic(logConfig)

	return logger, nil
}

func (p promLogLevelSetter) SetLevel(level string) error {
	al := &promlog.AllowedLevel{}
	if err := al.Set(level); err != nil {
		fieldErr := field.NotSupported(
			field.NewPath("logging.level"),
			level,
			[]string{
				string(ngfAPI.ControllerLogLevelInfo),
				string(ngfAPI.ControllerLogLevelDebug),
				string(ngfAPI.ControllerLogLevelError),
			})
		return fieldErr
	}

	p.logger.SetLevel(al)
	return nil
}

// updateControlPlane updates the control plane configuration with the given user spec.
// If any fields are not set within the user spec, the default configuration values are used.
func updateControlPlane(
	cfg *ngfAPI.NginxGateway,
	logger logr.Logger,
	eventRecorder record.EventRecorder,
	configNSName types.NamespacedName,
	logLevelSetter logLevelSetter,
) error {
	// build up default configuration
	controlConfig := ngfAPI.NginxGatewaySpec{
		Logging: &ngfAPI.Logging{
			Level: helpers.GetPointer(ngfAPI.ControllerLogLevelInfo),
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
		logger.Info(msg)
		eventRecorder.Event(
			&ngfAPI.NginxGateway{
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
