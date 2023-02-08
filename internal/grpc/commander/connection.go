package commander

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/go-logr/logr"
	"github.com/nginx/agent/sdk/v2/grpc"
	"github.com/nginx/agent/sdk/v2/proto"
	"golang.org/x/sync/errgroup"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/grpc/commander/exchanger"
)

// State is the state of the connection.
type State int

const (
	// StateConnected means the connection is active (connected) but is not registered.
	StateConnected State = iota
	// StateRegistered means the connection is active and registered.
	StateRegistered
	// StateInvalid means the connection is active and attempted to register, but its registration info was invalid.
	// We cannot push config to a connection in this state.
	StateInvalid
)

// Connection represents a connection to an agent.
type Connection struct {
	cmdExchanger exchanger.CommandExchanger
	logger       logr.Logger
	id           string
	nginxID      string
	systemID     string
	state        State
}

// NewConnection creates a new instance of Connection.
//
// id is the unique ID of the connection.
// cmdExchanger sends and receives commands to and from the CommandChannelServer.
//
// The creator of Connection must call its run method in order for the Connection send and receive commands.
func NewConnection(
	id string,
	logger logr.Logger,
	cmdExchanger exchanger.CommandExchanger,
) *Connection {
	return &Connection{
		logger:       logger.WithName(fmt.Sprintf("connection-%s", id)),
		cmdExchanger: cmdExchanger,
		id:           id,
	}
}

// run is a blocking method that kicks off the Connection's receive loop and the CommandExchanger's Run loop.
// run will return when the context is canceled or if either loop returns an error.
func (c *Connection) run(parent context.Context) error {
	eg, ctx := errgroup.WithContext(parent)

	eg.Go(func() error {
		return c.receive(ctx)
	})

	eg.Go(func() error {
		return c.cmdExchanger.Run(ctx)
	})

	return eg.Wait()
}

// State returns the state of the Connection.
func (c *Connection) State() State {
	return c.state
}

// ID returns the unique ID of the Connection.
func (c *Connection) ID() string {
	return c.id
}

func (c *Connection) receive(ctx context.Context) error {
	defer func() {
		c.logger.Info("Stopping receive loop")
	}()
	c.logger.Info("Starting receive loop")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case cmd := <-c.cmdExchanger.Out():
			c.handleCommand(cmd)
		}
	}
}

// ReceiveFromUploadServer uploads data chunks from the UploadServer and logs them.
// FIXME(kate-osborn): NKG doesn't need this functionality and ideally we wouldn't have to implement and maintain this.
// Figure out how to remove this without causing errors in the agent.
func (c *Connection) ReceiveFromUploadServer(server proto.Commander_UploadServer) error {
	c.logger.Info("Upload request")

	for {
		_, err := server.Recv()

		if err != nil && !errors.Is(err, io.EOF) {
			c.logger.Error(err, "upload receive error")
			return err
		}

		c.logger.Info("Received chunk from upload channel")

		if errors.Is(err, io.EOF) {
			c.logger.Info("Upload completed")
			return server.SendAndClose(&proto.UploadStatus{Status: proto.UploadStatus_OK})
		}
	}
}

func (c *Connection) handleCommand(cmd *proto.Command) {
	if cmd == nil {
		return
	}

	switch cmd.Data.(type) {
	case *proto.Command_AgentConnectRequest:
		c.handleAgentConnectRequest(cmd)
	case *proto.Command_DataplaneStatus:
		c.handleDataplaneStatus(cmd)
	default:
		c.logger.Info("Ignoring command", "command data type", fmt.Sprintf("%T", cmd.Data))
	}
}

func (c *Connection) handleAgentConnectRequest(cmd *proto.Command) {
	req := cmd.GetAgentConnectRequest()

	c.logger.Info("Received agent connect request", "message ID", cmd.GetMeta().GetMessageId())

	c.register(getFirstNginxID(req.GetDetails()), req.GetMeta().SystemUid)

	res := createAgentConnectResponseCmd(
		cmd.GetMeta().GetMessageId(),
		c.nginxID,
		c.systemID,
	)

	c.cmdExchanger.In() <- res
}

func (c *Connection) handleDataplaneStatus(cmd *proto.Command) {
	status := cmd.GetDataplaneStatus()
	c.logger.Info("Received a dataplane status command", "status", status)

	// FIXME(kate-osborn): this check is required because on a controller restart event the agent will re-connect but
	// not re-register. This means we have to register the agent using the information in the first dataplane status
	// command we receive.
	if c.state != StateRegistered {
		c.register(getFirstNginxID(status.GetDetails()), status.SystemId)
	}
}

func (c *Connection) register(nginxID, systemID string) {
	if nginxID == "" || systemID == "" {
		c.state = StateInvalid
		c.logger.Info(
			"Cannot register agent; nginxID and systemID must be provided",
			"nginxID",
			nginxID,
			"systemID",
			systemID,
		)

		return
	}

	c.logger.Info("Registering agent", "nginxID", nginxID, "systemID", systemID)

	c.nginxID = nginxID
	c.systemID = systemID
	c.state = StateRegistered
}

func createAgentConnectResponseCmd(msgID, nginxID, systemID string) *proto.Command {
	return &proto.Command{
		Data: &proto.Command_AgentConnectResponse{
			AgentConnectResponse: &proto.AgentConnectResponse{
				AgentConfig: &proto.AgentConfig{
					Configs: &proto.ConfigReport{
						Meta: grpc.NewMessageMeta(msgID),
						Configs: []*proto.ConfigDescriptor{
							{
								NginxId:  nginxID,
								SystemId: systemID,
							},
						},
					},
				},
				Status: &proto.AgentConnectStatus{
					StatusCode: proto.AgentConnectStatus_CONNECT_OK,
					Message:    "Connected",
				},
			},
		},
		Meta: grpc.NewMessageMeta(msgID),
		Type: proto.Command_NORMAL,
	}
}

func getFirstNginxID(details []*proto.NginxDetails) (id string) {
	if len(details) > 0 {
		id = details[0].GetNginxId()
	}

	return
}
