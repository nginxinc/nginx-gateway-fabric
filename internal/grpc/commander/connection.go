package commander

import (
	"context"
	"fmt"

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

// connection represents a connection to an agent.
type connection struct {
	cmdExchanger exchanger.CommandExchanger
	logger       logr.Logger
	id           string
	nginxID      string
	systemID     string
	state        State
}

func (c *connection) ID() string {
	return c.id
}

func (c *connection) State() State {
	return c.state
}

// newConnection creates a new instance of connection.
//
// id is the unique ID of the connection.
// cmdExchanger sends and receives commands to and from the CommandChannelServer.
//
// The creator of connection must call its run method in order for the connection send and receive commands.
func newConnection(
	id string,
	logger logr.Logger,
	cmdExchanger exchanger.CommandExchanger,
) *connection {
	return &connection{
		logger:       logger,
		cmdExchanger: cmdExchanger,
		id:           id,
	}
}

// run is a blocking method that kicks off the connection's receive loop and the CommandExchanger's Run loop.
// run will return when the context is canceled or if either loop returns an error.
func (c *connection) run(parent context.Context) error {
	eg, ctx := errgroup.WithContext(parent)

	eg.Go(func() error {
		return c.receive(ctx)
	})

	eg.Go(func() error {
		return c.cmdExchanger.Run(ctx)
	})

	return eg.Wait()
}

func (c *connection) receive(ctx context.Context) error {
	defer func() {
		c.logger.Info("Stopping receive loop")
	}()
	c.logger.Info("Starting receive loop")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case cmd := <-c.cmdExchanger.In():
			c.handleCommand(ctx, cmd)
		}
	}
}

func (c *connection) handleCommand(ctx context.Context, cmd *proto.Command) {
	switch cmd.Data.(type) {
	case *proto.Command_AgentConnectRequest:
		c.handleAgentConnectRequest(ctx, cmd)
	default:
		c.logger.Info("Ignoring command", "command data type", fmt.Sprintf("%T", cmd.Data))
	}
}

func (c *connection) handleAgentConnectRequest(ctx context.Context, cmd *proto.Command) {
	req := cmd.GetAgentConnectRequest()

	c.logger.Info("Received agent connect request", "message ID", cmd.GetMeta().GetMessageId())

	requestStatusCode := proto.AgentConnectStatus_CONNECT_OK
	msg := "Connected"

	if err := c.register(getFirstNginxID(req.GetDetails()), req.GetMeta().SystemUid); err != nil {
		requestStatusCode = proto.AgentConnectStatus_CONNECT_REJECTED_OTHER
		msg = err.Error()

		c.logger.Error(err, "failed to register agent")
	}

	res := createAgentConnectResponseCmd(
		cmd.GetMeta().GetMessageId(),
		requestStatusCode,
		msg,
	)

	select {
	case <-ctx.Done():
		return
	case c.cmdExchanger.Out() <- res:
	}
}

func (c *connection) register(nginxID, systemID string) error {
	if nginxID == "" || systemID == "" {
		c.state = StateInvalid
		return fmt.Errorf("missing nginxID: '%s' and/or systemID: '%s'", nginxID, systemID)
	}

	c.logger.Info("Registering agent", "nginxID", nginxID, "systemID", systemID)

	c.nginxID = nginxID
	c.systemID = systemID
	c.state = StateRegistered

	return nil
}

func createAgentConnectResponseCmd(
	msgID string,
	statusCode proto.AgentConnectStatus_StatusCode,
	statusMsg string,
) *proto.Command {
	return &proto.Command{
		Data: &proto.Command_AgentConnectResponse{
			AgentConnectResponse: &proto.AgentConnectResponse{
				Status: &proto.AgentConnectStatus{
					StatusCode: statusCode,
					Message:    statusMsg,
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
