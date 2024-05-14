package observability_test

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/kinds"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/config/validation"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/policies"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/policies/observability"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/policies/policiesfakes"
)

type policyModFunc func(policy *ngfAPI.ObservabilityPolicy) *ngfAPI.ObservabilityPolicy

func createValidPolicy() *ngfAPI.ObservabilityPolicy {
	return &ngfAPI.ObservabilityPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
		},
		Spec: ngfAPI.ObservabilityPolicySpec{
			TargetRefs: []v1alpha2.LocalPolicyTargetReference{
				{
					Group: gatewayv1.GroupName,
					Kind:  kinds.HTTPRoute,
					Name:  "gateway",
				},
			},
			Tracing: &ngfAPI.Tracing{
				Strategy: ngfAPI.TraceStrategyRatio,
				Context:  helpers.GetPointer[ngfAPI.TraceContext](ngfAPI.TraceContextExtract),
				SpanName: helpers.GetPointer("spanName"),
				SpanAttributes: []ngfAPI.SpanAttribute{
					{Key: "key", Value: "value"},
				},
			},
		},
		Status: v1alpha2.PolicyStatus{},
	}
}

func createModifiedPolicy(mod policyModFunc) *ngfAPI.ObservabilityPolicy {
	return mod(createValidPolicy())
}

func TestValidator_Validate(t *testing.T) {
	tests := []struct {
		name             string
		policy           *ngfAPI.ObservabilityPolicy
		globalSettings   *policies.GlobalPolicySettings
		expErrSubstrings []string
	}{
		{
			name:             "global settings are nil",
			policy:           createValidPolicy(),
			expErrSubstrings: []string{"NginxProxy configuration is either invalid or not attached"},
		},
		{
			name:             "global settings are invalid",
			policy:           createValidPolicy(),
			globalSettings:   &policies.GlobalPolicySettings{NginxProxyValid: false},
			expErrSubstrings: []string{"NginxProxy configuration is either invalid or not attached"},
		},
		{
			name: "invalid target ref; unsupported group",
			policy: createModifiedPolicy(func(p *ngfAPI.ObservabilityPolicy) *ngfAPI.ObservabilityPolicy {
				p.Spec.TargetRefs[0].Group = "Unsupported"
				return p
			}),
			globalSettings:   &policies.GlobalPolicySettings{NginxProxyValid: true},
			expErrSubstrings: []string{"spec.targetRefs.group"},
		},
		{
			name: "invalid target ref; unsupported kind",
			policy: createModifiedPolicy(func(p *ngfAPI.ObservabilityPolicy) *ngfAPI.ObservabilityPolicy {
				p.Spec.TargetRefs[0].Kind = "Unsupported"
				return p
			}),
			globalSettings:   &policies.GlobalPolicySettings{NginxProxyValid: true},
			expErrSubstrings: []string{"spec.targetRefs.kind"},
		},
		{
			name: "invalid strategy",
			policy: createModifiedPolicy(func(p *ngfAPI.ObservabilityPolicy) *ngfAPI.ObservabilityPolicy {
				p.Spec.Tracing.Strategy = "invalid"
				return p
			}),
			globalSettings:   &policies.GlobalPolicySettings{NginxProxyValid: true},
			expErrSubstrings: []string{"spec.tracing.strategy"},
		},
		{
			name: "invalid context",
			policy: createModifiedPolicy(func(p *ngfAPI.ObservabilityPolicy) *ngfAPI.ObservabilityPolicy {
				p.Spec.Tracing.Context = helpers.GetPointer[ngfAPI.TraceContext]("invalid")
				return p
			}),
			globalSettings:   &policies.GlobalPolicySettings{NginxProxyValid: true},
			expErrSubstrings: []string{"spec.tracing.context"},
		},
		{
			name: "invalid span name",
			policy: createModifiedPolicy(func(p *ngfAPI.ObservabilityPolicy) *ngfAPI.ObservabilityPolicy {
				p.Spec.Tracing.SpanName = helpers.GetPointer("invalid$$$")
				return p
			}),
			globalSettings:   &policies.GlobalPolicySettings{NginxProxyValid: true},
			expErrSubstrings: []string{"spec.tracing.spanName"},
		},
		{
			name: "invalid span attribute key",
			policy: createModifiedPolicy(func(p *ngfAPI.ObservabilityPolicy) *ngfAPI.ObservabilityPolicy {
				p.Spec.Tracing.SpanAttributes[0].Key = "invalid$$$"
				return p
			}),
			globalSettings:   &policies.GlobalPolicySettings{NginxProxyValid: true},
			expErrSubstrings: []string{"spec.tracing.spanAttributes.key"},
		},
		{
			name: "invalid span attribute value",
			policy: createModifiedPolicy(func(p *ngfAPI.ObservabilityPolicy) *ngfAPI.ObservabilityPolicy {
				p.Spec.Tracing.SpanAttributes[0].Value = "invalid$$$"
				return p
			}),
			globalSettings:   &policies.GlobalPolicySettings{NginxProxyValid: true},
			expErrSubstrings: []string{"spec.tracing.spanAttributes.value"},
		},
		{
			name:             "valid",
			policy:           createValidPolicy(),
			globalSettings:   &policies.GlobalPolicySettings{NginxProxyValid: true},
			expErrSubstrings: nil,
		},
	}

	v := observability.NewValidator(validation.GenericValidator{})

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			conds := v.Validate(test.policy, test.globalSettings)

			if len(test.expErrSubstrings) == 0 {
				g.Expect(conds).To(BeEmpty())
			} else {
				g.Expect(conds).ToNot(BeEmpty())
			}

			for _, str := range test.expErrSubstrings {
				var msg string
				for _, cond := range conds {
					if strings.Contains(cond.Message, str) {
						msg = cond.Message
						break
					}
				}
				g.Expect(msg).To(ContainSubstring(str), fmt.Sprintf("error not found in %v", conds))
			}
		})
	}
}

func TestValidator_ValidatePanics(t *testing.T) {
	v := observability.NewValidator(nil)

	validate := func() {
		_ = v.Validate(&policiesfakes.FakePolicy{}, nil)
	}

	g := NewWithT(t)

	g.Expect(validate).To(Panic())
}

func TestValidator_Conflicts(t *testing.T) {
	tests := []struct {
		polA      *ngfAPI.ObservabilityPolicy
		polB      *ngfAPI.ObservabilityPolicy
		name      string
		conflicts bool
	}{
		{
			name: "no conflicts",
			polA: &ngfAPI.ObservabilityPolicy{
				Spec: ngfAPI.ObservabilityPolicySpec{
					Tracing: &ngfAPI.Tracing{},
				},
			},
			polB: &ngfAPI.ObservabilityPolicy{
				Spec: ngfAPI.ObservabilityPolicySpec{},
			},
			conflicts: false,
		},
		{
			name: "conflicts",
			polA: &ngfAPI.ObservabilityPolicy{
				Spec: ngfAPI.ObservabilityPolicySpec{
					Tracing: &ngfAPI.Tracing{},
				},
			},
			polB: &ngfAPI.ObservabilityPolicy{
				Spec: ngfAPI.ObservabilityPolicySpec{
					Tracing: &ngfAPI.Tracing{},
				},
			},
			conflicts: true,
		},
	}

	v := observability.NewValidator(nil)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			g.Expect(v.Conflicts(test.polA, test.polB)).To(Equal(test.conflicts))
		})
	}
}

func TestValidator_ConflictsPanics(t *testing.T) {
	v := observability.NewValidator(nil)

	conflicts := func() {
		_ = v.Conflicts(&policiesfakes.FakePolicy{}, &policiesfakes.FakePolicy{})
	}

	g := NewWithT(t)

	g.Expect(conflicts).To(Panic())
}
