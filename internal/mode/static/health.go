package static

import (
	"errors"
	"net/http"
	"sync"
)

type healthChecker struct {
	// firstBatchError is set when the first batch fails to configure nginx
	// and we don't want to set ourselves as ready on the next batch if nothing changes
	firstBatchError error
	lock            sync.RWMutex
	ready           bool
}

// readyCheck returns the ready-state of the Pod. It satisfies the controller-runtime Checker type.
// We are considered ready after the handler processed the first batch. In case there is NGINX configuration
// to write, it must be written and NGINX must be reloaded successfully.
func (h *healthChecker) readyCheck(_ *http.Request) error {
	h.lock.RLock()
	defer h.lock.RUnlock()

	if !h.ready {
		return errors.New("nginx has not yet become ready to accept traffic")
	}

	return nil
}

// setAsReady marks the health check as ready.
func (h *healthChecker) setAsReady() {
	h.lock.Lock()
	defer h.lock.Unlock()

	h.ready = true
	h.firstBatchError = nil
}
