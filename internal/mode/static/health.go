package static

import (
	"errors"
	"net/http"
)

type healthChecker struct {
	ready bool
}

// readyCheck returns the ready-state of the Pod. It satisfies the controller-runtime Checker type.
// We are considered ready when we have written out the initial nginx configuration and can serve traffic.
func (h *healthChecker) readyCheck(_ *http.Request) error {
	if !h.ready {
		return errors.New("nginx has not yet been configured on startup")
	}

	return nil
}
