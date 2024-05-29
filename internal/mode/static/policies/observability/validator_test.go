package observability_test

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/kinds"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/config/validation"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/policies"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/policies/observability"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/policies/policiesfakes"
	staticConds "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/conditions"
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
					Name:  "route",
				},
			},
			Tracing: &ngfAPI.Tracing{
				Strategy: ngfAPI.TraceStrategyRatio,
				Context:  helpers.GetPointer(ngfAPI.TraceContextExtract),
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
	globalSettings := &policies.GlobalSettings{
		NginxProxyValid:  true,
		TelemetryEnabled: true,
	}

	tests := []struct {
		name           string
		policy         *ngfAPI.ObservabilityPolicy
		globalSettings *policies.GlobalSettings
		expConditions  []conditions.Condition
	}{
		{
			name:   "validation context is nil",
			policy: createValidPolicy(),
			expConditions: []conditions.Condition{
				staticConds.NewPolicyNotAcceptedNginxProxyNotSet(staticConds.PolicyMessageNginxProxyInvalid),
			},
		},
		{
			name:           "validation context is invalid",
			policy:         createValidPolicy(),
			globalSettings: &policies.GlobalSettings{NginxProxyValid: false},
			expConditions: []conditions.Condition{
				staticConds.NewPolicyNotAcceptedNginxProxyNotSet(staticConds.PolicyMessageNginxProxyInvalid),
			},
		},
		{
			name:           "telemetry is not enabled",
			policy:         createValidPolicy(),
			globalSettings: &policies.GlobalSettings{NginxProxyValid: true, TelemetryEnabled: false},
			expConditions: []conditions.Condition{
				staticConds.NewPolicyNotAcceptedNginxProxyNotSet(staticConds.PolicyMessageTelemetryNotEnabled),
			},
		},
		{
			name: "invalid target ref; unsupported group",
			policy: createModifiedPolicy(func(p *ngfAPI.ObservabilityPolicy) *ngfAPI.ObservabilityPolicy {
				p.Spec.TargetRefs[0].Group = "Unsupported"
				return p
			}),
			globalSettings: globalSettings,
			expConditions: []conditions.Condition{
				staticConds.NewPolicyInvalid("spec.targetRefs.group: Unsupported value: \"Unsupported\": " +
					"supported values: \"gateway.networking.k8s.io\""),
			},
		},
		{
			name: "invalid target ref; unsupported kind",
			policy: createModifiedPolicy(func(p *ngfAPI.ObservabilityPolicy) *ngfAPI.ObservabilityPolicy {
				p.Spec.TargetRefs[0].Kind = "Unsupported"
				return p
			}),
			globalSettings: globalSettings,
			expConditions: []conditions.Condition{
				staticConds.NewPolicyInvalid("spec.targetRefs.kind: Unsupported value: \"Unsupported\": " +
					"supported values: \"HTTPRoute\", \"GRPCRoute\""),
			},
		},
		{
			name: "invalid strategy",
			policy: createModifiedPolicy(func(p *ngfAPI.ObservabilityPolicy) *ngfAPI.ObservabilityPolicy {
				p.Spec.Tracing.Strategy = "invalid"
				return p
			}),
			globalSettings: globalSettings,
			expConditions: []conditions.Condition{
				staticConds.NewPolicyInvalid("spec.tracing.strategy: Unsupported value: \"invalid\": " +
					"supported values: \"ratio\", \"parent\""),
			},
		},
		{
			name: "invalid context",
			policy: createModifiedPolicy(func(p *ngfAPI.ObservabilityPolicy) *ngfAPI.ObservabilityPolicy {
				p.Spec.Tracing.Context = helpers.GetPointer[ngfAPI.TraceContext]("invalid")
				return p
			}),
			globalSettings: globalSettings,
			expConditions: []conditions.Condition{
				staticConds.NewPolicyInvalid("spec.tracing.context: Unsupported value: \"invalid\": " +
					"supported values: \"extract\", \"inject\", \"propagate\", \"ignore\""),
			},
		},
		{
			name: "invalid span name",
			policy: createModifiedPolicy(func(p *ngfAPI.ObservabilityPolicy) *ngfAPI.ObservabilityPolicy {
				p.Spec.Tracing.SpanName = helpers.GetPointer("invalid$$$")
				return p
			}),
			globalSettings: globalSettings,
			expConditions: []conditions.Condition{
				staticConds.NewPolicyInvalid("spec.tracing.spanName: Invalid value: \"invalid$$$\": " +
					"a valid value must have all '\"' escaped and must not contain any '$' or end with an " +
					"unescaped '\\' (regex used for validation is '([^\"$\\\\]|\\\\[^$])*')"),
			},
		},
		{
			name: "invalid span attribute key",
			policy: createModifiedPolicy(func(p *ngfAPI.ObservabilityPolicy) *ngfAPI.ObservabilityPolicy {
				p.Spec.Tracing.SpanAttributes[0].Key = "invalid$$$"
				return p
			}),
			globalSettings: globalSettings,
			expConditions: []conditions.Condition{
				staticConds.NewPolicyInvalid("spec.tracing.spanAttributes.key: Invalid value: \"invalid$$$\": " +
					"a valid value must have all '\"' escaped and must not contain any '$' or end with an " +
					"unescaped '\\' (regex used for validation is '([^\"$\\\\]|\\\\[^$])*')"),
			},
		},
		{
			name: "invalid span attribute value",
			policy: createModifiedPolicy(func(p *ngfAPI.ObservabilityPolicy) *ngfAPI.ObservabilityPolicy {
				p.Spec.Tracing.SpanAttributes[0].Value = "invalid$$$"
				return p
			}),
			globalSettings: globalSettings,
			expConditions: []conditions.Condition{
				staticConds.NewPolicyInvalid("spec.tracing.spanAttributes.value: Invalid value: \"invalid$$$\": " +
					"a valid value must have all '\"' escaped and must not contain any '$' or end with an " +
					"unescaped '\\' (regex used for validation is '([^\"$\\\\]|\\\\[^$])*')"),
			},
		},
		{
			name:           "valid",
			policy:         createValidPolicy(),
			globalSettings: globalSettings,
			expConditions:  nil,
		},
	}

	v := observability.NewValidator(validation.GenericValidator{})

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			conds := v.Validate(test.policy, test.globalSettings)
			g.Expect(conds).To(Equal(test.expConditions))
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
