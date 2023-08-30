package main

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation"
)

const (
	// nolint:lll
	// Regex from: https://github.com/kubernetes-sigs/gateway-api/blob/v0.8.0/apis/v1beta1/shared_types.go#L551
	controllerNameRegex = `^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*\/[A-Za-z0-9\/\-._~%!$&'()*+,;=:]+$` //nolint:lll
)

func validateGatewayControllerName(value string) error {
	if len(value) == 0 {
		return errors.New("must be set")
	}

	fields := strings.Split(value, "/")
	l := len(fields)
	if l < 2 {
		return errors.New("invalid format; must be DOMAIN/PATH")
	}

	if fields[0] != domain {
		return fmt.Errorf("invalid domain: %s; domain must be: %s", fields[0], domain)
	}

	re := regexp.MustCompile(controllerNameRegex)
	if !re.MatchString(value) {
		return fmt.Errorf("invalid gateway controller name: %s; expected format is DOMAIN/PATH", value)
	}

	return nil
}

func validateResourceName(value string) error {
	if len(value) == 0 {
		return errors.New("must be set")
	}

	// used by Kubernetes to validate resource names
	messages := validation.IsDNS1123Subdomain(value)
	if len(messages) > 0 {
		msg := strings.Join(messages, "; ")
		return fmt.Errorf("invalid format: %s", msg)
	}

	return nil
}

func validateNamespaceName(value string) error {
	// used by Kubernetes to validate resource namespace names
	messages := validation.IsDNS1123Label(value)
	if len(messages) > 0 {
		msg := strings.Join(messages, "; ")
		return fmt.Errorf("invalid format: %s", msg)
	}

	return nil
}

func parseNamespacedResourceName(value string) (types.NamespacedName, error) {
	if value == "" {
		return types.NamespacedName{}, errors.New("must be set")
	}

	parts := strings.Split(value, "/")
	if len(parts) != 2 {
		return types.NamespacedName{}, errors.New("invalid format; must be NAMESPACE/NAME")
	}

	if err := validateNamespaceName(parts[0]); err != nil {
		return types.NamespacedName{}, fmt.Errorf("invalid namespace name: %w", err)
	}
	if err := validateResourceName(parts[1]); err != nil {
		return types.NamespacedName{}, fmt.Errorf("invalid resource name: %w", err)
	}

	return types.NamespacedName{
		Namespace: parts[0],
		Name:      parts[1],
	}, nil
}

func validateIP(ip string) error {
	if ip == "" {
		return errors.New("IP address must be set")
	}
	if net.ParseIP(ip) == nil {
		return fmt.Errorf("%q must be a valid IP address", ip)
	}

	return nil
}
