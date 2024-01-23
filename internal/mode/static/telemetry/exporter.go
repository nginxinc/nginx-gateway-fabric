package telemetry

import (
	"context"

	"github.com/go-logr/logr"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Exporter

// Exporter exports telemetry data to some destination.
// Note: this is a temporary interface. It will be finalized once the Exporter of the common telemetry library
// https://github.com/nginxinc/nginx-gateway-fabric/issues/1318 is implemented.
type Exporter interface {
	Export(ctx context.Context, data Data) error
}

// LoggingExporter logs telemetry data.
type LoggingExporter struct {
	logger logr.Logger
}

// NewLoggingExporter creates a new LoggingExporter.
func NewLoggingExporter(logger logr.Logger) *LoggingExporter {
	return &LoggingExporter{
		logger: logger,
	}
}

// Export logs the provided telemetry data.
func (e *LoggingExporter) Export(_ context.Context, data Data) error {
	e.logger.Info("Exporting telemetry", "data", data)
	return nil
}
