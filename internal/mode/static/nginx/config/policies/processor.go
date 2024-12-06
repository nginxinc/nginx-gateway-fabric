package policies

// UpstreamSettingsProcessor defines an interface for an UpstreamSettingsPolicy to implement the process function.
type UpstreamSettingsProcessor interface {
	Process(policies []Policy) UpstreamSettings
}

// UpstreamSettings contains settings from UpstreamSettingsPolicy.
type UpstreamSettings struct {
	ZoneSize             string
	KeepAliveTime        string
	KeepAliveTimeout     string
	KeepAliveConnections int32
	KeepAliveRequests    int32
}
