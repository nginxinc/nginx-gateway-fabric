package static

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type leaderOnlyRunnable struct {
	electedCh chan struct{}
	runnable  manager.Runnable
}

func newLeaderOnlyRunnable(electedCh chan struct{}, runnable manager.Runnable) *leaderOnlyRunnable {
	return &leaderOnlyRunnable{
		electedCh: electedCh,
		runnable:  runnable,
	}
}

func (l *leaderOnlyRunnable) Start(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return nil
	case <-l.electedCh:
		break
	}

	return l.runnable.Start(ctx)
}
