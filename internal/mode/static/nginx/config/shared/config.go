package shared

// Map defines an NGINX map.
type Map struct {
	Source       string
	Variable     string
	Parameters   []MapParameter
	UseHostnames bool
}

// MapParameter defines a Value and Result pair in a Map.
type MapParameter struct {
	Value  string
	Result string
}

// IPFamily holds the IP family configuration to be used by NGINX.
type IPFamily struct {
	IPv4 bool
	IPv6 bool
}

// RewriteClientIP holds the configuration for the rewrite client IP settings.
type RewriteClientIPSettings struct {
	RealIPHeader  string
	RealIPFrom    []string
	Recursive     bool
	ProxyProtocol bool
}
