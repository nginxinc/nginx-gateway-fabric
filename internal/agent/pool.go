package agent

import (
	"sync"

	"github.com/go-logr/logr"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/grpc/commander"
)

// Pool is a concurrent safe pool of commander.Agents.
type Pool struct {
	agents map[string]commander.Agent
	logger logr.Logger

	lock sync.Mutex
}

// NewPool returns a new instance of Pool.
func NewPool(logger logr.Logger) *Pool {
	return &Pool{
		agents: make(map[string]commander.Agent),
		logger: logger,
	}
}

// AddAgent adds an agent to the Pool.
func (ap *Pool) AddAgent(agent commander.Agent) {
	ap.lock.Lock()
	defer ap.lock.Unlock()

	ap.agents[agent.ID()] = agent

	ap.logger.Info("Added new agent", "id", agent.ID(), "total number of agents", len(ap.agents))
}

// RemoveAgent removes an agent from the Pool with the given ID.
func (ap *Pool) RemoveAgent(id string) {
	ap.lock.Lock()
	defer ap.lock.Unlock()

	delete(ap.agents, id)
	ap.logger.Info("Removed agent", "id", id, "total number of agents", len(ap.agents))
}

// GetAgent returns the agent with the given ID from the Pool.
func (ap *Pool) GetAgent(id string) commander.Agent {
	ap.lock.Lock()
	defer ap.lock.Unlock()

	return ap.agents[id]
}

// Size is used for testing purposes.
func (ap *Pool) Size() int {
	ap.lock.Lock()
	defer ap.lock.Unlock()

	return len(ap.agents)
}
