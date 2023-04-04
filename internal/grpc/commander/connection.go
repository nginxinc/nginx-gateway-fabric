package commander

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/go-logr/logr"
	"github.com/gogo/protobuf/types"
	"github.com/nginx/agent/sdk/v2/checksum"
	"github.com/nginx/agent/sdk/v2/grpc"
	"github.com/nginx/agent/sdk/v2/proto"
	"golang.org/x/sync/errgroup"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/grpc/commander/exchanger"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/agent"
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

type configApplyStatus int

const (
	configApplyStatusSuccess configApplyStatus = iota
	configApplyStatusFailure
)

func (s configApplyStatus) String() string {
	switch s {
	case configApplyStatusSuccess:
		return "success"
	case configApplyStatusFailure:
		return "failure"
	default:
		return "unknown"
	}
}

type configApplyResponse struct {
	correlationID string
	message       string
	status        configApplyStatus
}

// configUpdatedChSize is the size of the channel that notifies the connection that the config has been updated.
// The length is 1 because we do not want to miss a notification while the connection is processing the last config.
const configUpdatedChSize = 1

// connection represents a connection to an agent.
type connection struct {
	cmdExchanger          exchanger.CommandExchanger
	observedConfig        observedNginxConfig
	configUpdatedCh       chan struct{}
	configApplyResponseCh chan configApplyResponse
	pendingConfig         *agent.NginxConfig
	logger                logr.Logger
	id                    string
	nginxID               string
	systemID              string
	state                 State
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
	configSubject observedNginxConfig,
) *connection {
	return &connection{
		logger:                logger,
		cmdExchanger:          cmdExchanger,
		observedConfig:        configSubject,
		configUpdatedCh:       make(chan struct{}, configUpdatedChSize),
		configApplyResponseCh: make(chan configApplyResponse),
		id:                    id,
	}
}

func (c *connection) ID() string {
	return c.id
}

func (c *connection) State() State {
	return c.state
}

func (c *connection) Update() {
	select {
	case c.configUpdatedCh <- struct{}{}:
		c.logger.Info("Queued config update")
	default:
	}
}

func createDownloadCommand(msgID, systemID, nginxID string) *proto.Command {
	return &proto.Command{
		Meta: &proto.Metadata{
			MessageId: msgID,
		},
		Type: proto.Command_DOWNLOAD,
		Data: &proto.Command_NginxConfig{
			NginxConfig: &proto.NginxConfig{
				Action: proto.NginxConfigAction_APPLY,
				ConfigData: &proto.ConfigDescriptor{
					SystemId: systemID,
					NginxId:  nginxID,
				},
			},
		},
	}
}

func (c *connection) sendConfig(request *proto.DownloadRequest, downloadServer proto.Commander_DownloadServer) error {
	config := c.pendingConfig

	if config.ID != request.GetMeta().GetMessageId() {
		err := fmt.Errorf(
			"pending config ID %q does not match request %q",
			config.ID,
			request.GetMeta().GetMessageId(),
		)
		c.logger.Error(err, "failed to send config")
		return err
	}

	cfg := &proto.NginxConfig{
		Action: proto.NginxConfigAction_APPLY,
		ConfigData: &proto.ConfigDescriptor{
			SystemId: c.systemID,
			NginxId:  c.nginxID,
		},
		Zconfig: config.Config,
		Zaux:    config.Aux,
		DirectoryMap: &proto.DirectoryMap{
			Directories: config.Directories,
		},
	}

	payload, err := json.Marshal(cfg)
	if err != nil {
		c.logger.Error(err, "failed to send config")
		return err
	}

	metadata := &proto.Metadata{
		Timestamp: types.TimestampNow(),
		MessageId: request.GetMeta().GetMessageId(),
	}

	payloadChecksum := checksum.Checksum(payload)
	chunks := checksum.Chunk(payload, 4*1024)

	err = downloadServer.Send(&proto.DataChunk{
		Chunk: &proto.DataChunk_Header{
			Header: &proto.ChunkedResourceHeader{
				Meta:      metadata,
				Chunks:    int32(len(chunks)),
				Checksum:  payloadChecksum,
				ChunkSize: 4 * 1024,
			},
		},
	})

	if err != nil {
		c.logger.Error(err, "failed to send config")
		return err
	}

	for id, chunk := range chunks {
		c.logger.Info("Sending data chunk", "chunk ID", id)
		err = downloadServer.Send(&proto.DataChunk{
			Chunk: &proto.DataChunk_Data{
				Data: &proto.ChunkedResourceChunk{
					ChunkId: int32(id),
					Data:    chunk,
					Meta:    metadata,
				},
			},
		})

		if err != nil {
			c.logger.Error(err, "failed to send chunk")
			return err
		}
	}

	c.logger.Info("Download finished")

	return nil
}

// run is a blocking method that kicks off the connection's receive loop and the CommandExchanger's Run loop.
// run will return when the context is canceled or if either loop returns an error.
func (c *connection) run(parent context.Context) error {
	defer func() {
		c.observedConfig.Remove(c)
	}()

	eg, ctx := errgroup.WithContext(parent)

	eg.Go(func() error {
		return c.receive(ctx)
	})

	eg.Go(func() error {
		return c.cmdExchanger.Run(ctx)
	})

	eg.Go(func() error {
		return c.updateConfigLoop(ctx)
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

func (c *connection) updateConfigLoop(ctx context.Context) error {
	defer func() {
		c.logger.Info("Stopping update config loop")
	}()
	c.logger.Info("Starting update config loop")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.configUpdatedCh:
			c.waitForConfigApply(ctx)
		}
	}
}

func (c *connection) waitForConfigApply(ctx context.Context) {
	config := c.observedConfig.GetLatestConfig()
	if config == nil {
		c.logger.Info("Latest config is nil, skipping update")
		return
	}

	c.pendingConfig = config

	c.logger.Info("Updating to latest config", "config generation", config.ID)

	status := c.statusForID(ctx, config.ID)

	select {
	case <-ctx.Done():
		c.logger.Error(ctx.Err(), "failed to update config")
		return
	case c.cmdExchanger.Out() <- createDownloadCommand(config.ID, c.nginxID, c.systemID):
	}

	now := time.Now()
	c.logger.Info("Waiting for config status", "config generation", config.ID)

	select {
	case <-ctx.Done():
		return
	case s := <-status:
		elapsedTime := time.Since(now)
		c.logger.Info(
			fmt.Sprintf("Config apply complete [%s]", s.status),
			"message",
			s.message,
			"config generation",
			config.ID,
			"duration",
			elapsedTime.String(),
		)
	}
}

// statusForID returns a channel that will receive a configApplyResponse when a final (
// not pending) status is received for the given config ID.
func (c *connection) statusForID(ctx context.Context, id string) <-chan configApplyResponse {
	statusForID := make(chan configApplyResponse)

	go func() {
		defer close(statusForID)

		for {
			select {
			case <-ctx.Done():
				return
			case status := <-c.configApplyResponseCh:
				// Not every status contains a correlation ID, so we only need to check it if it's not empty.
				// This is a workaround for some inconsistencies in the way the agent reports config apply statuses.
				if status.correlationID != "" && status.correlationID != id {
					c.logger.Info("Config status is for wrong generation",
						"actual config generation",
						status.correlationID,
						"expected config generation",
						id,
						"status",
						status.status,
						"message",
						status.message,
					)
					continue
				}

				select {
				case <-ctx.Done():
					return
				case statusForID <- status:
					return
				}
			}
		}
	}()

	return statusForID
}

func (c *connection) handleCommand(ctx context.Context, cmd *proto.Command) {
	switch cmd.Data.(type) {
	case *proto.Command_AgentConnectRequest:
		c.handleAgentConnectRequestCmd(ctx, cmd)
	case *proto.Command_DataplaneStatus:
		c.handleDataplaneStatus(ctx, cmd.GetDataplaneStatus())
	case *proto.Command_NginxConfigResponse:
		c.handleNginxConfigResponse(ctx, cmd.GetNginxConfigResponse())
	default:
		c.logger.Info("Ignoring command", "data type", fmt.Sprintf("%T", cmd.Data))
	}
}

func (c *connection) handleAgentConnectRequestCmd(ctx context.Context, cmd *proto.Command) {
	req := cmd.GetAgentConnectRequest()

	c.logger.Info("Received agent connect request")

	c.logger = c.logger.WithValues("podName", req.GetMeta().DisplayName)

	requestStatusCode := proto.AgentConnectStatus_CONNECT_OK
	msg := "Connected"

	if err := c.register(getFirstNginxID(req.GetDetails()), req.GetMeta().SystemUid); err != nil {
		requestStatusCode = proto.AgentConnectStatus_CONNECT_REJECTED_OTHER
		msg = err.Error()

		c.logger.Error(err, "failed to register agent")
	}

	res := createAgentConnectResponseCmd(cmd.GetMeta().GetMessageId(), requestStatusCode, msg)

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

	// trigger an update
	c.Update()
	// register for future config updates
	c.observedConfig.Register(c)

	return nil
}

// receiveFromUploadServer uploads data chunks from the UploadServer and logs them.
// FIXME(kate-osborn): NKG doesn't need this functionality and ideally we wouldn't have to implement and maintain this.
// Figure out how to remove this without causing errors in the agent.
func (c *connection) receiveFromUploadServer(server proto.Commander_UploadServer) error {
	c.logger.Info("Upload request")

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

		c.logger.Info("Received chunk from upload channel")

		if errors.Is(err, io.EOF) {
			c.logger.Info("Upload completed")
			return server.SendAndClose(&proto.UploadStatus{Status: proto.UploadStatus_OK})
		}
	}
}

func (c *connection) handleDataplaneStatus(ctx context.Context, status *proto.DataplaneStatus) {
	// Right now, we only care about AgentActivityStatuses that contain NginxConfigStatuses.
	if status.GetAgentActivityStatus() != nil {
		for _, activityStatus := range status.GetAgentActivityStatus() {
			if cfgStatus := activityStatus.GetNginxConfigStatus(); cfgStatus != nil {
				c.handleNginxConfigStatus(ctx, cfgStatus)
			}
		}
	}
}

func (c *connection) handleNginxConfigStatus(ctx context.Context, status *proto.NginxConfigStatus) {
	c.logger.Info("Received nginx config status", "status", status.Status, "message", status.Message)
	// If status is pending then we need to wait for the next status update
	if status.Status == proto.NginxConfigStatus_PENDING {
		return
	}

	applyStatus := configApplyStatusSuccess
	if status.Status == proto.NginxConfigStatus_ERROR {
		applyStatus = configApplyStatusFailure
	}

	res := configApplyResponse{
		correlationID: status.CorrelationId,
		status:        applyStatus,
		message:       status.Message,
	}

	c.sendConfigApplyResponse(ctx, res)
}

func (c *connection) handleNginxConfigResponse(ctx context.Context, res *proto.NginxConfigResponse) {
	status := res.Status

	c.logger.Info("Received nginx config response", "status", status.Status, "message", status.Message)

	// We only care about ERROR status because it indicates that the config apply action is complete.
	// An OK status can indicate that the config apply action is still in progress or that it is complete. However,
	// the Agent will send a DataplaneStatus update on a successful config apply, so we don't need to handle it here.
	// We handle the error case here, because in some cases, the Agent will not send a DataplaneStatus update on a
	// failed config apply.
	if status.Status != proto.CommandStatusResponse_CMD_ERROR {
		return
	}

	car := configApplyResponse{
		status:  configApplyStatusFailure,
		message: status.Error,
	}

	c.sendConfigApplyResponse(ctx, car)
}

func (c *connection) sendConfigApplyResponse(ctx context.Context, response configApplyResponse) {
	select {
	case <-ctx.Done():
		return
	case c.configApplyResponseCh <- response:
	default:
		// If there's no listener on c.configApplyResponseCh, then there's no pending config apply
		// and these status updates are extraneous.
		c.logger.Info(
			"Ignoring config apply response; no pending config apply",
			"config generation",
			response.correlationID,
		)
	}
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
