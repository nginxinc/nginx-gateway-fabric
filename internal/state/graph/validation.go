package graph

import (
	"errors"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation"
)

func validateHostname(hostname string) error {
	if hostname == "" {
		return fmt.Errorf("cannot be empty string")
	}

	if strings.Contains(hostname, "*") {
		return fmt.Errorf("wildcards are not supported")
	}

	msgs := validation.IsDNS1123Subdomain(hostname)
	if len(msgs) > 0 {
		combined := strings.Join(msgs, ",")
		return errors.New(combined)
	}

	return nil
}
