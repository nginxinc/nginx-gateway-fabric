package commander

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/go-logr/logr"
	"github.com/gogo/protobuf/types"
	"github.com/nginx/agent/sdk/v2/checksum"
	"github.com/nginx/agent/sdk/v2/grpc"
	"github.com/nginx/agent/sdk/v2/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/agent"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/agent/config"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/observer"
)

// nolint:lll
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 github.com/nginx/agent/sdk/v2/proto.Commander_CommandChannelServer
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 github.com/nginx/agent/sdk/v2/proto.Commander_UploadServer

const (
	serverUUIDKey = "uuid"
	connectedMsg  = "CONNECTED"
)

type configUpdater interface {
	Start(ctx context.Context) error
	Requests() chan<- *config.Request
}

type agentStore interface {
	Add(info agent.ConnectInfo)
	Delete(id string)
	Get(id string) (agent.ConnectInfo, bool)
}

// Commander implements the proto.CommanderServer interface.
type Commander struct {
	subject observer.Subject[*config.NginxConfig]
	store   agentStore
	logger  logr.Logger

	updaters     map[string]configUpdater
	updatersLock sync.Mutex
}

// NewCommander returns a new instance of the Commander.
func NewCommander(
	store agentStore,
	subject observer.Subject[*config.NginxConfig],
	logger logr.Logger,
) *Commander {
	return &Commander{
		logger:   logger,
		subject:  subject,
		updaters: make(map[string]configUpdater),
		store:    store,
	}
}

func (c *Commander) CommandChannel(server proto.Commander_CommandChannelServer) error {
	id, err := getUUIDFromContext(server.Context())
	if err != nil {
		c.logger.Error(err, "cannot get the UUID of the agent")
		return err
	}

	if err = c.startConfigUpdater(server, id); err != nil {
		if isContextCanceled(err) {
			return nil
		}

		c.logger.Error(err, "error starting config updater for agent", "agentID", id)
		return err
	}

	return nil
}

func (c *Commander) startConfigUpdater(server proto.Commander_CommandChannelServer, agentID string) error {
	info, ok := c.store.Get(agentID)

	if !ok {
		var err error
		info, err = c.waitForAgentConnect(server, agentID)
		if err != nil {
			return err
		}

		c.store.Add(info)
	}

	updater := config.NewUpdater(
		server,
		info,
		c.subject,
		c.logger.WithName("configUpdater").WithValues("agentID", agentID, "podName", info.PodName),
	)

	c.addUpdater(agentID, updater)

	defer func() {
		c.store.Delete(agentID)
		c.deleteUpdater(agentID)
		c.logger.Info("CommandChannel closed", "agentID", agentID)
	}()

	c.logger.Info("CommandChannel established", "agentID", agentID)

	return updater.Start(server.Context())
}

// Download implements the Download method of the Commander gRPC service.
// An agent uses this method to download the NGINX configuration.
func (c *Commander) Download(request *proto.DownloadRequest, server proto.Commander_DownloadServer) error {
	id, err := getUUIDFromContext(server.Context())
	if err != nil {
		c.logger.Error(err, "failed download; cannot get the UUID of the agent")
		return err
	}

	c.logger.Info("Download Request", "agentID", id)

	if err = c.download(request, server, id); err != nil {
		if isContextCanceled(err) {
			c.logger.Info("Download not completed; context canceled", "agentID", id)
			return nil
		}

		c.logger.Error(err, "failed download", "agentID", id)
		return err
	}

	return nil
}

func (c *Commander) download(request *proto.DownloadRequest, server proto.Commander_DownloadServer, id string) error {
	updater := c.getUpdater(id)

	if updater == nil {
		return fmt.Errorf("no config updater registered for agent with ID %q", id)
	}

	msgID := getMessageIDFromMeta(request.GetMeta())
	req := config.NewRequest(msgID)
	ctx := server.Context()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case updater.Requests() <- req:
		c.logger.Info("Sent config update request", "req ID", msgID)
	}

	cfg, err := req.WaitForReply(ctx)
	if err != nil {
		return err
	}

	return sendConfigToDownloadServer(cfg, server, msgID)
}

// Upload implements the Upload method of the Commander gRPC service.
// FIXME(kate-osborn): NKG doesn't need this functionality and ideally we wouldn't have to implement and maintain this.
// Figure out how to remove this without causing errors in the agent.
func (c *Commander) Upload(server proto.Commander_UploadServer) error {
	c.logger.Info("Commander Upload requested")

	for {
		// Recv blocks until it receives a message into or the stream is
		// done. It returns io.EOF when the client has performed a CloseSend. On
		// any non-EOF error, the stream is aborted and the error contains the
		// RPC status.
		_, err := server.Recv()

		if err != nil && !errors.Is(err, io.EOF) {
			c.logger.Error(err, "upload receive error")
			return err
		}

		if errors.Is(err, io.EOF) {
			c.logger.Info("Upload completed")
			return server.SendAndClose(&proto.UploadStatus{Status: proto.UploadStatus_OK})
		}
	}
}

func (c *Commander) waitForAgentConnect(
	server proto.Commander_CommandChannelServer,
	id string,
) (agent.ConnectInfo, error) {
	c.logger.Info("Waiting for agent to send connect request", "agentID", id)

	cmd, err := c.waitForConnectRequestCmd(server)
	if err != nil {
		return agent.ConnectInfo{}, fmt.Errorf("error waiting for connect request: %w", err)
	}

	connectInfo := agent.NewConnectInfo(id, cmd.GetAgentConnectRequest())

	code := proto.AgentConnectStatus_CONNECT_OK
	msg := connectedMsg

	validateErr := connectInfo.Validate()
	if validateErr != nil {
		code = proto.AgentConnectStatus_CONNECT_REJECTED_OTHER
		msg = err.Error()
	}

	err = c.sendConnectResponse(server, id, connectResponse(getMessageIDFromMeta(cmd.GetMeta()), code, msg))

	return connectInfo, errors.Join(validateErr, err)
}

func (c *Commander) sendConnectResponse(
	server proto.Commander_CommandChannelServer,
	id string,
	cmd *proto.Command,
) error {
	ctx := server.Context()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		// We don't return an error here, because it isn't necessary for the agent to receive this response.
		// It's just for debugging purposes.
		if err := server.Send(cmd); err != nil {
			c.logger.Error(err, "failed to send connect response", "agentID", id)
		}
	}

	return nil
}

func (c *Commander) waitForConnectRequestCmd(server proto.Commander_CommandChannelServer) (*proto.Command, error) {
	ctx := server.Context()

	for {
		cmd, err := server.Recv()
		if err != nil {
			return nil, err
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			if cmd != nil {
				if req := cmd.GetAgentConnectRequest(); req != nil {
					return cmd, nil
				}
				c.logger.Info("Ignoring command", "command data type", fmt.Sprintf("%T", cmd.Data))
			} else {
				// The agent should never send us a nil command, but we catch this case out of an abundance of caution.
				// We don't want to return an error in this case because that would break the CommandChannel
				// connection with the agent. Instead, we log the abnormality and continue processing.
				c.logger.Error(errors.New("received nil command"), "expected non-nil command")
			}
		}
	}
}

func (c *Commander) deleteUpdater(agentID string) {
	c.updatersLock.Lock()
	defer c.updatersLock.Unlock()

	delete(c.updaters, agentID)
}

func (c *Commander) addUpdater(agentID string, updater *config.Updater) {
	c.updatersLock.Lock()
	defer c.updatersLock.Unlock()

	c.updaters[agentID] = updater
}

func (c *Commander) getUpdater(id string) configUpdater {
	c.updatersLock.Lock()
	defer c.updatersLock.Unlock()

	return c.updaters[id]
}

func sendConfigToDownloadServer(cfg *proto.NginxConfig, server proto.Commander_DownloadServer, msgID string) error {
	payload, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	payloadChecksum := checksum.Checksum(payload)
	chunks := checksum.Chunk(payload, 4*1024)

	meta := &proto.Metadata{
		Timestamp: types.TimestampNow(),
		MessageId: msgID,
	}

	err = server.Send(&proto.DataChunk{
		Chunk: &proto.DataChunk_Header{
			Header: &proto.ChunkedResourceHeader{
				Meta:      meta,
				Chunks:    int32(len(chunks)),
				Checksum:  payloadChecksum,
				ChunkSize: 4 * 1024,
			},
		},
	})

	if err != nil {
		return err
	}

	for id, chunk := range chunks {
		err = server.Send(&proto.DataChunk{
			Chunk: &proto.DataChunk_Data{
				Data: &proto.ChunkedResourceChunk{
					ChunkId: int32(id),
					Data:    chunk,
					Meta:    meta,
				},
			},
		})

		if err != nil {
			return err
		}
	}

	return nil
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

func connectResponse(msgID string, statusCode proto.AgentConnectStatus_StatusCode, msg string) *proto.Command {
	return &proto.Command{
		Data: &proto.Command_AgentConnectResponse{
			AgentConnectResponse: &proto.AgentConnectResponse{
				Status: &proto.AgentConnectStatus{
					StatusCode: statusCode,
					Message:    msg,
				},
			},
		},
		Meta: grpc.NewMessageMeta(msgID),
		Type: proto.Command_NORMAL,
	}
}

func getMessageIDFromMeta(meta *proto.Metadata) string {
	if meta != nil {
		return meta.GetMessageId()
	}

	return ""
}

func isContextCanceled(err error) bool {
	if errors.Is(err, context.Canceled) {
		return true
	}

	if st, ok := status.FromError(err); ok {
		return st.Code() == codes.Canceled
	}

	return false
}
