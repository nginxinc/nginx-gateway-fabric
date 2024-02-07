package graph

import (
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
)

func TestValidateBackendTLSPolicy(t *testing.T) {
	targetRefNormalCase := &v1alpha2.PolicyTargetReferenceWithSectionName{
		PolicyTargetReference: v1alpha2.PolicyTargetReference{
			Kind:      "Service",
			Name:      "service1",
			Namespace: (*gatewayv1.Namespace)(helpers.GetPointer("test")),
		},
	}

	localObjectRefNormalCase := []gatewayv1.LocalObjectReference{
		{
			Kind:  "ConfigMap",
			Name:  "configmap",
			Group: "",
		},
	}

	localObjectRefInvalidName := []gatewayv1.LocalObjectReference{
		{
			Kind:  "ConfigMap",
			Name:  "invalid",
			Group: "",
		},
	}

	localObjectRefInvalidKind := []gatewayv1.LocalObjectReference{
		{
			Kind:  "Secret",
			Name:  "secret",
			Group: "",
		},
	}

	localObjectRefInvalidGroup := []gatewayv1.LocalObjectReference{
		{
			Kind:  "ConfigMap",
			Name:  "configmap",
			Group: "bhu",
		},
	}

	localObjectRefTooManyCerts := append(localObjectRefNormalCase, localObjectRefInvalidName...)

	getAncestorRef := func(ctlrName, parentName string) v1alpha2.PolicyAncestorStatus {
		return v1alpha2.PolicyAncestorStatus{
			ControllerName: gatewayv1.GatewayController(ctlrName),
			AncestorRef: gatewayv1.ParentReference{
				Name:      gatewayv1.ObjectName(parentName),
				Namespace: helpers.GetPointer(gatewayv1.Namespace("test")),
			},
		}
	}

	ancestors := []v1alpha2.PolicyAncestorStatus{
		getAncestorRef("not-us", "not-us"),
		getAncestorRef("not-us", "not-us"),
		getAncestorRef("not-us", "not-us"),
		getAncestorRef("not-us", "not-us"),
		getAncestorRef("not-us", "not-us"),
		getAncestorRef("not-us", "not-us"),
		getAncestorRef("not-us", "not-us"),
		getAncestorRef("not-us", "not-us"),
		getAncestorRef("not-us", "not-us"),
		getAncestorRef("not-us", "not-us"),
		getAncestorRef("not-us", "not-us"),
		getAncestorRef("not-us", "not-us"),
		getAncestorRef("not-us", "not-us"),
		getAncestorRef("not-us", "not-us"),
		getAncestorRef("not-us", "not-us"),
		getAncestorRef("not-us", "not-us"),
	}

	ancestorsWithUs := append(ancestors, getAncestorRef("test", "gateway"))

	tests := []struct {
		tlsPolicy  *v1alpha2.BackendTLSPolicy
		caCertName types.NamespacedName
		name       string
		isValid    bool
		ignored    bool
	}{
		{
			name: "normal case with ca cert refs",
			tlsPolicy: &v1alpha2.BackendTLSPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tls-policy",
					Namespace: "test",
				},
				Spec: v1alpha2.BackendTLSPolicySpec{
					TargetRef: *targetRefNormalCase,
					TLS: v1alpha2.BackendTLSPolicyConfig{
						CACertRefs: localObjectRefNormalCase,
						Hostname:   "foo.test.com",
					},
				},
			},
			isValid:    true,
			caCertName: types.NamespacedName{Namespace: "test", Name: "configmap"},
		},
		{
			name: "normal case with ca cert refs and 16 ancestors including us",
			tlsPolicy: &v1alpha2.BackendTLSPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tls-policy",
					Namespace: "test",
				},
				Spec: v1alpha2.BackendTLSPolicySpec{
					TargetRef: *targetRefNormalCase,
					TLS: v1alpha2.BackendTLSPolicyConfig{
						CACertRefs: localObjectRefNormalCase,
						Hostname:   "foo.test.com",
					},
				},
				Status: v1alpha2.PolicyStatus{
					Ancestors: ancestorsWithUs,
				},
			},
			isValid:    true,
			caCertName: types.NamespacedName{Namespace: "test", Name: "configmap"},
		},
		{
			name: "normal case with well known certs",
			tlsPolicy: &v1alpha2.BackendTLSPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tls-policy",
					Namespace: "test",
				},
				Spec: v1alpha2.BackendTLSPolicySpec{
					TargetRef: *targetRefNormalCase,
					TLS: v1alpha2.BackendTLSPolicyConfig{
						WellKnownCACerts: (helpers.GetPointer(v1alpha2.WellKnownCACertSystem)),
						Hostname:         "foo.test.com",
					},
				},
			},
			isValid: true,
		},
		{
			name: "no hostname invalid case",
			tlsPolicy: &v1alpha2.BackendTLSPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tls-policy",
					Namespace: "test",
				},
				Spec: v1alpha2.BackendTLSPolicySpec{
					TargetRef: *targetRefNormalCase,
					TLS: v1alpha2.BackendTLSPolicyConfig{
						CACertRefs: localObjectRefNormalCase,
						Hostname:   "",
					},
				},
			},
			isValid:    false,
			caCertName: types.NamespacedName{Namespace: "test", Name: "configmap"},
		},
		{
			name: "invalid ca cert ref name",
			tlsPolicy: &v1alpha2.BackendTLSPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tls-policy",
					Namespace: "test",
				},
				Spec: v1alpha2.BackendTLSPolicySpec{
					TargetRef: *targetRefNormalCase,
					TLS: v1alpha2.BackendTLSPolicyConfig{
						CACertRefs: localObjectRefInvalidName,
						Hostname:   "foo.test.com",
					},
				},
			},
			isValid: false,
		},
		{
			name: "invalid ca cert ref kind",
			tlsPolicy: &v1alpha2.BackendTLSPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tls-policy",
					Namespace: "test",
				},
				Spec: v1alpha2.BackendTLSPolicySpec{
					TargetRef: *targetRefNormalCase,
					TLS: v1alpha2.BackendTLSPolicyConfig{
						CACertRefs: localObjectRefInvalidKind,
						Hostname:   "foo.test.com",
					},
				},
			},
			isValid: false,
		},
		{
			name: "invalid ca cert ref group",
			tlsPolicy: &v1alpha2.BackendTLSPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tls-policy",
					Namespace: "test",
				},
				Spec: v1alpha2.BackendTLSPolicySpec{
					TargetRef: *targetRefNormalCase,
					TLS: v1alpha2.BackendTLSPolicyConfig{
						CACertRefs: localObjectRefInvalidGroup,
						Hostname:   "foo.test.com",
					},
				},
			},
			isValid: false,
		},
		{
			name: "invalid case with well known certs",
			tlsPolicy: &v1alpha2.BackendTLSPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tls-policy",
					Namespace: "test",
				},
				Spec: v1alpha2.BackendTLSPolicySpec{
					TargetRef: *targetRefNormalCase,
					TLS: v1alpha2.BackendTLSPolicyConfig{
						WellKnownCACerts: (helpers.GetPointer(v1alpha2.WellKnownCACertType("unknown"))),
						Hostname:         "foo.test.com",
					},
				},
			},
			isValid: false,
		},
		{
			name: "invalid case neither TLS config option chosen",
			tlsPolicy: &v1alpha2.BackendTLSPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tls-policy",
					Namespace: "test",
				},
				Spec: v1alpha2.BackendTLSPolicySpec{
					TargetRef: *targetRefNormalCase,
					TLS: v1alpha2.BackendTLSPolicyConfig{
						Hostname: "foo.test.com",
					},
				},
			},
			isValid: false,
		},
		{
			name: "invalid case with too many ca cert refs",
			tlsPolicy: &v1alpha2.BackendTLSPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tls-policy",
					Namespace: "test",
				},
				Spec: v1alpha2.BackendTLSPolicySpec{
					TargetRef: *targetRefNormalCase,
					TLS: v1alpha2.BackendTLSPolicyConfig{
						CACertRefs: localObjectRefTooManyCerts,
						Hostname:   "foo.test.com",
					},
				},
			},
			isValid: false,
		},
		{
			name: "invalid case with too many ancestors",
			tlsPolicy: &v1alpha2.BackendTLSPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tls-policy",
					Namespace: "test",
				},
				Spec: v1alpha2.BackendTLSPolicySpec{
					TargetRef: *targetRefNormalCase,
					TLS: v1alpha2.BackendTLSPolicyConfig{
						CACertRefs: localObjectRefNormalCase,
						Hostname:   "foo.test.com",
					},
				},
				Status: v1alpha2.PolicyStatus{
					Ancestors: ancestors,
				},
			},
			isValid:    false,
			caCertName: types.NamespacedName{Namespace: "test", Name: "configmap"},
			ignored:    true,
		},
	}

	configMaps := map[types.NamespacedName]*v1.ConfigMap{
		{Namespace: "test", Name: "configmap"}: {
			ObjectMeta: metav1.ObjectMeta{
				Name:      "configmap",
				Namespace: "test",
			},
			Data: map[string]string{
				"ca.crt": caBlock,
			},
		},
		{Namespace: "test", Name: "invalid"}: {
			ObjectMeta: metav1.ObjectMeta{
				Name:      "invalid",
				Namespace: "test",
			},
			Data: map[string]string{
				"ca.crt": "invalid",
			},
		},
	}

	configMapResolver := newConfigMapResolver(configMaps)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			gateway := &Gateway{
				Source: &gatewayv1.Gateway{ObjectMeta: metav1.ObjectMeta{Name: "gateway", Namespace: "test"}},
			}

			valid, ignored, caCertName, conds := validateBackendTLSPolicy(
				test.tlsPolicy,
				configMapResolver,
				"test",
				gateway,
			)

			g.Expect(valid).To(Equal(test.isValid))
			g.Expect(caCertName).To(Equal(test.caCertName))
			g.Expect(ignored).To(Equal(test.ignored))
			if !test.isValid && !test.ignored {
				g.Expect(conds).To(HaveLen(1))
			} else {
				g.Expect(conds).To(HaveLen(0))
			}
		})
	}
}
