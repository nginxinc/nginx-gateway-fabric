package upstreamsettings

import (
	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/config/policies"
)

// Processor processes UpstreamSettingsPolicies. It implements policies.UpstreamSettingsProcessor.
type Processor struct{}

// NewProcessor returns a new Processor.
func NewProcessor() Processor {
	return Processor{}
}

// Process processes policies into an UpstreamSettings object. The policies are already validated and are guaranteed
// to not contain overlapping settings. This method merges all fields in the policies into a single UpstreamSettings
// object.
func (g Processor) Process(pols []policies.Policy) policies.UpstreamSettings {
	return processPolicies(pols)
}

func processPolicies(pols []policies.Policy) policies.UpstreamSettings {
	upstreamSettings := policies.UpstreamSettings{}

	for _, pol := range pols {
		usp, ok := pol.(*ngfAPI.UpstreamSettingsPolicy)
		if !ok {
			continue
		}

		// we can assume that there will be no instance of two or more policies setting the same
		// field for the same service
		if usp.Spec.ZoneSize != nil {
			upstreamSettings.ZoneSize = string(*usp.Spec.ZoneSize)
		}

		if usp.Spec.KeepAlive != nil {
			if usp.Spec.KeepAlive.Connections != nil {
				upstreamSettings.KeepAlive.Connections = *usp.Spec.KeepAlive.Connections
			}

			if usp.Spec.KeepAlive.Requests != nil {
				upstreamSettings.KeepAlive.Requests = *usp.Spec.KeepAlive.Requests
			}

			if usp.Spec.KeepAlive.Time != nil {
				upstreamSettings.KeepAlive.Time = string(*usp.Spec.KeepAlive.Time)
			}

			if usp.Spec.KeepAlive.Timeout != nil {
				upstreamSettings.KeepAlive.Timeout = string(*usp.Spec.KeepAlive.Timeout)
			}
		}
	}

	return upstreamSettings
}
