package validation

// HTTPRedirectValidator validates values for a redirect, which in NGINX is done with the return directive.
// For example, return 302 "https://example.com:8080";
type HTTPRedirectValidator struct{}

// HTTPRequestHeaderValidator validates values for request headers,
// which in NGINX is done with the proxy_set_header directive.
type HTTPRequestHeaderValidator struct{}

var supportedRedirectSchemes = map[string]struct{}{
	"http":  {},
	"https": {},
}

// ValidateRedirectScheme validates a scheme to be used in the return directive for a redirect.
// NGINX rules are not restrictive, but it is easier to validate just for two allowed values http and https,
// dictated by the Gateway API spec.
func (HTTPRedirectValidator) ValidateRedirectScheme(scheme string) (valid bool, supportedValues []string) {
	return validateInSupportedValues(scheme, supportedRedirectSchemes)
}

var redirectHostnameExamples = []string{"host", "example.com"}

func (HTTPRedirectValidator) ValidateRedirectHostname(hostname string) error {
	return validateEscapedStringNoVarExpansion(hostname, redirectHostnameExamples)
}

func (HTTPRedirectValidator) ValidateRedirectPort(_ int32) error {
	// any value is allowed
	return nil
}

var supportedRedirectStatusCodes = map[int]struct{}{
	301: {},
	302: {},
}

// ValidateRedirectStatusCode validates a status code to be used in the return directive for a redirect.
// NGINX allows 0..999. However, let's be conservative and only allow 301 and 302 (the values allowed by the Gateway API
// spec). Note that in the future, we might reserve some codes for internal redirects, so better not to allow all
// possible code values. We can always relax the validation later in case there is a need.
func (HTTPRedirectValidator) ValidateRedirectStatusCode(statusCode int) (valid bool, supportedValues []string) {
	return validateInSupportedValues(statusCode, supportedRedirectStatusCodes)
}

func (HTTPRequestHeaderValidator) ValidateRequestHeaderName(name string) error {
	return validateHeaderName(name)
}

var requestHeaderValueExamples = []string{"my-header-value", "example/12345=="}

func (HTTPRequestHeaderValidator) ValidateRequestHeaderValue(value string) error {
	// Variables in header values are supported by NGINX but not required by the Gateway API.
	return validateEscapedStringNoVarExpansion(value, requestHeaderValueExamples)
}
