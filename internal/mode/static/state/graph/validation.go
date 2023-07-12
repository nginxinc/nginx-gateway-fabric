package graph

import (
	"errors"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation"
)

func validateHostname(hostname string) error {
	if hostname == "" {
		return errors.New("cannot be empty string")
	}

	if strings.HasPrefix(hostname, "*.") {
		msgs := validation.IsWildcardDNS1123Subdomain(hostname)
		if len(msgs) > 0 {
			combined := strings.Join(msgs, ",")
			return errors.New(combined)
		}
		return nil
	}

	msgs := validation.IsDNS1123Subdomain(hostname)
	if len(msgs) > 0 {
		combined := strings.Join(msgs, ",")
		return errors.New(combined)
	}

	return nil
}

// panicForBrokenWebhookAssumption panics with the error message because an assumption about the webhook validation
// (run by NKG) is broken.
// Use it when you expect a validated Gateway API resource, but the actual resource breaks the validation constraints.
// For example, a field that must not be nil is nil.
func panicForBrokenWebhookAssumption(err error) {
	panic(fmt.Errorf("webhook validation assumption was broken: %w", err))
}
