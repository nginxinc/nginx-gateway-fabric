package validation

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

// Validators include validators for API resources from the perspective of a data-plane.
// It is used for fields that propagate into the data plane configuration. For example, the path in a routing rule.
// However, not all such fields are validated: NGF will not validate a field using Validators if it is confident that
// the field is valid.
type Validators struct {
	HTTPFieldsValidator HTTPFieldsValidator
	GenericValidator    GenericValidator
}

// HTTPFieldsValidator validates the HTTP-related fields of Gateway API resources from the perspective of
// a data-plane. Data-plane implementations must implement this interface.
//
//counterfeiter:generate . HTTPFieldsValidator
type HTTPFieldsValidator interface {
	ValidatePathInMatch(path string) error
	ValidateHeaderNameInMatch(name string) error
	ValidateHeaderValueInMatch(value string) error
	ValidateQueryParamNameInMatch(name string) error
	ValidateQueryParamValueInMatch(name string) error
	ValidateMethodInMatch(method string) (valid bool, supportedValues []string)
	ValidateRedirectScheme(scheme string) (valid bool, supportedValues []string)
	ValidateRedirectPort(port int32) error
	ValidateRedirectStatusCode(statusCode int) (valid bool, supportedValues []string)
	ValidateHostname(hostname string) error
	ValidateRewritePath(path string) error
	ValidateFilterHeaderName(name string) error
	ValidateFilterHeaderValue(value string) error
}

// GenericValidator validates any generic values from NGF API resources from the perspective of a data-plane.
// These could be values that we want to re-validate in case of any CRD schema manipulation.
//
//counterfeiter:generate . GenericValidator
type GenericValidator interface {
	ValidateEscapedStringNoVarExpansion(value string) error
	ValidateServiceName(name string) error
	ValidateNginxDuration(duration string) error
	ValidateEndpoint(endpoint string) error
}
