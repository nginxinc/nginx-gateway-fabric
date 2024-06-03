package observability_test

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/policies"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/policies/observability"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/policies/policiesfakes"
)

func TestGenerate(t *testing.T) {
	ratio := helpers.GetPointer[int32](25)
	zeroRatio := helpers.GetPointer[int32](0)
	context := helpers.GetPointer[ngfAPI.TraceContext](ngfAPI.TraceContextExtract)
	spanName := helpers.GetPointer("my-span")

	tests := []struct {
		name           string
		policy         policies.Policy
		globalSettings *policies.GlobalSettings
		expStrings     []string
	}{
		{
			name: "strategy set to default ratio",
			policy: &ngfAPI.ObservabilityPolicy{
				Spec: ngfAPI.ObservabilityPolicySpec{
					Tracing: &ngfAPI.Tracing{
						Strategy: ngfAPI.TraceStrategyRatio,
					},
				},
			},
			expStrings: []string{
				"otel_trace on;",
			},
		},
		{
			name: "strategy set to custom ratio",
			policy: &ngfAPI.ObservabilityPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-policy",
					Namespace: "test-namespace",
				},
				Spec: ngfAPI.ObservabilityPolicySpec{
					Tracing: &ngfAPI.Tracing{
						Strategy: ngfAPI.TraceStrategyRatio,
						Ratio:    ratio,
					},
				},
			},
			expStrings: []string{
				"otel_trace $otel_ratio_25;",
			},
		},
		{
			name: "strategy set to zero ratio",
			policy: &ngfAPI.ObservabilityPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-policy",
					Namespace: "test-namespace",
				},
				Spec: ngfAPI.ObservabilityPolicySpec{
					Tracing: &ngfAPI.Tracing{
						Strategy: ngfAPI.TraceStrategyRatio,
						Ratio:    zeroRatio,
					},
				},
			},
			expStrings: []string{
				"otel_trace off;",
			},
		},
		{
			name: "strategy set to parent",
			policy: &ngfAPI.ObservabilityPolicy{
				Spec: ngfAPI.ObservabilityPolicySpec{
					Tracing: &ngfAPI.Tracing{
						Strategy: ngfAPI.TraceStrategyParent,
					},
				},
			},
			expStrings: []string{
				"otel_trace $otel_parent_sampled;",
			},
		},
		{
			name: "context is set",
			policy: &ngfAPI.ObservabilityPolicy{
				Spec: ngfAPI.ObservabilityPolicySpec{
					Tracing: &ngfAPI.Tracing{
						Context: context,
					},
				},
			},
			expStrings: []string{
				"otel_trace off;",
				"otel_trace_context extract;",
			},
		},
		{
			name: "spanName is set",
			policy: &ngfAPI.ObservabilityPolicy{
				Spec: ngfAPI.ObservabilityPolicySpec{
					Tracing: &ngfAPI.Tracing{
						SpanName: spanName,
					},
				},
			},
			expStrings: []string{
				"otel_trace off;",
				"otel_span_name \"my-span\";",
			},
		},
		{
			name: "span attributes set",
			policy: &ngfAPI.ObservabilityPolicy{
				Spec: ngfAPI.ObservabilityPolicySpec{
					Tracing: &ngfAPI.Tracing{
						SpanAttributes: []ngfAPI.SpanAttribute{
							{Key: "test-key", Value: "test-value"},
						},
					},
				},
			},
			expStrings: []string{
				"otel_trace off;",
				"otel_span_attr \"test-key\" \"test-value\";",
			},
		},
		{
			name: "global span attributes set",
			policy: &ngfAPI.ObservabilityPolicy{
				Spec: ngfAPI.ObservabilityPolicySpec{
					Tracing: &ngfAPI.Tracing{},
				},
			},
			globalSettings: &policies.GlobalSettings{
				TracingSpanAttributes: []ngfAPI.SpanAttribute{
					{Key: "test-global-key", Value: "test-global-value"},
				},
			},
			expStrings: []string{
				"otel_trace off;",
				"otel_span_attr \"test-global-key\" \"test-global-value\";",
			},
		},
		{
			name: "all fields populated",
			policy: &ngfAPI.ObservabilityPolicy{
				Spec: ngfAPI.ObservabilityPolicySpec{
					Tracing: &ngfAPI.Tracing{
						Strategy: ngfAPI.TraceStrategyRatio,
						Context:  context,
						SpanName: spanName,
						SpanAttributes: []ngfAPI.SpanAttribute{
							{Key: "test-key", Value: "test-value"},
						},
					},
				},
			},
			globalSettings: &policies.GlobalSettings{
				TracingSpanAttributes: []ngfAPI.SpanAttribute{
					{Key: "test-global-key", Value: "test-global-value"},
				},
			},
			expStrings: []string{
				"otel_trace on;",
				"otel_trace_context extract;",
				"otel_span_name \"my-span\";",
				"otel_span_attr \"test-key\" \"test-value\";",
				"otel_span_attr \"test-global-key\" \"test-global-value\";",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			cfgString := string(observability.Generate(test.policy, test.globalSettings))

			for _, str := range test.expStrings {
				g.Expect(cfgString).To(ContainSubstring(str))
			}
		})
	}
}

func TestGeneratePanics(t *testing.T) {
	g := NewWithT(t)

	generate := func() {
		observability.Generate(&policiesfakes.FakePolicy{}, nil)
	}

	g.Expect(generate).To(Panic())
}

func TestCreateRatioVarName(t *testing.T) {
	pol := &ngfAPI.ObservabilityPolicy{
		Spec: ngfAPI.ObservabilityPolicySpec{
			Tracing: &ngfAPI.Tracing{
				Ratio: helpers.GetPointer[int32](25),
			},
		},
	}

	g := NewWithT(t)
	g.Expect(observability.CreateRatioVarName(pol)).To(Equal("$otel_ratio_25"))
}
