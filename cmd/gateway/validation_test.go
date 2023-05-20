package main

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestValidateGatewayControllerName(t *testing.T) {
	tests := []struct {
		name   string
		value  string
		expErr bool
	}{
		{
			name:   "valid",
			value:  "k8s-gateway.nginx.org/nginx-gateway",
			expErr: false,
		},
		{
			name:   "valid - with subpath",
			value:  "k8s-gateway.nginx.org/nginx-gateway/my-gateway",
			expErr: false,
		},
		{
			name:   "valid - with complex subpath",
			value:  "k8s-gateway.nginx.org/nginx-gateway/my-gateway/v1",
			expErr: false,
		},
		{
			name:   "invalid - empty",
			value:  "",
			expErr: true,
		},
		{
			name:   "invalid - lacks path",
			value:  "k8s-gateway.nginx.org",
			expErr: true,
		},
		{
			name:   "invalid - lacks path, only slash is present",
			value:  "k8s-gateway.nginx.org/",
			expErr: true,
		},
		{
			name:   "invalid - invalid domain",
			value:  "invalid-domain/my-gateway",
			expErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			err := validateGatewayControllerName(test.value)

			if test.expErr {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).ToNot(HaveOccurred())
			}
		})
	}
}

func TestValidateResourceName(t *testing.T) {
	tests := []struct {
		name   string
		value  string
		expErr bool
	}{
		{
			name:   "valid",
			value:  "mygateway",
			expErr: false,
		},
		{
			name:   "valid - with dash",
			value:  "my-gateway",
			expErr: false,
		},
		{
			name:   "valid - with dash and numbers",
			value:  "my-gateway-123",
			expErr: false,
		},
		{
			name:   "invalid - empty",
			value:  "",
			expErr: true,
		},
		{
			name:   "invalid - invalid character '/'",
			value:  "my/gateway",
			expErr: true,
		},
		{
			name:   "invalid - invalid character '_'",
			value:  "my_gateway",
			expErr: true,
		},
		{
			name:   "invalid - invalid character '@'",
			value:  "my@gateway",
			expErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			err := validateResourceName(test.value)

			if test.expErr {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).ToNot(HaveOccurred())
			}
		})
	}
}

func TestValidateIP(t *testing.T) {
	tests := []struct {
		name      string
		expSubMsg string
		ip        string
		expErr    bool
	}{
		{
			name:      "var not set",
			ip:        "",
			expErr:    true,
			expSubMsg: "must be set",
		},
		{
			name:      "invalid ip address",
			ip:        "invalid",
			expErr:    true,
			expSubMsg: "must be a valid",
		},
		{
			name:   "valid ip address",
			ip:     "1.2.3.4",
			expErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			err := validateIP(tc.ip)
			if !tc.expErr {
				g.Expect(err).ToNot(HaveOccurred())
			} else {
				g.Expect(err.Error()).To(ContainSubstring(tc.expSubMsg))
			}
		})
	}
}
