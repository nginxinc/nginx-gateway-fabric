package state

import (
	"testing"

	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/helpers"
)

func TestValidateHTTPListener(t *testing.T) {
	tests := []struct {
		l        v1beta1.Listener
		expected bool
		msg      string
	}{
		{
			l: v1beta1.Listener{
				Port:     80,
				Protocol: v1beta1.HTTPProtocolType,
			},
			expected: true,
			msg:      "valid",
		},
		{
			l: v1beta1.Listener{
				Port:     81,
				Protocol: v1beta1.HTTPProtocolType,
			},
			expected: false,
			msg:      "invalid port",
		},
	}

	for _, test := range tests {
		result := validateHTTPListener(test.l)
		if result != test.expected {
			t.Errorf("validateListener() returned %v but expected %v for the case of %q", result, test.expected, test.msg)
		}
	}
}

func TestValidateHTTPSListener(t *testing.T) {
	gwNs := "gateway-ns"

	validSecretRef := &v1beta1.SecretObjectReference{
		Kind:      (*v1beta1.Kind)(helpers.GetStringPointer("Secret")),
		Name:      "secret",
		Namespace: (*v1beta1.Namespace)(helpers.GetStringPointer(gwNs)),
	}

	invalidSecretRefType := &v1beta1.SecretObjectReference{
		Kind:      (*v1beta1.Kind)(helpers.GetStringPointer("ConfigMap")),
		Name:      "secret",
		Namespace: (*v1beta1.Namespace)(helpers.GetStringPointer(gwNs)),
	}

	invalidSecretRefTNamespace := &v1beta1.SecretObjectReference{
		Kind:      (*v1beta1.Kind)(helpers.GetStringPointer("Secret")),
		Name:      "secret",
		Namespace: (*v1beta1.Namespace)(helpers.GetStringPointer("diff-ns")),
	}

	tests := []struct {
		l        v1beta1.Listener
		expected bool
		msg      string
	}{
		{
			l: v1beta1.Listener{
				Port:     443,
				Protocol: v1beta1.HTTPSProtocolType,
				TLS: &v1beta1.GatewayTLSConfig{
					Mode:            helpers.GetTLSModePointer(v1beta1.TLSModeTerminate),
					CertificateRefs: []v1beta1.SecretObjectReference{*validSecretRef},
				},
			},
			expected: true,
			msg:      "valid",
		},
		{
			l: v1beta1.Listener{
				Port:     80,
				Protocol: v1beta1.HTTPSProtocolType,
				TLS: &v1beta1.GatewayTLSConfig{
					Mode:            helpers.GetTLSModePointer(v1beta1.TLSModeTerminate),
					CertificateRefs: []v1beta1.SecretObjectReference{*validSecretRef},
				},
			},
			expected: false,
			msg:      "invalid port",
		},
		{
			l: v1beta1.Listener{
				Port:     443,
				Protocol: v1beta1.HTTPSProtocolType,
				TLS: &v1beta1.GatewayTLSConfig{
					Mode: helpers.GetTLSModePointer(v1beta1.TLSModeTerminate),
				},
			},
			expected: false,
			msg:      "invalid - no cert ref",
		},
		{
			l: v1beta1.Listener{
				Port:     443,
				Protocol: v1beta1.HTTPSProtocolType,
				TLS: &v1beta1.GatewayTLSConfig{
					Mode:            helpers.GetTLSModePointer(v1beta1.TLSModePassthrough),
					CertificateRefs: []v1beta1.SecretObjectReference{*validSecretRef},
				},
			},
			expected: false,
			msg:      "invalid tls mode",
		},
		{
			l: v1beta1.Listener{
				Port:     443,
				Protocol: v1beta1.HTTPSProtocolType,
				TLS: &v1beta1.GatewayTLSConfig{
					Mode:            helpers.GetTLSModePointer(v1beta1.TLSModeTerminate),
					CertificateRefs: []v1beta1.SecretObjectReference{*invalidSecretRefType},
				},
			},
			expected: false,
			msg:      "invalid cert ref kind",
		},
		{
			l: v1beta1.Listener{
				Port:     443,
				Protocol: v1beta1.HTTPSProtocolType,
				TLS: &v1beta1.GatewayTLSConfig{
					Mode:            helpers.GetTLSModePointer(v1beta1.TLSModeTerminate),
					CertificateRefs: []v1beta1.SecretObjectReference{*invalidSecretRefTNamespace},
				},
			},
			expected: false,
			msg:      "invalid cert ref namespace",
		},
		{
			l: v1beta1.Listener{
				Port:     443,
				Protocol: v1beta1.HTTPSProtocolType,
			},
			expected: false,
			msg:      "invalid - no tls config",
		},
	}

	for _, test := range tests {
		result := validateHTTPSListener(test.l, gwNs)
		if result != test.expected {
			t.Errorf("validateHTTPSListener() returned %v but expected %v for the case of %q", result, test.expected, test.msg)
		}
	}
}
