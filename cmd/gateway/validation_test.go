package main

import (
	"testing"

	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
)

func TestValidateGatewayControllerName(t *testing.T) {
	tests := []struct {
		name   string
		value  string
		expErr bool
	}{
		{
			name:   "valid",
			value:  "gateway.nginx.org/nginx-gateway",
			expErr: false,
		},
		{
			name:   "valid - with subpath",
			value:  "gateway.nginx.org/nginx-gateway/my-gateway",
			expErr: false,
		},
		{
			name:   "valid - with complex subpath",
			value:  "gateway.nginx.org/nginx-gateway/my-gateway/v1",
			expErr: false,
		},
		{
			name:   "invalid - empty",
			value:  "",
			expErr: true,
		},
		{
			name:   "invalid - lacks path",
			value:  "gateway.nginx.org",
			expErr: true,
		},
		{
			name:   "invalid - lacks path, only slash is present",
			value:  "gateway.nginx.org/",
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
			g := NewWithT(t)

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
			name:   "valid - with dot",
			value:  "my.gateway",
			expErr: false,
		},
		{
			name:   "valid - with numbers",
			value:  "mygateway123",
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
			g := NewWithT(t)

			err := validateResourceName(test.value)

			if test.expErr {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).ToNot(HaveOccurred())
			}
		})
	}
}

func TestValidateNamespaceName(t *testing.T) {
	tests := []struct {
		name   string
		value  string
		expErr bool
	}{
		{
			name:   "valid",
			value:  "mynamespace",
			expErr: false,
		},
		{
			name:   "valid - with dash",
			value:  "my-namespace",
			expErr: false,
		},
		{
			name:   "valid - with numbers",
			value:  "mynamespace123",
			expErr: false,
		},
		{
			name:   "invalid - empty",
			value:  "",
			expErr: true,
		},
		{
			name:   "invalid - invalid character '.'",
			value:  "my.namespace",
			expErr: true,
		},
		{
			name:   "invalid - invalid character '/'",
			value:  "my/namespace",
			expErr: true,
		},
		{
			name:   "invalid - invalid character '_'",
			value:  "my_namespace",
			expErr: true,
		},
		{
			name:   "invalid - invalid character '@'",
			value:  "my@namespace",
			expErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			err := validateNamespaceName(test.value)

			if test.expErr {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).ToNot(HaveOccurred())
			}
		})
	}
}

func TestParseNamespacedResourceName(t *testing.T) {
	tests := []struct {
		name              string
		value             string
		expectedErrPrefix string
		expectedNsName    types.NamespacedName
		expectErr         bool
	}{
		{
			name:  "valid",
			value: "test/my-gateway",
			expectedNsName: types.NamespacedName{
				Namespace: "test",
				Name:      "my-gateway",
			},
			expectErr: false,
		},
		{
			name:              "empty",
			value:             "",
			expectedNsName:    types.NamespacedName{},
			expectErr:         true,
			expectedErrPrefix: "must be set",
		},
		{
			name:              "wrong number of parts",
			value:             "test",
			expectedNsName:    types.NamespacedName{},
			expectErr:         true,
			expectedErrPrefix: "invalid format; must be NAMESPACE/NAME",
		},
		{
			name:              "invalid namespace",
			value:             "t@st/my-gateway",
			expectedNsName:    types.NamespacedName{},
			expectErr:         true,
			expectedErrPrefix: "invalid namespace name",
		},
		{
			name:              "invalid name",
			value:             "test/my-g@teway",
			expectedNsName:    types.NamespacedName{},
			expectErr:         true,
			expectedErrPrefix: "invalid resource name",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			nsName, err := parseNamespacedResourceName(test.value)

			if test.expectErr {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(HavePrefix(test.expectedErrPrefix))
			} else {
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(nsName).To(Equal(test.expectedNsName))
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
			g := NewWithT(t)

			err := validateIP(tc.ip)
			if !tc.expErr {
				g.Expect(err).ToNot(HaveOccurred())
			} else {
				g.Expect(err.Error()).To(ContainSubstring(tc.expSubMsg))
			}
		})
	}
}

func TestValidatePort(t *testing.T) {
	tests := []struct {
		name   string
		port   int
		expErr bool
	}{
		{
			name:   "port under minimum allowed value",
			port:   1023,
			expErr: true,
		},
		{
			name:   "port over maximum allowed value",
			port:   65536,
			expErr: true,
		},
		{
			name:   "valid port",
			port:   9113,
			expErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)

			err := validatePort(tc.port)
			if !tc.expErr {
				g.Expect(err).ToNot(HaveOccurred())
			} else {
				g.Expect(err).To(HaveOccurred())
			}
		})
	}
}

func TestEnsureNoPortCollisions(t *testing.T) {
	g := NewWithT(t)

	g.Expect(ensureNoPortCollisions(9113, 8081)).To(Succeed())
	g.Expect(ensureNoPortCollisions(9113, 9113)).ToNot(Succeed())
}
