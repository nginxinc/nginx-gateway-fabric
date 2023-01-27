package state

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
				conditions.NewListenerPortUnavailable("Port 81 is not supported for HTTP, use 80"),
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
				conditions.NewListenerPortUnavailable("Port 80 is not supported for HTTPS, use 443"),
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
				conditions.NewListenerUnsupportedValue("tls.options are not supported"),
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
				conditions.NewListenerUnsupportedValue(`tls.mode "Passthrough" is not supported, use "Terminate"`),
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
			expected: conditions.NewListenerInvalidCertificateRef(`Group must be empty, got "some-group"`),
			name:     "invalid cert ref group",
		},
		{
			l: v1beta1.Listener{
				Port: 443,
				TLS: &v1beta1.GatewayTLSConfig{
					Mode:            helpers.GetTLSModePointer(v1beta1.TLSModeTerminate),
					CertificateRefs: []v1beta1.SecretObjectReference{invalidSecretRefKind},
				},
			},
			expected: conditions.NewListenerInvalidCertificateRef(`Kind must be Secret, got "ConfigMap"`),
			name:     "invalid cert ref kind",
		},
		{
			l: v1beta1.Listener{
				Port: 443,
				TLS: &v1beta1.GatewayTLSConfig{
					Mode:            helpers.GetTLSModePointer(v1beta1.TLSModeTerminate),
					CertificateRefs: []v1beta1.SecretObjectReference{invalidSecretRefTNamespace},
				},
			},
			expected: conditions.NewListenerInvalidCertificateRef("Referenced Secret must belong to the same " +
				"namespace as the Gateway"),
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
				conditions.NewListenerUnsupportedValue("Only 1 certificateRef is supported, got 2"),
			},
			name: "too many cert refs",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			result := validateHTTPSListener(test.l, gwNs)
			g.Expect(result).To(Equal(test.expected))
		})
	}
}

func TestValidateListenerHostname(t *testing.T) {
	tests := []struct {
		hostname          *v1beta1.Hostname
		expectedCondition conditions.Condition
		name              string
		expectedValid     bool
	}{
		{
			hostname:          nil,
			expectedValid:     true,
			expectedCondition: conditions.Condition{},
			name:              "nil hostname",
		},
		{
			hostname:          (*v1beta1.Hostname)(helpers.GetStringPointer("")),
			expectedValid:     true,
			expectedCondition: conditions.Condition{},
			name:              "empty hostname",
		},
		{
			hostname:          (*v1beta1.Hostname)(helpers.GetStringPointer("foo.example.com")),
			expectedValid:     true,
			expectedCondition: conditions.Condition{},
			name:              "valid hostname",
		},
		{
			hostname:          (*v1beta1.Hostname)(helpers.GetStringPointer("*.example.com")),
			expectedValid:     false,
			expectedCondition: conditions.NewListenerUnsupportedValue("Wildcard hostnames are not supported"),
			name:              "wildcard hostname",
		},
		{
			hostname:      (*v1beta1.Hostname)(helpers.GetStringPointer("example$com")),
			expectedValid: false,
			expectedCondition: conditions.NewListenerUnsupportedValue(
				"Invalid hostname: a lowercase RFC 1123 subdomain must consist of lower case alphanumeric " +
					"characters, '-' or '.', and must start and end with an alphanumeric character " +
					"(e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?" +
					`(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')`),
			name: "invalid hostname",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			valid, cond := validateListenerHostname(test.hostname)

			g.Expect(valid).To(Equal(test.expectedValid))
			g.Expect(cond).To(Equal(test.expectedCondition))
		})
	}
}
