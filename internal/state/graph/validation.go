package graph

import (
	"errors"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation"
)

func validateHostname(hostname string) error {
	if hostname == "" {
		return errors.New("cannot be empty string")
	}

	if strings.Contains(hostname, "*") {
		return errors.New("wildcards are not supported")
	}

	msgs := validation.IsDNS1123Subdomain(hostname)
	if len(msgs) > 0 {
		combined := strings.Join(msgs, ",")
		return errors.New(combined)
	}

	return nil
}
