package async

type Promise[V any] struct {
	val  V
	err  error
	done <-chan struct{}
}

func (p *Promise[V]) Get() (V, error) {
	<-p.done
	return p.val, p.err
}
