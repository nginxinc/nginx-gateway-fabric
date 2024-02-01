package runnables

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// JobConfig is the configuration for a job.
type JobConfig struct {
	// Worker is the function that will be run for every job iteration.
	Worker func(context.Context)
	// Logger is the logger.
	Logger logr.Logger
	// Period defines the period of the job. The job will run every Period.
	Period time.Duration
}

// Job periodically runs a worker function.
type Job struct {
	cfg JobConfig
}

// NewJob creates a new job.
func NewJob(cfg JobConfig) *Job {
	return &Job{
		cfg: cfg,
	}
}

// Start starts the job.
// Implements controller-runtime manager.Runnable
func (j *Job) Start(ctx context.Context) error {
	j.cfg.Logger.Info("Starting job")

	const (
		// 10 min jitter is enough per telemetry destination recommendation
		// For the default period of 24 hours, jitter will be 10min /(24*60)min  = 0.0069
		jitterFactor = 10.0 / (24 * 60) // added jitter is bound by jitterFactor * period
		sliding      = true             // This means the period with jitter will be calculated after each worker call.
	)

	wait.JitterUntilWithContext(ctx, j.cfg.Worker, j.cfg.Period, jitterFactor, sliding)

	j.cfg.Logger.Info("Stopping job")
	return nil
}

var _ manager.Runnable = &Job{}
