package agent

import (
	"fmt"

	"github.com/nginx/agent/sdk/v2/proto"
)

// ConnectInfo is the identifying information that the agent sends in its connect request.
type ConnectInfo struct {
	ID       string
	NginxID  string
	SystemID string
	PodName  string
}

func NewConnectInfo(id string, req *proto.AgentConnectRequest) ConnectInfo {
	details := ConnectInfo{
		ID:      id,
		NginxID: getFirstNginxID(req.GetDetails()),
	}

	if meta := req.GetMeta(); meta != nil {
		details.SystemID = meta.GetSystemUid()
		details.PodName = meta.GetDisplayName()
	}

	return details
}

func (d ConnectInfo) Validate() error {
	if d.NginxID == "" || d.SystemID == "" {
		return fmt.Errorf("missing NginxID: '%s' and/or SystemID: '%s'", d.NginxID, d.SystemID)
	}

	return nil
}

func getFirstNginxID(details []*proto.NginxDetails) (id string) {
	if len(details) > 0 {
		id = details[0].GetNginxId()
	}

	return
}
