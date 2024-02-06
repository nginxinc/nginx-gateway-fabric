package static

import (
	"errors"
	"net/http"
	"sync"
)

// newNginxConfiguredOnStartChecker creates a new nginxConfiguredOnStartChecker.
func newNginxConfiguredOnStartChecker() *nginxConfiguredOnStartChecker {
	return &nginxConfiguredOnStartChecker{
		readyCh: make(chan struct{}),
	}
}

// nginxConfiguredOnStartChecker is used to check if nginx is successfully configured and if the NGF Pod is ready.
type nginxConfiguredOnStartChecker struct {
	// firstBatchError is set when the first batch fails to configure nginx
	// and we don't want to set ourselves as ready on the next batch if nothing changes
	firstBatchError error
	// readyCh is a channel that is initialized in newNginxConfiguredOnStartChecker and represents if the NGF Pod is ready.
	readyCh chan struct{}
	lock    sync.RWMutex
	ready   bool
}

// readyCheck returns the ready-state of the Pod. It satisfies the controller-runtime Checker type.
// We are considered ready after the handler processed the first batch. In case there is NGINX configuration
// to write, it must be written and NGINX must be reloaded successfully.
func (h *nginxConfiguredOnStartChecker) readyCheck(_ *http.Request) error {
	h.lock.RLock()
	defer h.lock.RUnlock()

	if !h.ready {
		return errors.New("nginx has not yet become ready to accept traffic")
	}

	return nil
}

// setAsReady marks the health check as ready.
func (h *nginxConfiguredOnStartChecker) setAsReady() {
	h.lock.Lock()
	defer h.lock.Unlock()

	h.ready = true
	h.firstBatchError = nil
	close(h.readyCh)
}

// getReadyCh returns a read-only channel, which determines if the NGF Pod is ready.
func (h *nginxConfiguredOnStartChecker) getReadyCh() <-chan struct{} {
	return h.readyCh
}
