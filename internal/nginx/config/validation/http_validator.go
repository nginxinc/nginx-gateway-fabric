package validation

import (
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/validation"
)

// HTTPValidator validates values that will propagate into the NGINX configuration http context.
// The validation rules are based on the nginx/config/http types and how they are used in the configuration templates
// of the nginx/config package. Changes to those might require changing the validation rules
type HTTPValidator struct {
	HTTPNJSMatchValidator
	HTTPRedirectValidator
}

var _ validation.HTTPFieldsValidator = HTTPValidator{}

var hostnameInServerExamples = []string{"host", "example.com"}

// ValidateHostnameInServer validates a hostname to be used in the server_name directive.
func (HTTPValidator) ValidateHostnameInServer(hostname string) error {
	return validateEscapedString(hostname, hostnameInServerExamples)
}
