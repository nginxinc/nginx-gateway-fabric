package commander

import (
	"fmt"

	"github.com/nginx/agent/sdk/v2/proto"
)

// CreateAgentConnectRequestCmd creates an AgentConnectRequest Command for test purposes.
func CreateAgentConnectRequestCmd(messageID string) *proto.Command {
	return &proto.Command{
		Meta: &proto.Metadata{
			MessageId: messageID,
		},
		Data: &proto.Command_AgentConnectRequest{
			AgentConnectRequest: &proto.AgentConnectRequest{
				Meta: &proto.AgentMeta{
					SystemUid: fmt.Sprintf("system-%s", messageID),
				},
				Details: []*proto.NginxDetails{
					{
						NginxId: "nginxID",
					},
				},
			},
		},
	}
}
