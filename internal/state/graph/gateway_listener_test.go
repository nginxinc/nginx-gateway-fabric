package graph

import (
	"testing"

	. "github.com/onsi/gomega"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/helpers"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/conditions"
)

func TestValidateHTTPListener(t *testing.T) {
	tests := []struct {
		l        v1beta1.Listener
		name     string
		expected []conditions.Condition
	}{
		{
			l: v1beta1.Listener{
				Port: 80,
			},
			expected: nil,
			name:     "valid",
		},
		{
			l: v1beta1.Listener{
				Port: 81,
			},
			expected: []conditions.Condition{
				conditions.NewListenerPortUnavailable(`port: Unsupported value: 81: supported values: "80"`),
			},
			name: "invalid port",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			result := validateHTTPListener(test.l)
			g.Expect(result).To(Equal(test.expected))
		})
	}
}

func TestValidateHTTPSListener(t *testing.T) {
	gwNs := "gateway-ns"

	validSecretRef := v1beta1.SecretObjectReference{
		Kind:      (*v1beta1.Kind)(helpers.GetStringPointer("Secret")),
		Name:      "secret",
		Namespace: (*v1beta1.Namespace)(helpers.GetStringPointer(gwNs)),
	}

	invalidSecretRefGroup := v1beta1.SecretObjectReference{
		Group:     (*v1beta1.Group)(helpers.GetStringPointer("some-group")),
		Kind:      (*v1beta1.Kind)(helpers.GetStringPointer("Secret")),
		Name:      "secret",
		Namespace: (*v1beta1.Namespace)(helpers.GetStringPointer(gwNs)),
	}

	invalidSecretRefKind := v1beta1.SecretObjectReference{
		Kind:      (*v1beta1.Kind)(helpers.GetStringPointer("ConfigMap")),
		Name:      "secret",
		Namespace: (*v1beta1.Namespace)(helpers.GetStringPointer(gwNs)),
	}

	invalidSecretRefTNamespace := v1beta1.SecretObjectReference{
		Kind:      (*v1beta1.Kind)(helpers.GetStringPointer("Secret")),
		Name:      "secret",
		Namespace: (*v1beta1.Namespace)(helpers.GetStringPointer("diff-ns")),
	}

	tests := []struct {
		l        v1beta1.Listener
		name     string
		expected []conditions.Condition
	}{
		{
			l: v1beta1.Listener{
				Port: 443,
				TLS: &v1beta1.GatewayTLSConfig{
					Mode:            helpers.GetTLSModePointer(v1beta1.TLSModeTerminate),
					CertificateRefs: []v1beta1.SecretObjectReference{validSecretRef},
				},
			},
			expected: nil,
			name:     "valid",
		},
		{
			l: v1beta1.Listener{
				Port: 80,
				TLS: &v1beta1.GatewayTLSConfig{
					Mode:            helpers.GetTLSModePointer(v1beta1.TLSModeTerminate),
					CertificateRefs: []v1beta1.SecretObjectReference{validSecretRef},
				},
			},
			expected: []conditions.Condition{
				conditions.NewListenerPortUnavailable(`port: Unsupported value: 80: supported values: "443"`),
			},
			name: "invalid port",
		},
		{
			l: v1beta1.Listener{
				Port: 443,
				TLS: &v1beta1.GatewayTLSConfig{
					Mode:            helpers.GetTLSModePointer(v1beta1.TLSModeTerminate),
					CertificateRefs: []v1beta1.SecretObjectReference{validSecretRef},
					Options:         map[v1beta1.AnnotationKey]v1beta1.AnnotationValue{"key": "val"},
				},
			},
			expected: []conditions.Condition{
				conditions.NewListenerUnsupportedValue("tls.options: Forbidden: options are not supported"),
			},
			name: "invalid options",
		},
		{
			l: v1beta1.Listener{
				Port: 443,
				TLS: &v1beta1.GatewayTLSConfig{
					Mode:            helpers.GetTLSModePointer(v1beta1.TLSModePassthrough),
					CertificateRefs: []v1beta1.SecretObjectReference{validSecretRef},
				},
			},
			expected: []conditions.Condition{
				conditions.NewListenerUnsupportedValue(
					`tls.mode: Unsupported value: "Passthrough": supported values: "Terminate"`,
				),
			},
			name: "invalid tls mode",
		},
		{
			l: v1beta1.Listener{
				Port: 443,
				TLS: &v1beta1.GatewayTLSConfig{
					Mode:            helpers.GetTLSModePointer(v1beta1.TLSModeTerminate),
					CertificateRefs: []v1beta1.SecretObjectReference{invalidSecretRefGroup},
				},
			},
			expected: conditions.NewListenerInvalidCertificateRef(
				`tls.certificateRefs[0].group: Unsupported value: "some-group": supported values: ""`,
			),
			name: "invalid cert ref group",
		},
		{
			l: v1beta1.Listener{
				Port: 443,
				TLS: &v1beta1.GatewayTLSConfig{
					Mode:            helpers.GetTLSModePointer(v1beta1.TLSModeTerminate),
					CertificateRefs: []v1beta1.SecretObjectReference{invalidSecretRefKind},
				},
			},
			expected: conditions.NewListenerInvalidCertificateRef(
				`tls.certificateRefs[0].kind: Unsupported value: "ConfigMap": supported values: "Secret"`,
			),
			name: "invalid cert ref kind",
		},
		{
			l: v1beta1.Listener{
				Port: 443,
				TLS: &v1beta1.GatewayTLSConfig{
					Mode:            helpers.GetTLSModePointer(v1beta1.TLSModeTerminate),
					CertificateRefs: []v1beta1.SecretObjectReference{invalidSecretRefTNamespace},
				},
			},
			expected: conditions.NewListenerInvalidCertificateRef(
				`tls.certificateRefs[0].namespace: Invalid value: "diff-ns": Referenced Secret must belong to ` +
					`the same namespace as the Gateway`,
			),
			name: "invalid cert ref namespace",
		},
		{
			l: v1beta1.Listener{
				Port: 443,
				TLS: &v1beta1.GatewayTLSConfig{
					Mode:            helpers.GetTLSModePointer(v1beta1.TLSModeTerminate),
					CertificateRefs: []v1beta1.SecretObjectReference{validSecretRef, validSecretRef},
				},
			},
			expected: []conditions.Condition{
				conditions.NewListenerUnsupportedValue("tls.certificateRefs: Too many: 2: must have at most 1 items"),
			},
			name: "too many cert refs",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			v := createHTTPSListenerValidator(gwNs)

			result := v(test.l)
			g.Expect(result).To(Equal(test.expected))
		})
	}
}

func TestValidateListenerHostname(t *testing.T) {
	tests := []struct {
		hostname  *v1beta1.Hostname
		name      string
		expectErr bool
	}{
		{
			hostname:  nil,
			expectErr: false,
			name:      "nil hostname",
		},
		{
			hostname:  (*v1beta1.Hostname)(helpers.GetStringPointer("")),
			expectErr: false,
			name:      "empty hostname",
		},
		{
			hostname:  (*v1beta1.Hostname)(helpers.GetStringPointer("foo.example.com")),
			expectErr: false,
			name:      "valid hostname",
		},
		{
			hostname:  (*v1beta1.Hostname)(helpers.GetStringPointer("*.example.com")),
			expectErr: true,
			name:      "wildcard hostname",
		},
		{
			hostname:  (*v1beta1.Hostname)(helpers.GetStringPointer("example$com")),
			expectErr: true,
			name:      "invalid hostname",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			conds := validateListenerHostname(v1beta1.Listener{Hostname: test.hostname})

			if test.expectErr {
				g.Expect(conds).ToNot(BeEmpty())
			} else {
				g.Expect(conds).To(BeEmpty())
			}
		})
	}
}

func TestValidateListenerAllowedRouteKind(t *testing.T) {
	tests := []struct {
		protocol  v1beta1.ProtocolType
		kind      v1beta1.Kind
		group     v1beta1.Group
		name      string
		expectErr bool
	}{
		{
			protocol:  v1beta1.TCPProtocolType,
			expectErr: false,
			name:      "unsupported protocol is ignored",
		},
		{
			protocol:  v1beta1.HTTPProtocolType,
			group:     "bad-group",
			kind:      "HTTPRoute",
			expectErr: true,
			name:      "invalid group",
		},
		{
			protocol:  v1beta1.HTTPProtocolType,
			group:     v1beta1.GroupName,
			kind:      "TCPRoute",
			expectErr: true,
			name:      "invalid kind",
		},
		{
			protocol:  v1beta1.HTTPProtocolType,
			group:     v1beta1.GroupName,
			kind:      "HTTPRoute",
			expectErr: false,
			name:      "valid HTTP",
		},
		{
			protocol:  v1beta1.HTTPSProtocolType,
			group:     v1beta1.GroupName,
			kind:      "HTTPRoute",
			expectErr: false,
			name:      "valid HTTPS",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			listener := v1beta1.Listener{
				Protocol: test.protocol,
				AllowedRoutes: &v1beta1.AllowedRoutes{
					Kinds: []v1beta1.RouteGroupKind{
						{
							Kind:  test.kind,
							Group: &test.group,
						},
					},
				},
			}

			conds := validateListenerAllowedRouteKind(listener)
			if test.expectErr {
				g.Expect(conds).ToNot(BeEmpty())
			} else {
				g.Expect(conds).To(BeEmpty())
			}
		})
	}
}
