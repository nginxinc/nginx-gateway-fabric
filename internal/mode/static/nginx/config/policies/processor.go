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
	// KeepAlive contains the keepalive settings.
	KeepAlive KeepAlive
}

// KeepAlive contains the keepalive settings.
type KeepAlive struct {
	// Time is the keepalive time value.
	Time string
	// Timeout is the keepalive timeout value.
	Timeout string
	// Connections is the keepalive connections value.
	Connections int32
	// Requests is the keepalive requests value.
	Requests int32
}
