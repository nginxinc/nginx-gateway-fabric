package http

// Server holds all configuration for an HTTP server.
type Server struct {
	SSL           *SSL
	ServerName    string
	Locations     []Location
	IsDefaultHTTP bool
	IsDefaultSSL  bool
}

// Location holds all configuration for an HTTP location.
type Location struct {
	Return       *Return
	Path         string
	ProxyPass    string
	HTTPMatchVar string
	Internal     bool
}

// Return represents an HTTP return.
type Return struct {
	Code StatusCode
	URL  string
}

// SSL holds all SSL related configuration.
type SSL struct {
	Certificate    string
	CertificateKey string
}

// StatusCode is an HTTP status code.
type StatusCode int

const (
	// StatusFound is the HTTP 302 status code.
	StatusFound StatusCode = 302
	// StatusNotFound is the HTTP 404 status code.
	StatusNotFound StatusCode = 404
)

// Upstream holds all configuration for an HTTP upstream.
type Upstream struct {
	Name    string
	Servers []UpstreamServer
}

// UpstreamServer holds all configuration for an HTTP upstream server.
type UpstreamServer struct {
	Address string
}

// SplitClient holds all configuration for an HTTP split client.
type SplitClient struct {
	VariableName  string
	Distributions []SplitClientDistribution
}

// SplitClientDistribution maps Percentage to Value in a SplitClient.
type SplitClientDistribution struct {
	Percent string
	Value   string
}
