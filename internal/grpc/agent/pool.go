package agent

import (
	"sync"

	"github.com/go-logr/logr"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/grpc/commander"
)

// Pool is a concurrent safe pool of commander.Connectors.
type Pool struct {
	connectors map[string]commander.Connector
	logger     logr.Logger

	lock sync.Mutex
}

// NewPool returns a new instance of Pool.
func NewPool(logger logr.Logger) *Pool {
	return &Pool{
		connectors: make(map[string]commander.Connector),
		logger:     logger,
	}
}

// AddConnector adds a connector to the Pool.
func (ap *Pool) AddConnector(conn commander.Connector) {
	ap.lock.Lock()
	defer ap.lock.Unlock()

	ap.connectors[conn.ID()] = conn

	ap.logger.Info("Added new connector", "id", conn.ID(), "total number of connectors", len(ap.connectors))
}

// RemoveConnector removes a connector from the Pool with the given ID.
func (ap *Pool) RemoveConnector(id string) {
	ap.lock.Lock()
	defer ap.lock.Unlock()

	delete(ap.connectors, id)
	ap.logger.Info("Removed connector", "id", id, "total number of connectors", len(ap.connectors))
}

// GetConnector returns the connector with the given ID from the Pool.
func (ap *Pool) GetConnector(id string) commander.Connector {
	ap.lock.Lock()
	defer ap.lock.Unlock()

	return ap.connectors[id]
}

// Size is used for testing purposes.
func (ap *Pool) Size() int {
	ap.lock.Lock()
	defer ap.lock.Unlock()

	return len(ap.connectors)
}
