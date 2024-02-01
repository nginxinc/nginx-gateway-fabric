package runnables

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// CronJobConfig is the configuration for a cronjob.
type CronJobConfig struct {
	// Worker is the function that will be run for every cronjob iteration.
	Worker func(context.Context)
	// Logger is the logger.
	Logger logr.Logger
	// Period defines the period of the cronjob. The cronjob will run every Period.
	Period time.Duration
}

// CronJob periodically runs a worker function.
type CronJob struct {
	cfg CronJobConfig
}

// NewCronJob creates a new cronjob.
func NewCronJob(cfg CronJobConfig) *CronJob {
	return &CronJob{
		cfg: cfg,
	}
}

// Start starts the cronjob.
// Implements controller-runtime manager.Runnable
func (j *CronJob) Start(ctx context.Context) error {
	j.cfg.Logger.Info("Starting cronjob")

	const (
		// 10 min jitter is enough per recommendation for current use cases
		// For the default period of 24 hours, jitter will be 10min /(24*60)min  = 0.0069
		jitterFactor = 10.0 / (24 * 60) // added jitter is bound by jitterFactor * period
		sliding      = true             // This means the period with jitter will be calculated after each worker call.
	)

	wait.JitterUntilWithContext(ctx, j.cfg.Worker, j.cfg.Period, jitterFactor, sliding)

	j.cfg.Logger.Info("Stopping cronjob")
	return nil
}

var _ manager.Runnable = &CronJob{}
