package telemetry

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// DataCollector collects telemetry data.
type DataCollector interface {
	Collect(ctx context.Context) (Data, error)
}

// JobConfig is the configuration for the telemetry job.
type JobConfig struct {
	// Exporter is the exporter to use for exporting telemetry data.
	Exporter Exporter
	// DataCollector is the collector to use for collecting telemetry data.
	DataCollector DataCollector
	// Logger is the logger.
	Logger logr.Logger
	// Period defines the period of the telemetry job. The job will run every Period.
	Period time.Duration
}

// Job periodically exports telemetry data using the provided exporter.
type Job struct {
	cfg JobConfig
}

// NewJob creates a new telemetry job.
func NewJob(cfg JobConfig) *Job {
	return &Job{
		cfg: cfg,
	}
}

// Start starts the telemetry job.
// Implements controller-runtime manager.Runnable
func (j *Job) Start(ctx context.Context) error {
	j.cfg.Logger.Info("Starting telemetry job")

	report := func(ctx context.Context) {
		// Gather telemetry
		j.cfg.Logger.V(1).Info("Gathering telemetry")

		// We will need to gather data as defined in https://github.com/nginxinc/nginx-gateway-fabric/issues/793
		data, err := j.cfg.DataCollector.Collect(ctx)
		if err != nil {
			j.cfg.Logger.Error(err, "Failed to collect telemetry")
		}

		// Export telemetry
		j.cfg.Logger.V(1).Info("Exporting telemetry")

		if err := j.cfg.Exporter.Export(ctx, data); err != nil {
			j.cfg.Logger.Error(err, "Failed to export telemetry")
		}
	}

	const (
		// 10 min jitter is enough per telemetry destination recommendation
		// For the default period of 24 hours, jitter will be 10min /(24*60)min  = 0.0069
		jitterFactor = 10.0 / (24 * 60) // added jitter is bound by jitterFactor * period
		sliding      = true             // This means the period with jitter will be calculated after each report() call.
	)

	wait.JitterUntilWithContext(ctx, report, j.cfg.Period, jitterFactor, sliding)

	j.cfg.Logger.Info("Stopping telemetry job")
	return nil
}

var _ manager.Runnable = &Job{}
