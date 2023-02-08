package commander

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/nginx/agent/sdk/v2/proto"
	"google.golang.org/grpc/metadata"
)

// nolint:lll
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 github.com/nginx/agent/sdk/v2/proto.Commander_CommandChannelServer
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 github.com/nginx/agent/sdk/v2/proto.Commander_UploadServer
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . ConnectorManager
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Connector

// Connector connects the CommandChannelServer for each agent to the agent connection information.
type Connector interface {
	// ID returns the unique ID of the connector.
	ID() string
	// State returns the State of the connector.
	State() State
	// ReceiveFromUploadServer receives data from the UploadServer.
	// FIXME(kate-osborn): NKG does not need this functionality and ideally we wouldn't have to maintain this method.
	// Figure out how to remove this functionality without causing errors in the agent.
	ReceiveFromUploadServer(server proto.Commander_UploadServer) error
}

// ConnectorManager manages all the active connectors.
type ConnectorManager interface {
	// AddConnector adds the connector to the manager.
	AddConnector(conn Connector)
	// RemoveConnector removes the connector with provided id from the manager.
	// If connector does not exist, the manager does not fail.
	RemoveConnector(id string)
	// GetConnector returns the connector for the provided id.
	GetConnector(id string) Connector
}

const serverUUIDKey = "uuid"

// Commander implements the proto.CommanderServer interface.
type Commander struct {
	connMgr ConnectorManager
	logger  logr.Logger
}

// NewCommander returns a new instance of the Commander.
func NewCommander(logger logr.Logger, connMgr ConnectorManager) *Commander {
	return &Commander{
		logger:  logger,
		connMgr: connMgr,
	}
}

// CommandChannel is a bidirectional streaming channel that is established by the agent and remains open for the
// agent's lifetime.
//
// On every invocation, the Commander will create a new connection with the UUID of the server,
// add the connection to the ConnectorManager, and invoke the connection's blocking Run method.
// If the UUID field is not present in the server's context metadata, no connection is created and an error is returned.
// Once the Run method returns, ths Commander will remove the connection from the ConnectorManager.
// This ensures that only active (connected) connections are tracked by the ConnectorManager.
func (c *Commander) CommandChannel(server proto.Commander_CommandChannelServer) error {
	c.logger.Info("Commander CommandChannel")

	id, err := getUUIDFromContext(server.Context())
	if err != nil {
		c.logger.Error(err, "cannot get the UUID of the agent")
		return err
	}

	c.logger.Info("New connection", "id", id)

	defer func() {
		c.logger.Info("Removing connection from manager")
		c.connMgr.RemoveConnector(id)
	}()

	ch := NewBidirectionalChannel(server, c.logger.WithName("channel"))

	conn := NewConnection(
		id,
		c.logger,
		ch,
	)

	c.logger.Info("Adding connection to manager")
	c.connMgr.AddConnector(conn)

	return conn.run(server.Context())
}

// Download will be implemented in a future PR.
func (c *Commander) Download(_ *proto.DownloadRequest, _ proto.Commander_DownloadServer) error {
	return nil
}

// Upload is not needed by NKG but if we don't implement it the agent will spew error messages.
// An active (connected) connection must exist in the ConnectorManager for the Upload to proceed.
func (c *Commander) Upload(server proto.Commander_UploadServer) error {
	c.logger.Info("Commander Upload requested")

	id, err := getUUIDFromContext(server.Context())
	if err != nil {
		c.logger.Error(err, "cannot get the UUID of the agent")
		return err
	}

	conn := c.connMgr.GetConnector(id)
	if conn == nil {
		return fmt.Errorf("cannot upload; no existing connection for id: %s", id)
	}

	if conn.State() != StateRegistered {
		return fmt.Errorf("cannot upload; agent for connection with id: %s is not registered", id)
	}

	return conn.ReceiveFromUploadServer(server)
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

	return vals[0], nil
}
