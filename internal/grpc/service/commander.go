package service

import (
	"errors"
	"fmt"
	"io"

	"github.com/go-logr/logr"
	"github.com/nginx/agent/sdk/v2/grpc"
	"github.com/nginx/agent/sdk/v2/proto"
)

// Commander implements the proto.CommanderServer interface.
// This code is for demo purposes only. It's the least amount of code I could write to demonstrate that the agent and
// control plane can communicate with each other. It is not the final version, so it isn't tested or commented.
// Code is a version of https://github.com/nginx/agent/blob/main/sdk/examples/services/command_service.go
type Commander struct {
	toClient chan *proto.Command
	logger   logr.Logger
}

func NewCommander(logger logr.Logger) *Commander {
	return &Commander{
		toClient: make(chan *proto.Command),
		logger:   logger,
	}
}

func (c *Commander) CommandChannel(stream proto.Commander_CommandChannelServer) error {
	c.logger.Info("Commander CommandChannel")

	go c.handleReceive(stream)

	for {
		select {
		case out := <-c.toClient:
			err := stream.Send(out)
			if errors.Is(err, io.EOF) {
				c.logger.Info("CommandChannel EOF")
				return nil
			}
			if err != nil {
				c.logger.Error(err, "failed to send outgoing command")
				continue
			}
		case <-stream.Context().Done():
			c.logger.Info("CommandChannel complete")
			return nil
		}
	}
}

func (c *Commander) Download(request *proto.DownloadRequest, _ proto.Commander_DownloadServer) error {
	c.logger.Info("Commander Download requested", "request", request.GetMeta())

	return nil
}

func (c *Commander) Upload(upload proto.Commander_UploadServer) error {
	c.logger.Info("Commander Upload requested")

	for {
		chunk, err := upload.Recv()

		if err != nil && !errors.Is(err, io.EOF) {
			c.logger.Error(err, "upload receive error")
			return err
		}

		c.logger.Info("Received chunk from upload channel", "chunk", chunk)

		if errors.Is(err, io.EOF) {
			c.logger.Info("Commander Upload completed")
			return upload.SendAndClose(&proto.UploadStatus{Status: proto.UploadStatus_OK})
		}
	}
}

func (c *Commander) handleReceive(server proto.Commander_CommandChannelServer) {
	for {
		cmd, err := server.Recv()
		if err != nil {
			c.logger.Error(err, "failed to receive command from CommandChannelServer")
			return
		}

		c.handleCommand(cmd)
	}
}

func (c *Commander) handleCommand(cmd *proto.Command) {
	if cmd != nil {
		switch commandData := cmd.Data.(type) {
		// The only command we care about right now is the AgentConnectRequest.
		case *proto.Command_AgentConnectRequest:
			c.logger.Info("Received a connection request from an agent", "data", commandData.AgentConnectRequest.GetMeta())
			c.sendAgentConnectResponse(cmd)
		default:
			c.logger.Info("ignoring command", "command data type", fmt.Sprintf("%T", cmd.Data))
		}
	}
}

func (c *Commander) sendAgentConnectResponse(cmd *proto.Command) {
	// get first nginx id for example
	nginxID := "0"
	if len(cmd.GetAgentConnectRequest().GetDetails()) > 0 {
		nginxID = cmd.GetAgentConnectRequest().GetDetails()[0].GetNginxId()
	}
	response := &proto.Command{
		Data: &proto.Command_AgentConnectResponse{
			AgentConnectResponse: &proto.AgentConnectResponse{
				AgentConfig: &proto.AgentConfig{
					Configs: &proto.ConfigReport{
						Meta: grpc.NewMessageMeta(cmd.Meta.MessageId),
						Configs: []*proto.ConfigDescriptor{
							{
								Checksum: "",
								NginxId:  nginxID,
								SystemId: cmd.GetAgentConnectRequest().GetMeta().GetSystemUid(),
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
		Meta: grpc.NewMessageMeta(cmd.Meta.MessageId),
		Type: proto.Command_NORMAL,
	}

	c.toClient <- response
}
