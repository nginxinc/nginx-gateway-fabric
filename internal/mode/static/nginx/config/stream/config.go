package stream

// Server holds all configuration for a stream server.
type Server struct {
	Listen     string
	ProxyPass  string
	Pass       string
	SSLPreread bool
}

// Upstream holds all configuration for a stream upstream.
type Upstream struct {
	Name     string
	ZoneSize string // format: 512k, 1m
	Servers  []UpstreamServer
}

// UpstreamServer holds all configuration for a stream upstream server.
type UpstreamServer struct {
	Address string
}
