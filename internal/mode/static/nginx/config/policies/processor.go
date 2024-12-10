package policies

// UpstreamSettingsProcessor is an interface that defines a method for processing a list of policies
// into upstream settings.
type UpstreamSettingsProcessor interface {
	Process(policies []Policy) UpstreamSettings
}

// UpstreamSettings contains settings from UpstreamSettingsPolicy.
type UpstreamSettings struct {
	// ZoneSize is the zone size setting.
	ZoneSize string
	// KeepAliveTime is the keep-alive time setting.
	KeepAliveTime string
	// KeepAliveTimeout is the keep-alive timeout setting.
	KeepAliveTimeout string
	// KeepAliveConnections is the keep-alive connections setting.
	KeepAliveConnections int32
	// KeepAliveRequests is the keep-alive requests setting.
	KeepAliveRequests int32
}
