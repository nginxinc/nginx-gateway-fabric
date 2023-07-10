package graph

import (
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/framework/conditions"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/framework/helpers"
	staticConds "github.com/nginxinc/nginx-kubernetes-gateway/internal/mode/static/state/conditions"
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
				Port: 0,
			},
			expected: staticConds.NewListenerUnsupportedValue(`port: Invalid value: 0: port must be between 1-65535`),
			name:     "invalid port",
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
	secretNs := "secret-ns"

	validSecretRef := v1beta1.SecretObjectReference{
		Kind:      (*v1beta1.Kind)(helpers.GetStringPointer("Secret")),
		Name:      "secret",
		Namespace: (*v1beta1.Namespace)(helpers.GetStringPointer(secretNs)),
	}

	invalidSecretRefGroup := v1beta1.SecretObjectReference{
		Group:     (*v1beta1.Group)(helpers.GetStringPointer("some-group")),
		Kind:      (*v1beta1.Kind)(helpers.GetStringPointer("Secret")),
		Name:      "secret",
		Namespace: (*v1beta1.Namespace)(helpers.GetStringPointer(secretNs)),
	}

	invalidSecretRefKind := v1beta1.SecretObjectReference{
		Kind:      (*v1beta1.Kind)(helpers.GetStringPointer("ConfigMap")),
		Name:      "secret",
		Namespace: (*v1beta1.Namespace)(helpers.GetStringPointer(secretNs)),
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
				Port: 0,
				TLS: &v1beta1.GatewayTLSConfig{
					Mode:            helpers.GetTLSModePointer(v1beta1.TLSModeTerminate),
					CertificateRefs: []v1beta1.SecretObjectReference{validSecretRef},
				},
			},
			expected: staticConds.NewListenerUnsupportedValue(`port: Invalid value: 0: port must be between 1-65535`),
			name:     "invalid port",
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
			expected: staticConds.NewListenerUnsupportedValue("tls.options: Forbidden: options are not supported"),
			name:     "invalid options",
		},
		{
			l: v1beta1.Listener{
				Port: 443,
				TLS: &v1beta1.GatewayTLSConfig{
					Mode:            helpers.GetTLSModePointer(v1beta1.TLSModePassthrough),
					CertificateRefs: []v1beta1.SecretObjectReference{validSecretRef},
				},
			},
			expected: staticConds.NewListenerUnsupportedValue(
				`tls.mode: Unsupported value: "Passthrough": supported values: "Terminate"`,
			),
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
			expected: staticConds.NewListenerInvalidCertificateRef(
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
			expected: staticConds.NewListenerInvalidCertificateRef(
				`tls.certificateRefs[0].kind: Unsupported value: "ConfigMap": supported values: "Secret"`,
			),
			name: "invalid cert ref kind",
		},
		{
			l: v1beta1.Listener{
				Port: 443,
				TLS: &v1beta1.GatewayTLSConfig{
					Mode:            helpers.GetTLSModePointer(v1beta1.TLSModeTerminate),
					CertificateRefs: []v1beta1.SecretObjectReference{validSecretRef, validSecretRef},
				},
			},
			expected: staticConds.NewListenerUnsupportedValue(
				"tls.certificateRefs: Too many: 2: must have at most 1 items",
			),
			name: "too many cert refs",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			v := createHTTPSListenerValidator()

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
			expectErr: false,
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

func TestGetAndValidateListenerSupportedKinds(t *testing.T) {
	HTTPRouteGroupKind := []v1beta1.RouteGroupKind{
		{
			Kind:  "HTTPRoute",
			Group: helpers.GetPointer[v1beta1.Group](v1beta1.GroupName),
		},
	}
	TCPRouteGroupKind := []v1beta1.RouteGroupKind{
		{
			Kind:  "TCPRoute",
			Group: helpers.GetPointer[v1beta1.Group](v1beta1.GroupName),
		},
	}
	tests := []struct {
		protocol  v1beta1.ProtocolType
		name      string
		kind      []v1beta1.RouteGroupKind
		expected  []v1beta1.RouteGroupKind
		expectErr bool
	}{
		{
			protocol:  v1beta1.TCPProtocolType,
			expectErr: false,
			name:      "unsupported protocol is ignored",
			kind:      TCPRouteGroupKind,
			expected:  []v1beta1.RouteGroupKind{},
		},
		{
			protocol: v1beta1.HTTPProtocolType,
			kind: []v1beta1.RouteGroupKind{
				{
					Kind:  "HTTPRoute",
					Group: helpers.GetPointer[v1beta1.Group]("bad-group"),
				},
			},
			expectErr: true,
			name:      "invalid group",
			expected:  []v1beta1.RouteGroupKind{},
		},
		{
			protocol:  v1beta1.HTTPProtocolType,
			kind:      TCPRouteGroupKind,
			expectErr: true,
			name:      "invalid kind",
			expected:  []v1beta1.RouteGroupKind{},
		},
		{
			protocol:  v1beta1.HTTPProtocolType,
			kind:      HTTPRouteGroupKind,
			expectErr: false,
			name:      "valid HTTP",
			expected:  HTTPRouteGroupKind,
		},
		{
			protocol:  v1beta1.HTTPSProtocolType,
			kind:      HTTPRouteGroupKind,
			expectErr: false,
			name:      "valid HTTPS",
			expected:  HTTPRouteGroupKind,
		},
		{
			protocol:  v1beta1.HTTPSProtocolType,
			expectErr: false,
			name:      "valid HTTPS no kind specified",
			expected: []v1beta1.RouteGroupKind{
				{
					Kind: "HTTPRoute",
				},
			},
		},
		{
			protocol: v1beta1.HTTPProtocolType,
			kind: []v1beta1.RouteGroupKind{
				{
					Kind:  "HTTPRoute",
					Group: helpers.GetPointer[v1beta1.Group](v1beta1.GroupName),
				},
				{
					Kind:  "bad-kind",
					Group: helpers.GetPointer[v1beta1.Group](v1beta1.GroupName),
				},
			},
			expectErr: true,
			name:      "valid and invalid kinds",
			expected:  HTTPRouteGroupKind,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			listener := v1beta1.Listener{
				Protocol: test.protocol,
			}

			if test.kind != nil {
				listener.AllowedRoutes = &v1beta1.AllowedRoutes{
					Kinds: test.kind,
				}
			}

			conds, kinds := getAndValidateListenerSupportedKinds(listener)
			g.Expect(helpers.Diff(test.expected, kinds)).To(BeEmpty())
			if test.expectErr {
				g.Expect(conds).ToNot(BeEmpty())
			} else {
				g.Expect(conds).To(BeEmpty())
			}
		})
	}
}

func TestValidateListenerLabelSelector(t *testing.T) {
	tests := []struct {
		selector  *metav1.LabelSelector
		from      v1beta1.FromNamespaces
		name      string
		expectErr bool
	}{
		{
			from:      v1beta1.NamespacesFromSelector,
			selector:  &metav1.LabelSelector{},
			expectErr: false,
			name:      "valid spec",
		},
		{
			from:      v1beta1.NamespacesFromSelector,
			selector:  nil,
			expectErr: true,
			name:      "invalid spec",
		},
		{
			from:      v1beta1.NamespacesFromAll,
			selector:  nil,
			expectErr: false,
			name:      "ignored from type",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			listener := v1beta1.Listener{
				AllowedRoutes: &v1beta1.AllowedRoutes{
					Namespaces: &v1beta1.RouteNamespaces{
						From:     &test.from,
						Selector: test.selector,
					},
				},
			}

			conds := validateListenerLabelSelector(listener)
			if test.expectErr {
				g.Expect(conds).ToNot(BeEmpty())
			} else {
				g.Expect(conds).To(BeEmpty())
			}
		})
	}
}

func TestValidateListenerPort(t *testing.T) {
	validPorts := []v1beta1.PortNumber{1, 80, 443, 1000, 50000, 65535}
	invalidPorts := []v1beta1.PortNumber{-1, 0, 65536, 80000}

	for _, p := range validPorts {
		t.Run(fmt.Sprintf("valid port %d", p), func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(validateListenerPort(p)).To(Succeed())
		})
	}

	for _, p := range invalidPorts {
		t.Run(fmt.Sprintf("invalid port %d", p), func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(validateListenerPort(p)).ToNot(Succeed())
		})
	}
}
