package dataplane

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/resolver"
)

// PathType is the type of the path in a PathRule.
type PathType string

const (
	// PathTypePrefix indicates that the path is a prefix.
	PathTypePrefix PathType = "prefix"
	// PathTypeExact indicates that the path is exact.
	PathTypeExact PathType = "exact"
)

// Configuration is an intermediate representation of dataplane configuration.
type Configuration struct {
	// SSLKeyPairs holds all unique SSLKeyPairs.
	SSLKeyPairs map[SSLKeyPairID]SSLKeyPair
	// HTTPServers holds all HTTPServers.
	HTTPServers []VirtualServer
	// SSLServers holds all SSLServers.
	SSLServers []VirtualServer
	// Upstreams holds all unique Upstreams.
	Upstreams []Upstream
	// BackendGroups holds all unique BackendGroups.
	BackendGroups []BackendGroup
	// Tracing holds the Otel tracing configuration for the Dataplane.
	Tracing Tracing
	// Version represents the version of the generated configuration.
	Version int
}

// SSLKeyPairID is a unique identifier for a SSLKeyPair.
// The ID is safe to use as a file name.
type SSLKeyPairID string

// SSLKeyPair is an SSL private/public key pair.
type SSLKeyPair struct {
	// Cert is the certificate.
	Cert []byte
	// Key is the private key.
	Key []byte
}

// VirtualServer is a virtual server.
type VirtualServer struct {
	// SSL holds the SSL configuration for the server.
	SSL *SSL
	// Hostname is the hostname of the server.
	Hostname string
	// PathRules is a collection of routing rules.
	PathRules []PathRule
	// IsDefault indicates whether the server is the default server.
	IsDefault bool
	// Port is the port of the server.
	Port int32
}

// Upstream is a pool of endpoints to be load balanced.
type Upstream struct {
	// Name is the name of the Upstream. Will be unique for each service/port combination.
	Name string
	// ErrorMsg contains the error message if the Upstream is invalid.
	ErrorMsg string
	// Endpoints are the endpoints of the Upstream.
	Endpoints []resolver.Endpoint
}

// SSL is the SSL configuration for a server.
type SSL struct {
	// KeyPairID is the ID of the corresponding SSLKeyPair for the server.
	KeyPairID SSLKeyPairID
}

// PathRule represents routing rules that share a common path.
type PathRule struct {
	// Path is a path. For example, '/hello'.
	Path string
	// PathType is the type of the path.
	PathType PathType
	// MatchRules holds routing rules.
	MatchRules []MatchRule
}

// InvalidHTTPFilter is a special filter for handling the case when configured filters are invalid.
type InvalidHTTPFilter struct{}

// HTTPFilters hold the filters for a MatchRule.
type HTTPFilters struct {
	// InvalidFilter is a special filter that indicates whether the filters are invalid. If this is the case,
	// the data plane must return 500 error, and all other filters are nil.
	InvalidFilter *InvalidHTTPFilter
	// RequestRedirect holds the HTTPRequestRedirectFilter.
	RequestRedirect *HTTPRequestRedirectFilter
	// RequestHeaderModifiers holds the HTTPHeaderFilter.
	RequestHeaderModifiers *HTTPHeaderFilter
}

// HTTPHeader represents an HTTP header.
type HTTPHeader struct {
	// Name is the name of the header.
	Name string
	// Value is the value of the header.
	Value string
}

// HTTPHeaderFilter manipulates HTTP headers.
type HTTPHeaderFilter struct {
	// Set adds or replaces headers.
	Set []HTTPHeader
	// Add adds headers. It appends to any existing values associated with a header name.
	Add []HTTPHeader
	// Remove removes headers.
	Remove []string
}

// HTTPRequestRedirectFilter redirects HTTP requests.
type HTTPRequestRedirectFilter struct {
	// Scheme is the scheme of the redirect.
	Scheme *string
	// Hostname is the hostname of the redirect.
	Hostname *string
	// Port is the port of the redirect.
	Port *int32
	// StatusCode is the HTTP status code of the redirect.
	StatusCode *int
}

// HTTPHeaderMatch matches an HTTP header.
type HTTPHeaderMatch struct {
	// Name is the name of the header to match.
	Name string
	// Value is the value of the header to match.
	Value string
}

// HTTPQueryParamMatch matches an HTTP query parameter.
type HTTPQueryParamMatch struct {
	// Name is the name of the query parameter to match.
	Name string
	// Value is the value of the query parameter to match.
	Value string
}

// MatchRule represents a routing rule. It corresponds directly to a Match in the HTTPRoute resource.
// An HTTPRoute is guaranteed to have at least one rule with one match.
// If no rule or match is specified by the user, the default rule {{path:{ type: "PathPrefix", value: "/"}}}
// is set by the schema.
type MatchRule struct {
	// Filters holds the filters for the MatchRule.
	Filters HTTPFilters
	// Source is the ObjectMeta of the resource that includes the rule.
	Source *metav1.ObjectMeta
	// Match holds the match for the rule.
	Match Match
	// BackendGroup is the group of Backends that the rule routes to.
	BackendGroup BackendGroup
}

// Match represents a match for a routing rule which consist of matches against various HTTP request attributes.
type Match struct {
	// Method matches against the HTTP method.
	Method *string
	// Headers matches against the HTTP headers.
	Headers []HTTPHeaderMatch
	// QueryParams matches against the HTTP query parameters.
	QueryParams []HTTPQueryParamMatch
}

// BackendGroup represents a group of Backends for a routing rule in an HTTPRoute.
type BackendGroup struct {
	// Source is the NamespacedName of the HTTPRoute the group belongs to.
	Source types.NamespacedName
	// Backends is a list of Backends in the Group.
	Backends []Backend
	// RuleIdx is the index of the corresponding rule in the HTTPRoute.
	RuleIdx int
}

// Name returns the name of the backend group.
// This name must be unique across all HTTPRoutes and all rules within the same HTTPRoute.
// The RuleIdx is used to make the name unique across all rules within the same HTTPRoute.
// The RuleIdx may change for a given rule if an update is made to the HTTPRoute, but it will always match the index
// of the rule in the stored HTTPRoute.
func (bg *BackendGroup) Name() string {
	return fmt.Sprintf("%s__%s_rule%d", bg.Source.Namespace, bg.Source.Name, bg.RuleIdx)
}

// Backend represents a Backend for a routing rule.
type Backend struct {
	// UpstreamName is the name of the upstream for this backend.
	UpstreamName string
	// Weight is the weight of the BackendRef.
	// The possible values of weight are 0-1,000,000.
	// If weight is 0, no traffic should be forwarded for this entry.
	Weight int32
	// Valid indicates whether the Backend is valid.
	Valid bool
}

// Tracing represents Otel tracing configuration for the dataplane.
type Tracing struct {
	// ExporterEndpoint specifies the address of OTLP/gRPC endpoint that will accept telemetry data.
	ExporterEndpoint string
	// ServiceName is the “service.name” attribute of the OTel resource.
	ServiceName string
	// Interval specifies the tracing interval.
	Interval string
	// BatchSize specifies the maximum number of spans to be sent in one batch per worker.
	BatchSize int
	// BatchCount specifies the number of pending batches per worker, spans exceeding the limit are dropped.
	BatchCount int
	// Enabled specifies if tracing is enabled.
	Enabled bool
}
