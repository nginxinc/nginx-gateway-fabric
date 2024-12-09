package policies

// UpstreamSettingsProcessor defines an interface for an UpstreamSettingsPolicy processor
// to implement the process function.
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
