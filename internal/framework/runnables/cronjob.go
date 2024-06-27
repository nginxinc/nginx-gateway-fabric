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
	// ReadyCh delays the start of the job until the channel is closed.
	ReadyCh <-chan struct{}
	// Logger is the logger.
	Logger logr.Logger
	// Period defines the period of the cronjob. The cronjob will run every Period.
	Period time.Duration
	// JitterFactor sets the jitter for the cronjob. If positive, the period is jittered before every
	// run of the worker. If jitterFactor is not positive, the period is unchanged and not jittered.
	JitterFactor float64
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
// Implements controller-runtime manager.Runnable.
func (j *CronJob) Start(ctx context.Context) error {
	select {
	case <-j.cfg.ReadyCh:
	case <-ctx.Done():
		j.cfg.Logger.Info("Context canceled, failed to start cronjob")
		return ctx.Err()
	}

	j.cfg.Logger.Info("Starting cronjob")

	sliding := true // This means the period with jitter will be calculated after each worker call.

	wait.JitterUntilWithContext(ctx, j.cfg.Worker, j.cfg.Period, j.cfg.JitterFactor, sliding)

	j.cfg.Logger.Info("Stopping cronjob")
	return nil
}

var _ manager.Runnable = &CronJob{}
