// Package async inspired by: https://github.com/jonbodner/gcon
package async

// Promise represents the result of an async operation.
// A call to Wait() will block until the result is available.
type Promise[V any] struct {
	val  V
	err  error
	done <-chan struct{}
}

// Wait blocks until the Promise's result is available.
func (p *Promise[V]) Wait() (V, error) {
	<-p.done
	return p.val, p.err
}
