package config

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/nginx/agent/sdk/v2/proto"
	"golang.org/x/sync/errgroup"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/agent"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/async"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/observer"
)

const (
	// configUpdatedChSize is 1 to allow an update to be queued while the current update is processing.
	// This guarantees that we will not miss updates.
	configUpdatedChSize = 1
	// applyTimeout is the time the Updater waits for the agent to send a status update after a config apply.
	applyTimeout = 1 * time.Minute
	// configApplyResponsesTimeout is the time the Updater waits to place a response on the configApplyResponses
	// channel.
	configApplyResponsesTimeout = 100 * time.Millisecond
)

// Updater is responsible for updating the NGINX configuration on the agent.
type Updater struct {
	// server is the bidirectional command channel server. It sends and receives commands to and from the agent.
	server proto.Commander_CommandChannelServer
	// configSubject pushes *NginxConfig to the Updater every time the *NginxConfig is updated.
	configSubject observer.Subject[*NginxConfig]
	// latestConfig synchronously stores the latest *NginxConfig received by the configSubject.
	latestConfig atomic.Value
	// configUpdateCompleted is the channel the Updater writes to when the latest config update is complete.
	// It signals that the agent is ready for another config.
	configUpdateCompleted chan struct{}
	// configUpdated is the channel that the configSubject writes to when the *NginxConfig is updated.
	configUpdated chan struct{}
	// configRequests is the channel that the commander writes to when the agent requests to Download the config.
	// The agent will reply to the request with the latest config.
	configRequests chan *Request
	// configApplyResponses is the channel that the Updater writes to when the server receives a config apply response.
	// It is used to verify whether a config apply was successful or not.
	configApplyResponses chan *applyResponse
	// connectInfo is the identifying information that the agent provides in its connect request.
	// This information is needed to create the *proto.NginxConfig payload.
	connectInfo agent.ConnectInfo
	// logger is the Updater's logger.
	logger logr.Logger
}

func NewUpdater(
	server proto.Commander_CommandChannelServer,
	info agent.ConnectInfo,
	configSubject observer.Subject[*NginxConfig],
	logger logr.Logger,
) *Updater {
	return &Updater{
		connectInfo:           info,
		server:                server,
		logger:                logger,
		configSubject:         configSubject,
		configUpdated:         make(chan struct{}, configUpdatedChSize),
		configRequests:        make(chan *Request),
		configApplyResponses:  make(chan *applyResponse),
		configUpdateCompleted: make(chan struct{}),
	}
}

// Start starts the Updater's loops and registers itself with the configSubject.
// It will block until the context is canceled, or an error occurs in one of the loops.
// On termination, it will remove itself from the configSubject.
func (u *Updater) Start(parent context.Context) error {
	u.logger.Info("Starting agent config updater")

	eg, ctx := errgroup.WithContext(parent)

	eg.Go(func() error {
		return u.receiveCommandLoop(ctx)
	})

	eg.Go(func() error {
		return u.updateConfigLoop(ctx)
	})

	eg.Go(func() error {
		return u.configRequestLoop(ctx)
	})

	u.configSubject.Register(u)

	defer func() {
		u.logger.Info("Stopping agent config updater")
		u.configSubject.Remove(u)
	}()

	return eg.Wait()
}

// Requests returns a write-only channel of *Request.
// Writing to this channel is equivalent to requesting the latest stored Nginx configuration.
func (u *Updater) Requests() chan<- *Request {
	return u.configRequests
}

// ID returns the ID of the updater. The observer.Subject calls this method.
func (u *Updater) ID() string {
	return u.connectInfo.ID
}

// Update is the method that the observer.Subject calls when the NginxConfig is updated.
// This method is required
func (u *Updater) Update(config observer.VersionedConfig) {
	select {
	case u.configUpdated <- struct{}{}:
	default:
	}

	u.latestConfig.Store(config)
}

// receiveCommandLoop receives commands from the server until an error occurs or the context is canceled.
func (u *Updater) receiveCommandLoop(ctx context.Context) error {
	for {
		cmd, err := u.server.Recv()
		if err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			u.handleCommand(ctx, cmd)
		}
	}
}

// updateConfigLoop receives from the configUpdated channel until an error occurs or the context is canceled.
// It sends a download command to the agent when it receives from the configUpdate channel and blocks until the
// update is complete or the context is canceled.
func (u *Updater) updateConfigLoop(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-u.configUpdated:
			if err := u.sendDownloadCommand(ctx); err != nil {
				return err
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-u.configUpdateCompleted:
			}
		}
	}
}

// configRequestLoop reads from the configRequest channel until an error occurs or the context is canceled.
// When it receives a Request, it calls handleRequest, which will block until the Request has been completed.
func (u *Updater) configRequestLoop(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case req := <-u.configRequests:
			if err := u.handleRequest(ctx, req); err != nil {
				return err
			}
		}
	}
}

// handleRequest replies to the Request with the latest NGINX configuration.
// It then waits until it receives a definitive response from the agent on the success of the config apply.
func (u *Updater) handleRequest(ctx context.Context, req *Request) error {
	reqID := req.id

	u.logger.Info("Handling request", "requestID", reqID)

	statusPromise := async.MonitorWithMatcher(
		ctx,
		applyTimeout,
		u.configApplyResponses,
		newApplyResponseMatcher(reqID),
	)

	config, ok := u.latestConfig.Load().(*NginxConfig)
	if !ok {
		panic(fmt.Sprintf("expected *NginxConfig, got %T", config))
	}

	u.logger.Info("Replying to request with config", "reqID", reqID, "version", config.version)

	if err := req.reply(ctx, u.toProtoConfig(config)); err != nil {
		return err
	}

	status, err := statusPromise.Wait()

	deadlineExceeded := errors.Is(err, context.DeadlineExceeded)
	if err != nil && !deadlineExceeded {
		return err
	}

	if deadlineExceeded {
		u.logger.Info("Timed out waiting for status", "reqID", reqID, "version", config.version)
	} else {
		u.logger.Info(
			fmt.Sprintf("Config apply complete [%s]", status.status),
			"message",
			status.message,
			"reqID",
			status.correlationID,
			"version",
			config.version,
		)
	}

	select {
	case <-ctx.Done():
		return nil
	case u.configUpdateCompleted <- struct{}{}:
		u.logger.Info("Config update complete")
	default:
	}

	return nil
}

// newApplyResponseMatcher returns an async.Matcher function that matches on applyResponses that are determinate (
// success/failure). It discards applyResponses that are indeterminate (
// pending/unknown status) or applyResponses that have a correlationID that does not match the one provided.
func newApplyResponseMatcher(correlationID string) func(res *applyResponse) bool {
	return func(res *applyResponse) bool {
		if res.correlationID != "" && res.correlationID != correlationID {
			return false
		}
		return res.status == applyStatusSuccess || res.status == applyStatusFailure
	}
}

// handleCommand handles commands sent by the agent.
// The updater is only concerned with commands that are config apply responses.
// This method attempts to convert commands to an applyResponse. If a command can't be converted,
// it is ignored. Otherwise, it attempts to place the command on the configApplyResponses channel.
// If the channel blocks for longer than the configApplyResponsesTimeout,
// it assumes there is no config apply in progress and ignores the command.
func (u *Updater) handleCommand(ctx context.Context, cmd *proto.Command) {
	if cmd == nil {
		// The agent should never send us a nil command, but we catch this case out of an abundance of caution.
		// We don't want to return an error in this case because that would break the CommandChannel
		// connection with the agent. Instead, we log the abnormality and continue processing.
		u.logger.Error(errors.New("received nil command"), "expected non-nil command")
		return
	}

	response := cmdToApplyStatus(cmd)
	if response == nil {
		u.logger.Info("Ignoring command", "type", fmt.Sprintf("%T", cmd.Data))
		return
	}

	u.logger.Info("Handling command", "command", cmd)

	select {
	case <-ctx.Done():
		return
	case u.configApplyResponses <- response:
	case <-time.After(configApplyResponsesTimeout):
		u.logger.Info(
			"Ignoring config apply response; no config apply in progress",
			"status",
			response.status,
			"correlationID",
			response.correlationID,
			"msg",
			response.message,
		)
	}
}

func (u *Updater) sendDownloadCommand(ctx context.Context) error {
	cmd := &proto.Command{
		Meta: &proto.Metadata{
			MessageId: uuid.NewString(),
		},
		Type: proto.Command_DOWNLOAD,
		Data: &proto.Command_NginxConfig{
			NginxConfig: &proto.NginxConfig{
				Action: proto.NginxConfigAction_APPLY,
				ConfigData: &proto.ConfigDescriptor{
					SystemId: u.connectInfo.SystemID,
					NginxId:  u.connectInfo.NginxID,
				},
			},
		},
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		u.logger.Info("Sending download command")
		return u.server.Send(cmd)
	}
}

func (u *Updater) toProtoConfig(config *NginxConfig) *proto.NginxConfig {
	return &proto.NginxConfig{
		Action: proto.NginxConfigAction_APPLY,
		ConfigData: &proto.ConfigDescriptor{
			SystemId: u.connectInfo.SystemID,
			NginxId:  u.connectInfo.NginxID,
		},
		Zconfig: config.config,
		Zaux:    config.aux,
		DirectoryMap: &proto.DirectoryMap{
			Directories: config.directories,
		},
	}
}
