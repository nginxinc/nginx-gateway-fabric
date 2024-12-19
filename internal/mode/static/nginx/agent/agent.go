package agent

import "fmt"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . NginxUpdater

// NginxUpdater is an interface for updating NGINX using the NGINX agent.
type NginxUpdater interface {
	UpdateNginxConfig(int)
	UpdateUpstreamServers()
}

// NginxUpdaterImpl implements the NginxUpdater interface.
type NginxUpdaterImpl struct {
	Plus bool
}

// UpdateNginxConfig sends the nginx configuration to the agent.
func (n *NginxUpdaterImpl) UpdateNginxConfig(files int) {
	fmt.Println("Sending nginx configuration to agent.", "numFiles", files)
}

// UpdateUpstreamServers sends an APIRequest to the agent to update upstream servers using the NGINX Plus API.
// Only applicable when using NGINX Plus.
func (n *NginxUpdaterImpl) UpdateUpstreamServers() {
	if !n.Plus {
		return
	}

	fmt.Println("Updating upstream servers using NGINX Plus API.")
}
