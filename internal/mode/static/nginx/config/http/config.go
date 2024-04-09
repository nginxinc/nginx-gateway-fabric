package http

// Server holds all configuration for an HTTP server.
type Server struct {
	SSL           *SSL
	ServerName    string
	Locations     []Location
	IsDefaultHTTP bool
	IsDefaultSSL  bool
	Port          int32
}

// Location holds all configuration for an HTTP location.
type Location struct {
	Path            string
	ProxyPass       string
	HTTPMatchKey    string
	ProxySetHeaders []Header
	ProxySSLVerify  *ProxySSLVerify
	Return          *Return
	Rewrites        []string
}

// Header defines a HTTP header to be passed to the proxied server.
type Header struct {
	Name  string
	Value string
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

// Map defines an NGINX map.
type Map struct {
	Source     string
	Variable   string
	Parameters []MapParameter
}

// Parameter defines a Value and Result pair in a Map.
type MapParameter struct {
	Value  string
	Result string
}

// ProxySSLVerify holds the proxied HTTPS server verification configuration.
type ProxySSLVerify struct {
	TrustedCertificate string
	Name               string
}

// httpMatch is an internal representation of an HTTPRouteMatch.
// This struct is marshaled into a string and stored as a variable in the nginx location block for the route's path.
// The NJS httpmatches module will look up this variable on the request object and compare the request against the
// Method, Headers, and QueryParams contained in httpMatch.
// If the request satisfies the httpMatch, NGINX will redirect the request to the location RedirectPath.
type RouteMatch struct {
	// Method is the HTTPMethod of the HTTPRouteMatch.
	Method string `json:"method,omitempty"`
	// RedirectPath is the path to redirect the request to if the request satisfies the match conditions.
	RedirectPath string `json:"redirectPath,omitempty"`
	// Headers is a list of HTTPHeaders name value pairs with the format "{name}:{value}".
	Headers []string `json:"headers,omitempty"`
	// QueryParams is a list of HTTPQueryParams name value pairs with the format "{name}={value}".
	QueryParams []string `json:"params,omitempty"`
	// Any represents a match with no match conditions.
	Any bool `json:"any,omitempty"`
}
