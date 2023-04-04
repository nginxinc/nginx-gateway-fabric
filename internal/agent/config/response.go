package config

import (
	"strings"

	"github.com/nginx/agent/sdk/v2/proto"
)

// applyResponse represents a response to a config apply action on the agent.
type applyResponse struct {
	correlationID string
	message       string
	status        applyStatus
}

// applyStatus is the status of the config apply action on the agent
type applyStatus string

const (
	applyStatusSuccess applyStatus = "success"
	applyStatusFailure applyStatus = "failure"
	applyStatusPending applyStatus = "pending"
	applyStatusUnknown applyStatus = "unknown"
)

// FIXME(kate-osborn): We should only need to check for dataplane status.
// There's a bug in the agent where sometimes a dataplane status command is not sent when the config apply fails.
func cmdToApplyStatus(cmd *proto.Command) *applyResponse {
	switch cmd.Data.(type) {
	case *proto.Command_NginxConfigResponse:
		return convertNginxConfigResponse(cmd.GetNginxConfigResponse())
	case *proto.Command_DataplaneStatus:
		return convertDataplaneStatus(cmd.GetDataplaneStatus())
	default:
		return nil
	}
}

func convertDataplaneStatus(status *proto.DataplaneStatus) *applyResponse {
	activityStatuses := status.GetAgentActivityStatus()

	if activityStatuses == nil {
		return nil
	}

	s := getFirstNginxConfigStatus(activityStatuses)
	if s == nil {
		return nil
	}

	return &applyResponse{
		correlationID: s.CorrelationId,
		message:       s.Message,
		status:        convertNginxConfigStatus(s.Status),
	}
}

func convertNginxConfigResponse(res *proto.NginxConfigResponse) *applyResponse {
	status := res.Status

	if status == nil {
		return nil
	}

	if strings.Contains(status.Message, "upload") {
		// ignore upload config responses
		return nil
	}

	return &applyResponse{
		message: status.Message,
		status:  convertCommandStatusResponse(status),
	}
}

func convertNginxConfigStatus(status proto.NginxConfigStatus_Status) applyStatus {
	switch status {
	case proto.NginxConfigStatus_OK:
		return applyStatusSuccess
	case proto.NginxConfigStatus_PENDING:
		return applyStatusPending
	case proto.NginxConfigStatus_ERROR:
		return applyStatusFailure
	default:
		return applyStatusUnknown
	}
}

func convertCommandStatusResponse(status *proto.CommandStatusResponse) applyStatus {
	if status.Status == proto.CommandStatusResponse_CMD_OK {
		if strings.Contains(status.Message, "config applied successfully") {
			return applyStatusSuccess
		}

		return applyStatusPending
	}

	if status.Status == proto.CommandStatusResponse_CMD_ERROR {
		return applyStatusFailure
	}

	return applyStatusUnknown
}

// TODO: figure out if it's possible to have multiple NginxConfigStatus in a single DataplaneStatus.
func getFirstNginxConfigStatus(status []*proto.AgentActivityStatus) *proto.NginxConfigStatus {
	for _, s := range status {
		if s.GetNginxConfigStatus() != nil {
			return s.GetNginxConfigStatus()
		}
	}
	return nil
}
