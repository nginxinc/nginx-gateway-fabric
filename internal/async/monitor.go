package async

import (
	"context"
	"time"
)

// Matcher is a function that returns true when V matches the desired object.
type Matcher[V any] func(v V) bool

// MonitorWithMatcher monitors the channel until one of the following happens:
// - it receives an item that satisfies the Matcher
// - the context is canceled
// - the timeout is reached.
// It immediately returns a Promise that will contain the result once it is available.
func MonitorWithMatcher[V any](
	ctx context.Context,
	timeout time.Duration,
	ch <-chan V,
	matcher Matcher[V],
) *Promise[V] {
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, timeout)

	done := make(chan struct{})
	p := Promise[V]{
		done: done,
	}

	go func() {
		defer func() {
			close(done)
			cancel()
		}()
		for {
			select {
			case <-ctx.Done():
				var zero V
				p.val, p.err = zero, ctx.Err()
				return
			case v := <-ch:
				if matcher(v) {
					p.val, p.err = v, nil
					return
				}
			}
		}
	}()
	return &p
}
