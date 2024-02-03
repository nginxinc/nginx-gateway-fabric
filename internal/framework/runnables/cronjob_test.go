package runnables

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/telemetry/telemetryfakes"
)

func TestCronJob(t *testing.T) {
	g := NewWithT(t)

	healthCollector := &telemetryfakes.FakeHealthChecker{}

	readyChannel := make(chan struct{})
	healthCollector.GetReadyChReturns(readyChannel)

	timeout := 10 * time.Second
	var callCount int

	valCh := make(chan int, 128)
	worker := func(context.Context) {
		callCount++
		valCh <- callCount
	}

	cfg := CronJobConfig{
		Worker:  worker,
		Logger:  zap.New(),
		Period:  1 * time.Millisecond, // 1ms is much smaller than timeout so the CronJob should run a few times
		ReadyCh: healthCollector.GetReadyCh(),
	}
	job := NewCronJob(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	errCh := make(chan error)
	go func() {
		errCh <- job.Start(ctx)
		close(errCh)
	}()
	close(readyChannel)

	minReports := 2 // ensure that the CronJob reports more than once: it doesn't exit after the first run

	g.Eventually(valCh).Should(Receive(BeNumerically(">=", minReports)))

	cancel()
	g.Eventually(errCh).Should(Receive(BeNil()))
	g.Eventually(errCh).Should(BeClosed())
}

func TestCronJob_ContextCanceled(t *testing.T) {
	g := NewWithT(t)

	healthCollector := &telemetryfakes.FakeHealthChecker{}

	readyChannel := make(chan struct{})
	healthCollector.GetReadyChReturns(readyChannel)

	timeout := 10 * time.Second
	var callCount int

	valCh := make(chan int, 128)
	worker := func(context.Context) {
		callCount++
		valCh <- callCount
	}

	cfg := CronJobConfig{
		Worker:  worker,
		Logger:  zap.New(),
		Period:  1 * time.Millisecond, // 1ms is much smaller than timeout so the CronJob should run a few times
		ReadyCh: healthCollector.GetReadyCh(),
	}
	job := NewCronJob(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	errCh := make(chan error)
	go func() {
		errCh <- job.Start(ctx)
		close(errCh)
	}()
	cancel()

	g.Eventually(errCh).Should(Receive())
	g.Eventually(errCh).Should(BeClosed())
}
