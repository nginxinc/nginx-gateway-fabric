package http

// Server holds all configuration for an HTTP server.
type Server struct {
	SSL           *SSL
	ServerName    string
	Listen        string
	Locations     []Location
	Includes      []string
	IsDefaultHTTP bool
	IsDefaultSSL  bool
	GRPC          bool
}

// Location holds all configuration for an HTTP location.
type Location struct {
	Path            string
	ProxyPass       string
	HTTPMatchKey    string
	HTTPMatchVar    string
	ProxySetHeaders []Header
	ProxySSLVerify  *ProxySSLVerify
	Return          *Return
	ResponseHeaders ResponseHeaders
	Rewrites        []string
	Includes        []string
	GRPC            bool
}

// Header defines an HTTP header to be passed to the proxied server.
type Header struct {
	Name  string
	Value string
}

// ResponseHeaders holds all response headers to be added, set, or removed.
type ResponseHeaders struct {
	Add    []Header
	Set    []Header
	Remove []string
}

// Return represents an HTTP return.
type Return struct {
	Body string
	Code StatusCode
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
	// StatusInternalServerError is the HTTP 500 status code.
	StatusInternalServerError StatusCode = 500
)

// Upstream holds all configuration for an HTTP upstream.
type Upstream struct {
	Name     string
	ZoneSize string // format: 512k, 1m
	Servers  []UpstreamServer
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

// ProxySSLVerify holds the proxied HTTPS server verification configuration.
type ProxySSLVerify struct {
	TrustedCertificate string
	Name               string
}
