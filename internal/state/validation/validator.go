package validation

// Validators include validators for Gateway API resources from the perspective of a data-plane.
type Validators struct {
	HTTPFieldsValidator HTTPFieldsValidator
}

// HTTPFieldsValidator validates the HTTP-related fields of Gateway API resources from the perspective of
// a data-plane. Data-plane implementations must implement this interface.
//
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . HTTPFieldsValidator
type HTTPFieldsValidator interface {
	ValidateHostnameInServer(hostname string) error
	ValidatePathInPrefixMatch(path string) error
	ValidateHeaderNameInMatch(name string) error
	ValidateHeaderValueInMatch(value string) error
	ValidateQueryParamNameInMatch(name string) error
	ValidateQueryParamValueInMatch(name string) error
	ValidateMethodInMatch(method string) (valid bool, supportedValues []string)
	ValidateRedirectScheme(scheme string) (valid bool, supportedValues []string)
	ValidateRedirectHostname(hostname string) error
	ValidateRedirectPort(port int32) error
	ValidateRedirectStatusCode(statusCode int) (valid bool, supportedValues []string)
}
