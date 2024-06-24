package stream

// Server holds all configuration for a stream server.
type Server struct {
	Listen     string
	ProxyPass  string
	Pass       string
	SSLPreread bool
}
