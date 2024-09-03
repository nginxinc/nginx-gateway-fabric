package graph

import (
	"errors"
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	v1 "sigs.k8s.io/gateway-api/apis/v1"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/kinds"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/validation/validationfakes"
)

func TestGetNginxProxy(t *testing.T) {
	tests := []struct {
		nps   map[types.NamespacedName]*ngfAPI.NginxProxy
		gc    *v1.GatewayClass
		expNP *NginxProxy
		name  string
	}{
		{
			nps: map[types.NamespacedName]*ngfAPI.NginxProxy{
				{Name: "np1"}: {},
			},
			gc:    nil,
			expNP: nil,
			name:  "nil gatewayclass",
		},
		{
			nps: map[types.NamespacedName]*ngfAPI.NginxProxy{},
			gc: &v1.GatewayClass{
				Spec: v1.GatewayClassSpec{
					ParametersRef: &v1.ParametersReference{
						Group: ngfAPI.GroupName,
						Kind:  v1.Kind(kinds.NginxProxy),
						Name:  "np1",
					},
				},
			},
			expNP: nil,
			name:  "no nginxproxy resources",
		},
		{
			nps: map[types.NamespacedName]*ngfAPI.NginxProxy{
				{Name: "np1"}: {
					ObjectMeta: metav1.ObjectMeta{
						Name: "np1",
					},
				},
				{Name: "np2"}: {
					ObjectMeta: metav1.ObjectMeta{
						Name: "np2",
					},
				},
			},
			gc: &v1.GatewayClass{
				Spec: v1.GatewayClassSpec{
					ParametersRef: &v1.ParametersReference{
						Group: ngfAPI.GroupName,
						Kind:  v1.Kind(kinds.NginxProxy),
						Name:  "np2",
					},
				},
			},
			expNP: &NginxProxy{
				Source: &ngfAPI.NginxProxy{
					ObjectMeta: metav1.ObjectMeta{
						Name: "np2",
					},
					Spec: ngfAPI.NginxProxySpec{
						IPFamily: helpers.GetPointer(ngfAPI.Dual),
					},
				},
				Valid: true,
			},
			name: "returns correct resource",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			g.Expect(buildNginxProxy(test.nps, test.gc, &validationfakes.FakeGenericValidator{})).To(Equal(test.expNP))
		})
	}
}

func TestIsNginxProxyReferenced(t *testing.T) {
	tests := []struct {
		gc     *GatewayClass
		npName types.NamespacedName
		name   string
		expRes bool
	}{
		{
			gc: &GatewayClass{
				Source: &v1.GatewayClass{
					Spec: v1.GatewayClassSpec{
						ParametersRef: &v1.ParametersReference{
							Group: ngfAPI.GroupName,
							Kind:  v1.Kind(kinds.NginxProxy),
							Name:  "nginx-proxy",
						},
					},
				},
			},
			npName: types.NamespacedName{},
			expRes: false,
			name:   "nil nginxproxy",
		},
		{
			gc:     nil,
			npName: types.NamespacedName{Name: "nginx-proxy"},
			expRes: false,
			name:   "nil gatewayclass",
		},
		{
			gc: &GatewayClass{
				Source: nil,
			},
			npName: types.NamespacedName{Name: "nginx-proxy"},
			expRes: false,
			name:   "nil gatewayclass source",
		},
		{
			gc: &GatewayClass{
				Source: &v1.GatewayClass{
					Spec: v1.GatewayClassSpec{
						ParametersRef: &v1.ParametersReference{
							Group: ngfAPI.GroupName,
							Kind:  v1.Kind(kinds.NginxProxy),
							Name:  "nginx-proxy",
						},
					},
				},
			},
			npName: types.NamespacedName{Name: "nginx-proxy"},
			expRes: true,
			name:   "references the NginxProxy",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			g.Expect(isNginxProxyReferenced(test.npName, test.gc)).To(Equal(test.expRes))
		})
	}
}

func TestGCReferencesAnyNginxProxy(t *testing.T) {
	tests := []struct {
		gc     *v1.GatewayClass
		name   string
		expRes bool
	}{
		{
			gc:     nil,
			expRes: false,
			name:   "nil gatewayclass",
		},
		{
			gc: &v1.GatewayClass{
				Spec: v1.GatewayClassSpec{},
			},
			expRes: false,
			name:   "nil paramsRef",
		},
		{
			gc: &v1.GatewayClass{
				Spec: v1.GatewayClassSpec{
					ParametersRef: &v1.ParametersReference{
						Group: v1.Group("wrong-group"),
						Kind:  v1.Kind(kinds.NginxProxy),
						Name:  "wrong-group",
					},
				},
			},
			expRes: false,
			name:   "wrong group name",
		},
		{
			gc: &v1.GatewayClass{
				Spec: v1.GatewayClassSpec{
					ParametersRef: &v1.ParametersReference{
						Group: ngfAPI.GroupName,
						Kind:  v1.Kind("WrongKind"),
						Name:  "wrong-kind",
					},
				},
			},
			expRes: false,
			name:   "wrong kind",
		},
		{
			gc: &v1.GatewayClass{
				Spec: v1.GatewayClassSpec{
					ParametersRef: &v1.ParametersReference{
						Group: ngfAPI.GroupName,
						Kind:  v1.Kind(kinds.NginxProxy),
						Name:  "nginx-proxy",
					},
				},
			},
			expRes: true,
			name:   "references an NginxProxy",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			g.Expect(gcReferencesAnyNginxProxy(test.gc)).To(Equal(test.expRes))
		})
	}
}

func createValidValidator() *validationfakes.FakeGenericValidator {
	v := &validationfakes.FakeGenericValidator{}
	v.ValidateEscapedStringNoVarExpansionReturns(nil)
	v.ValidateEndpointReturns(nil)
	v.ValidateServiceNameReturns(nil)
	v.ValidateNginxDurationReturns(nil)

	return v
}

func createInvalidValidator() *validationfakes.FakeGenericValidator {
	v := &validationfakes.FakeGenericValidator{}
	v.ValidateEscapedStringNoVarExpansionReturns(errors.New("error"))
	v.ValidateEndpointReturns(errors.New("error"))
	v.ValidateServiceNameReturns(errors.New("error"))
	v.ValidateNginxDurationReturns(errors.New("error"))

	return v
}

func TestValidateNginxProxy(t *testing.T) {
	tests := []struct {
		np              *ngfAPI.NginxProxy
		validator       *validationfakes.FakeGenericValidator
		name            string
		expErrSubstring string
		expectErrCount  int
	}{
		{
			name:      "valid nginxproxy",
			validator: createValidValidator(),
			np: &ngfAPI.NginxProxy{
				Spec: ngfAPI.NginxProxySpec{
					Telemetry: &ngfAPI.Telemetry{
						ServiceName: helpers.GetPointer("my-svc"),
						Exporter: &ngfAPI.TelemetryExporter{
							Interval: helpers.GetPointer[ngfAPI.Duration]("5ms"),
							Endpoint: "my-endpoint",
						},
						SpanAttributes: []ngfAPI.SpanAttribute{
							{Key: "key", Value: "value"},
						},
					},
					IPFamily: helpers.GetPointer[ngfAPI.IPFamilyType](ngfAPI.Dual),
					RewriteClientIP: &ngfAPI.RewriteClientIP{
						SetIPRecursively: helpers.GetPointer(true),
						TrustedAddresses: []ngfAPI.TrustedAddress{"2001:db8:a0b:12f0::1/32", "1.1.1.1"},
						Mode:             helpers.GetPointer(ngfAPI.RewriteClientIPModeProxyProtocol),
					},
				},
			},
			expectErrCount: 0,
		},
		{
			name:      "invalid serviceName",
			validator: createInvalidValidator(),
			np: &ngfAPI.NginxProxy{
				Spec: ngfAPI.NginxProxySpec{
					Telemetry: &ngfAPI.Telemetry{
						ServiceName: helpers.GetPointer("my-svc"), // any value is invalid by the validator
					},
				},
			},
			expErrSubstring: "telemetry.serviceName",
			expectErrCount:  1,
		},
		{
			name:      "invalid endpoint",
			validator: createInvalidValidator(),
			np: &ngfAPI.NginxProxy{
				Spec: ngfAPI.NginxProxySpec{
					Telemetry: &ngfAPI.Telemetry{
						Exporter: &ngfAPI.TelemetryExporter{
							Endpoint: "my-endpoint", // any value is invalid by the validator
						},
					},
				},
			},
			expErrSubstring: "telemetry.exporter.endpoint",
			expectErrCount:  1,
		},
		{
			name:      "invalid interval",
			validator: createInvalidValidator(),
			np: &ngfAPI.NginxProxy{
				Spec: ngfAPI.NginxProxySpec{
					Telemetry: &ngfAPI.Telemetry{
						Exporter: &ngfAPI.TelemetryExporter{
							Interval: helpers.GetPointer[ngfAPI.Duration](
								"my-interval",
							), // any value is invalid by the validator
						},
					},
				},
			},
			expErrSubstring: "telemetry.exporter.interval",
			expectErrCount:  1,
		},
		{
			name:      "invalid spanAttributes",
			validator: createInvalidValidator(),
			np: &ngfAPI.NginxProxy{
				Spec: ngfAPI.NginxProxySpec{
					Telemetry: &ngfAPI.Telemetry{
						SpanAttributes: []ngfAPI.SpanAttribute{
							{Key: "my-key", Value: "my-value"}, // any value is invalid by the validator
						},
					},
				},
			},
			expErrSubstring: "telemetry.spanAttributes",
			expectErrCount:  2,
		},
		{
			name:      "invalid ipFamily type",
			validator: createInvalidValidator(),
			np: &ngfAPI.NginxProxy{
				Spec: ngfAPI.NginxProxySpec{
					Telemetry: &ngfAPI.Telemetry{},
					IPFamily:  helpers.GetPointer[ngfAPI.IPFamilyType]("invalid"),
				},
			},
			expErrSubstring: "spec.ipFamily",
			expectErrCount:  1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			allErrs := validateNginxProxy(test.validator, test.np)
			g.Expect(allErrs).To(HaveLen(test.expectErrCount))
			if len(allErrs) > 0 {
				g.Expect(allErrs.ToAggregate().Error()).To(ContainSubstring(test.expErrSubstring))
			}
		})
	}
}

func TestValidateRewriteClientIP(t *testing.T) {
	tests := []struct {
		np             *ngfAPI.NginxProxy
		validator      *validationfakes.FakeGenericValidator
		name           string
		errorString    string
		expectErrCount int
	}{
		{
			name:      "valid rewriteClientIP",
			validator: createValidValidator(),
			np: &ngfAPI.NginxProxy{
				Spec: ngfAPI.NginxProxySpec{
					RewriteClientIP: &ngfAPI.RewriteClientIP{
						SetIPRecursively: helpers.GetPointer(true),
						TrustedAddresses: []ngfAPI.TrustedAddress{"2001:db8:a0b:12f0::1/32", "10.56.32.11/32"},
						Mode:             helpers.GetPointer(ngfAPI.RewriteClientIPModeProxyProtocol),
					},
				},
			},
			expectErrCount: 0,
		},
		{
			name:      "invalid CIDR in trustedAddresses",
			validator: createInvalidValidator(),
			np: &ngfAPI.NginxProxy{
				Spec: ngfAPI.NginxProxySpec{
					RewriteClientIP: &ngfAPI.RewriteClientIP{
						SetIPRecursively: helpers.GetPointer(true),
						TrustedAddresses: []ngfAPI.TrustedAddress{"2001:db8::/129", "10.0.0.1"},
						Mode:             helpers.GetPointer(ngfAPI.RewriteClientIPModeProxyProtocol),
					},
				},
			},
			expectErrCount: 1,
			errorString: "spec.rewriteClientIP.trustedAddresses.2001:db8::/129: " +
				"Invalid value: \"2001:db8::/129\": must be a valid IP address or CIDR range",
		},
		{
			name:      "invalid IP and valid CIDR in trustedAddresses",
			validator: createInvalidValidator(),
			np: &ngfAPI.NginxProxy{
				Spec: ngfAPI.NginxProxySpec{
					RewriteClientIP: &ngfAPI.RewriteClientIP{
						SetIPRecursively: helpers.GetPointer(true),
						TrustedAddresses: []ngfAPI.TrustedAddress{"2001:db8::1/48", "256.100.50.25"},
						Mode:             helpers.GetPointer(ngfAPI.RewriteClientIPModeProxyProtocol),
					},
				},
			},
			expectErrCount: 1,
			errorString: "spec.rewriteClientIP.trustedAddresses.256.100.50.25: " +
				"Invalid value: \"256.100.50.25\": must be a valid IP address or CIDR range",
		},
		{
			name:      "invalid when mode is set and trustedAddresses is empty",
			validator: createInvalidValidator(),
			np: &ngfAPI.NginxProxy{
				Spec: ngfAPI.NginxProxySpec{
					RewriteClientIP: &ngfAPI.RewriteClientIP{
						Mode: helpers.GetPointer(ngfAPI.RewriteClientIPModeProxyProtocol),
					},
				},
			},
			expectErrCount: 1,
			errorString:    "spec.rewriteClientIP: Required value: trustedAddresses field required when mode is set",
		},
		{
			name:      "invalid when trustedAddresses is greater in length than 16",
			validator: createInvalidValidator(),
			np: &ngfAPI.NginxProxy{
				Spec: ngfAPI.NginxProxySpec{
					RewriteClientIP: &ngfAPI.RewriteClientIP{
						Mode: helpers.GetPointer(ngfAPI.RewriteClientIPModeProxyProtocol),
						TrustedAddresses: []ngfAPI.TrustedAddress{
							"2001:db8:a0b:12f0::1/32", "2001:db8:a0b:12f0::1/32", "2001:db8:a0b:12f0::1/32",
							"2001:db8:a0b:12f0::1/32", "2001:db8:a0b:12f0::1/32", "2001:db8:a0b:12f0::1/32",
							"2001:db8:a0b:12f0::1/32", "2001:db8:a0b:12f0::1/32", "2001:db8:a0b:12f0::1/32",
							"2001:db8:a0b:12f0::1/32", "2001:db8:a0b:12f0::1/32", "2001:db8:a0b:12f0::1/32",
							"2001:db8:a0b:12f0::1/32", "2001:db8:a0b:12f0::1/32",
							"2001:db8:a0b:12f0::1/32", "2001:db8:a0b:12f0::1/32", "2001:db8:a0b:12f0::1/32",
						},
					},
				},
			},
			expectErrCount: 1,
			errorString:    "spec.rewriteClientIP.trustedAddresses: Too long: may not be longer than 16",
		},
		{
			name:      "invalid when mode is not proxyProtocol or XForwardedFor",
			validator: createInvalidValidator(),
			np: &ngfAPI.NginxProxy{
				Spec: ngfAPI.NginxProxySpec{
					RewriteClientIP: &ngfAPI.RewriteClientIP{
						Mode:             helpers.GetPointer(ngfAPI.RewriteClientIPModeType("invalid")),
						TrustedAddresses: []ngfAPI.TrustedAddress{"2001:db8:a0b:12f0::1/32", "10.0.0.1/32"},
					},
				},
			},
			expectErrCount: 1,
			errorString: "spec.rewriteClientIP.mode: Unsupported value: \"invalid\": " +
				"supported values: \"ProxyProtocol\", \"XForwardedFor\"",
		},
		{
			name:      "invalid when mode is not proxyProtocol or XForwardedFor and trustedAddresses is empty",
			validator: createInvalidValidator(),
			np: &ngfAPI.NginxProxy{
				Spec: ngfAPI.NginxProxySpec{
					RewriteClientIP: &ngfAPI.RewriteClientIP{
						Mode: helpers.GetPointer(ngfAPI.RewriteClientIPModeType("invalid")),
					},
				},
			},
			expectErrCount: 2,
			errorString: "[spec.rewriteClientIP: Required value: trustedAddresses field " +
				"required when mode is set, spec.rewriteClientIP.mode: " +
				"Unsupported value: \"invalid\": supported values: \"ProxyProtocol\", \"XForwardedFor\"]",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			allErrs := validateRewriteClientIP(test.np)
			g.Expect(allErrs).To(HaveLen(test.expectErrCount))
			if len(allErrs) > 0 {
				fmt.Println(allErrs.ToAggregate().Error())
				g.Expect(allErrs.ToAggregate().Error()).To(Equal(test.errorString))
			}
		})
	}
}
