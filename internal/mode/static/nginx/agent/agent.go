package agent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	pb "github.com/nginx/agent/v3/api/grpc/mpi/v1"
	"google.golang.org/protobuf/types/known/structpb"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/agent/broadcast"
	agentgrpc "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/agent/grpc"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/dataplane"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/resolver"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/status"
)

const retryUpstreamTimeout = 5 * time.Second

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . NginxUpdater

// NginxUpdater is an interface for updating NGINX using the NGINX agent.
type NginxUpdater interface {
	UpdateConfig(deployment *Deployment, files []File) bool
	UpdateUpstreamServers(deployment *Deployment, conf dataplane.Configuration) bool
}

// NginxUpdaterImpl implements the NginxUpdater interface.
type NginxUpdaterImpl struct {
	CommandService   *commandService
	FileService      *fileService
	NginxDeployments *DeploymentStore
	logger           logr.Logger
	plus             bool
	retryTimeout     time.Duration
}

// NewNginxUpdater returns a new NginxUpdaterImpl instance.
func NewNginxUpdater(
	logger logr.Logger,
	reader client.Reader,
	statusQueue *status.Queue,
	plus bool,
) *NginxUpdaterImpl {
	connTracker := agentgrpc.NewConnectionsTracker()
	nginxDeployments := NewDeploymentStore(connTracker)

	commandService := newCommandService(
		logger.WithName("commandService"),
		reader,
		nginxDeployments,
		connTracker,
		statusQueue,
	)
	fileService := newFileService(logger.WithName("fileService"), nginxDeployments, connTracker)

	return &NginxUpdaterImpl{
		logger:           logger,
		plus:             plus,
		NginxDeployments: nginxDeployments,
		CommandService:   commandService,
		FileService:      fileService,
		retryTimeout:     retryUpstreamTimeout,
	}
}

// UpdateConfig sends the nginx configuration to the agent.
// Returns whether the configuration was sent to any agents.
//
// The flow of events is as follows:
// - Set the configuration files on the deployment.
// - Broadcast the message containing file metadata to all pods (subscriptions) for the deployment.
// - Agent receives a ConfigApplyRequest with the list of file metadata.
// - Agent calls GetFile for each file in the list, which we send back to the agent.
// - Agent updates nginx, and responds with a DataPlaneResponse.
// - Subscriber responds back to the broadcaster to inform that the transaction is complete.
// - If any errors occurred, they are set on the deployment for the handler to use in the status update.
func (n *NginxUpdaterImpl) UpdateConfig(
	deployment *Deployment,
	files []File,
) bool {
	n.logger.Info("Sending nginx configuration to agent")

	msg := deployment.SetFiles(files)
	applied := deployment.GetBroadcaster().Send(msg)

	deployment.SetLatestConfigError(deployment.GetConfigurationStatus())

	return applied
}

// UpdateUpstreamServers sends an APIRequest to the agent to update upstream servers using the NGINX Plus API.
// Only applicable when using NGINX Plus.
// Returns whether the configuration was sent to any agents.
func (n *NginxUpdaterImpl) UpdateUpstreamServers(
	deployment *Deployment,
	conf dataplane.Configuration,
) bool {
	if !n.plus {
		return false
	}

	broadcaster := deployment.GetBroadcaster()

	// reset the latest error to nil now that we're applying new config
	deployment.SetLatestUpstreamError(nil)

	// TODO(sberman): optimize this by only sending updates that are necessary.
	// Call GetUpstreams first (will need Subscribers to send responses back), and
	// then determine which upstreams actually need to be updated.

	var errs []error
	var applied bool
	actions := make([]*pb.NGINXPlusAction, 0, len(conf.Upstreams)+len(conf.StreamUpstreams))
	for _, upstream := range conf.Upstreams {
		action := &pb.NGINXPlusAction{
			Action: &pb.NGINXPlusAction_UpdateHttpUpstreamServers{
				UpdateHttpUpstreamServers: buildHTTPUpstreamServers(upstream),
			},
		}
		actions = append(actions, action)
	}

	for _, upstream := range conf.StreamUpstreams {
		action := &pb.NGINXPlusAction{
			Action: &pb.NGINXPlusAction_UpdateStreamServers{
				UpdateStreamServers: buildStreamUpstreamServers(upstream),
			},
		}
		actions = append(actions, action)
	}

	for _, action := range actions {
		msg := broadcast.NginxAgentMessage{
			Type:            broadcast.APIRequest,
			NGINXPlusAction: action,
		}

		requestApplied, err := n.sendRequest(broadcaster, msg, deployment)
		if err != nil {
			errs = append(errs, fmt.Errorf(
				"couldn't update upstream via the API: %w", deployment.GetConfigurationStatus()))
		}
		applied = applied || requestApplied
	}

	if len(errs) != 0 {
		deployment.SetLatestUpstreamError(errors.Join(errs...))
	} else if applied {
		n.logger.Info("Updated upstream servers using NGINX Plus API")
	}

	// Store the most recent actions on the deployment so any new subscribers can apply them when first connecting.
	deployment.SetNGINXPlusActions(actions)

	return applied
}

func buildHTTPUpstreamServers(upstream dataplane.Upstream) *pb.UpdateHTTPUpstreamServers {
	return &pb.UpdateHTTPUpstreamServers{
		HttpUpstreamName: upstream.Name,
		Servers:          buildUpstreamServers(upstream),
	}
}

func buildStreamUpstreamServers(upstream dataplane.Upstream) *pb.UpdateStreamServers {
	return &pb.UpdateStreamServers{
		UpstreamStreamName: upstream.Name,
		Servers:            buildUpstreamServers(upstream),
	}
}

func buildUpstreamServers(upstream dataplane.Upstream) []*structpb.Struct {
	servers := make([]*structpb.Struct, 0, len(upstream.Endpoints))

	for _, endpoint := range upstream.Endpoints {
		port, format := getPortAndIPFormat(endpoint)
		value := fmt.Sprintf(format, endpoint.Address, port)

		server := &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"server": structpb.NewStringValue(value),
			},
		}

		servers = append(servers, server)
	}

	return servers
}

func (n *NginxUpdaterImpl) sendRequest(
	broadcaster broadcast.Broadcaster,
	msg broadcast.NginxAgentMessage,
	deployment *Deployment,
) (bool, error) {
	// retry the API update request because sometimes nginx isn't quite ready after the config apply reload
	ctx, cancel := context.WithTimeout(context.Background(), n.retryTimeout)
	defer cancel()

	var applied bool
	if err := wait.PollUntilContextCancel(
		ctx,
		500*time.Millisecond,
		true, // poll immediately
		func(_ context.Context) (bool, error) {
			applied = broadcaster.Send(msg)
			if statusErr := deployment.GetConfigurationStatus(); statusErr != nil {
				return false, nil //nolint:nilerr // will get error once done polling
			}

			return true, nil
		},
	); err != nil {
		return applied, err
	}

	return applied, nil
}

func getPortAndIPFormat(ep resolver.Endpoint) (string, string) {
	var port string

	if ep.Port != 0 {
		port = fmt.Sprintf(":%d", ep.Port)
	}

	format := "%s%s"
	if ep.IPv6 {
		format = "[%s]%s"
	}

	return port, format
}
