package validation

import (
	"errors"
	"regexp"

	k8svalidation "k8s.io/apimachinery/pkg/util/validation"
)

// GenericValidator validates values for generic cases in the nginx conf.
type GenericValidator struct{}

// ValidateEscapedStringNoVarExpansion ensures that no invalid characters are included in the string value that
// could lead to unwanted nginx behavior.
func (GenericValidator) ValidateEscapedStringNoVarExpansion(value string) error {
	return validateEscapedStringNoVarExpansion(value, nil)
}

const (
	alphaNumericStringFmt    = `[a-zA-Z0-9_-]+`
	alphaNumericStringErrMsg = "must contain only alphanumeric characters or '-' or '_'"
)

var alphaNumericStringFmtRegexp = regexp.MustCompile("^" + alphaNumericStringFmt + "$")

// ValidateServiceName validates a service name that can only use alphanumeric characters.
func (GenericValidator) ValidateServiceName(name string) error {
	if !alphaNumericStringFmtRegexp.MatchString(name) {
		examples := []string{
			"svc1",
			"svc-1",
			"svc_1",
		}

		return errors.New(k8svalidation.RegexError(alphaNumericStringErrMsg, alphaNumericStringFmt, examples...))
	}

	return nil
}

const (
	durationStringFmt    = `\d{1,4}(ms|s)?`
	durationStringErrMsg = "must contain a number followed by 'ms' or 's'"
)

var durationStringFmtRegexp = regexp.MustCompile("^" + durationStringFmt + "$")

// ValidateNginxDuration validates a duration string that nginx can understand.
func (GenericValidator) ValidateNginxDuration(duration string) error {
	if !durationStringFmtRegexp.MatchString(duration) {
		examples := []string{
			"5ms",
			"10s",
		}

		return errors.New(k8svalidation.RegexError(durationStringFmt, durationStringErrMsg, examples...))
	}

	return nil
}

const (
	//nolint:lll
	endpointStringFmt    = `(?:http?:\/\/)?[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?(?:\.[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?)*(?::\d{1,5})?`
	endpointStringErrMsg = "must be an alphanumeric hostname with optional http scheme and optional port"
)

var endpointStringFmtRegexp = regexp.MustCompile("^" + endpointStringFmt + "$")

// ValidateEndpoint validates an alphanumeric endpoint, with optional http scheme and port.
func (GenericValidator) ValidateEndpoint(endpoint string) error {
	if !endpointStringFmtRegexp.MatchString(endpoint) {
		examples := []string{
			"my-endpoint",
			"my.endpoint:5678",
			"http://my-endpoint",
		}

		return errors.New(k8svalidation.RegexError(endpointStringFmt, endpointStringErrMsg, examples...))
	}

	return nil
}
