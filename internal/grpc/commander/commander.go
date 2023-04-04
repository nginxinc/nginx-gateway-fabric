package commander

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/go-logr/logr"
	"github.com/nginx/agent/sdk/v2/proto"
	"google.golang.org/grpc/metadata"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/agent"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/observer"
)

// nolint:lll
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 github.com/nginx/agent/sdk/v2/proto.Commander_CommandChannelServer
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 github.com/nginx/agent/sdk/v2/proto.Commander_UploadServer

const serverUUIDKey = "uuid"

type observedNginxConfig interface {
	observer.Subject
	GetLatestConfig() *agent.NginxConfig
}

// Commander implements the proto.CommanderServer interface.
type Commander struct {
	connections    map[string]*connection
	observedConfig observedNginxConfig
	logger         logr.Logger

	connLock sync.Mutex
}

// NewCommander returns a new instance of the Commander.
func NewCommander(logger logr.Logger, observedConfig observedNginxConfig) *Commander {
	return &Commander{
		logger:         logger,
		connections:    make(map[string]*connection),
		observedConfig: observedConfig,
	}
}

// CommandChannel is a bidirectional streaming channel that is established by the agent and remains open for the
// agent's lifetime.
//
// On every invocation, the Commander will create a new connection with the UUID of the server,
// add the connection to the AgentManager, and invoke the connection's blocking Run method.
// If the UUID field is not present in the server's context metadata, no connection is created and an error is returned.
// Once the Run method returns, ths Commander will remove the connection from the AgentManager.
// This ensures that only active (connected) connections are tracked by the AgentManager.
func (c *Commander) CommandChannel(server proto.Commander_CommandChannelServer) error {
	c.logger.Info("Commander CommandChannel")

	id, err := getUUIDFromContext(server.Context())
	if err != nil {
		c.logger.Error(err, "cannot get the UUID of the agent")
		return err
	}

	defer func() {
		c.removeConnection(id)
	}()

	idLogger := c.logger.WithValues("id", id)

	conn := newConnection(
		id,
		idLogger.WithName("connection"),
		NewBidirectionalChannel(server, idLogger.WithName("channel")),
		c.observedConfig,
	)

	c.addConnection(conn)

	return conn.run(server.Context())
}

// Download implements the Download method of the Commander gRPC service. An agent invokes this method to download the
// latest version of the NGINX configuration.
func (c *Commander) Download(request *proto.DownloadRequest, server proto.Commander_DownloadServer) error {
	c.logger.Info("Download requested", "message ID", request.GetMeta().GetMessageId())

	id, err := getUUIDFromContext(server.Context())
	if err != nil {
		c.logger.Error(err, "failed download")
		return err
	}

	conn := c.getConnection(id)
	if conn == nil {
		err := fmt.Errorf("connection with id: %s not found", id)
		c.logger.Error(err, "failed download")
		return err
	}

	// TODO: can there be a race condition here?
	if conn.State() != StateRegistered {
		err := fmt.Errorf("connection with id: %s is not registered", id)
		c.logger.Error(err, "failed upload")
		return err
	}

	return conn.sendConfig(request, server)
}

// Upload implements the Upload method of the Commander gRPC service.
// FIXME(kate-osborn): NKG doesn't need this functionality and ideally we wouldn't have to implement and maintain this.
// Figure out how to remove this without causing errors in the agent.
func (c *Commander) Upload(server proto.Commander_UploadServer) error {
	c.logger.Info("Commander Upload requested")

	id, err := getUUIDFromContext(server.Context())
	if err != nil {
		c.logger.Error(err, "failed upload; cannot get the UUID of the conn")
		return err
	}

	conn := c.getConnection(id)
	if conn == nil {
		err := fmt.Errorf("connection with id: %s not found", id)
		c.logger.Error(err, "failed upload")
		return err
	}

	// TODO: can there be a race condition here?
	if conn.State() != StateRegistered {
		err := fmt.Errorf("connection with id: %s is not registered", id)
		c.logger.Error(err, "failed upload")
		return err
	}

	return conn.receiveFromUploadServer(server)
}

func (c *Commander) removeConnection(id string) {
	c.connLock.Lock()
	defer c.connLock.Unlock()

	delete(c.connections, id)
	c.logger.Info("removed connection", "id", id, "total connections", len(c.connections))
}

func (c *Commander) addConnection(conn *connection) {
	c.connLock.Lock()
	defer c.connLock.Unlock()

	c.connections[conn.id] = conn
	c.logger.Info("added connection", "id", conn.id, "total connections", len(c.connections))
}

func (c *Commander) getConnection(id string) *connection {
	c.connLock.Lock()
	defer c.connLock.Unlock()

	return c.connections[id]
}

func getUUIDFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("metadata is not provided")
	}

	vals := md.Get(serverUUIDKey)
	if len(vals) == 0 {
		return "", errors.New("uuid is not in metadata")
	}

	if len(vals) > 1 {
		return "", errors.New("more than one value for uuid in metadata")
	}

	return vals[0], nil
}
