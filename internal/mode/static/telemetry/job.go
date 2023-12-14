package telemetry

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/util/wait"
)

// Data type (any or not) is yet to be determined
type Exporter interface {
	Export(ctx context.Context, data any) error
}

type StdoutExporter struct{}

func (s *StdoutExporter) Export(_ context.Context, data any) error {
	fmt.Printf("exporting data: %+v\n", data)
	return nil
}

type Data struct{}

type Job struct {
	exporter Exporter
	logger   logr.Logger
}

func NewJob(exporter Exporter, logger logr.Logger) *Job {
	return &Job{
		exporter: exporter,
		logger:   logger,
	}
}

func (j *Job) Start(ctx context.Context) error {
	j.logger.Info("starting telemetry job")
	wait.UntilWithContext(ctx, j.report, 10*time.Second)
	j.logger.Info("stopping telemetry job")
	return nil
}

func (j *Job) report(ctx context.Context) {
	j.logger.Info("reporting telemetry")

	err := j.exporter.Export(ctx, Data{})
	if err != nil {
		j.logger.Error(err, "failed to export telemetry")
	}
}
