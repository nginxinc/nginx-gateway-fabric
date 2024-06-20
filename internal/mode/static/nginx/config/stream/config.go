package stream

// Server holds all configuration for a stream server.
type Server struct {
	Listen      string
	Destination string
	SSLPreread  bool
	ProxyPass   bool
}
