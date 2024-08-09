package runnables

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestCronJob(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	readyChannel := make(chan struct{})

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
		ReadyCh: readyChannel,
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
	t.Parallel()
	g := NewWithT(t)

	readyChannel := make(chan struct{})

	cfg := CronJobConfig{
		Worker:  func(_ context.Context) {},
		Logger:  zap.New(),
		Period:  1 * time.Millisecond, // 1ms is much smaller than timeout so the CronJob should run a few times
		ReadyCh: readyChannel,
	}
	job := NewCronJob(cfg)

	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error)
	go func() {
		errCh <- job.Start(ctx)
		close(errCh)
	}()

	cancel()
	g.Eventually(errCh).Should(Receive(MatchError(context.Canceled)))
	g.Eventually(errCh).Should(BeClosed())
}
