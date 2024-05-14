package observability

import (
	"fmt"
	"slices"

	"k8s.io/apimachinery/pkg/util/validation/field"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/kinds"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/policies"
	staticConds "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/validation"
)

// Validator validates an ObservabilityPolicy.
// Implements policies.Validator interface.
type Validator struct {
	genericValidator validation.GenericValidator
}

// NewValidator returns a new instance of Validator.
func NewValidator(genericValidator validation.GenericValidator) *Validator {
	return &Validator{genericValidator: genericValidator}
}

// Validate validates the spec of an ObservabilityPolicy.
func (v *Validator) Validate(
	policy policies.Policy,
	globalSettings *policies.GlobalPolicySettings,
) []conditions.Condition {
	obs, ok := policy.(*ngfAPI.ObservabilityPolicy)
	if !ok {
		panic(fmt.Sprintf("expected ObservabilityPolicy, got: %T", policy))
	}

	if globalSettings == nil || !globalSettings.NginxProxyValid {
		return []conditions.Condition{staticConds.NewPolicyNotAcceptedNginxProxyNotSet()}
	}

	if err := validateTargetRefs(obs.Spec.TargetRefs); err != nil {
		return []conditions.Condition{staticConds.NewPolicyInvalid(err.Error())}
	}

	if err := v.validateSettings(obs.Spec); err != nil {
		return []conditions.Condition{staticConds.NewPolicyInvalid(err.Error())}
	}

	return nil
}

// Conflicts returns true if the two ObservabilityPolicies conflict.
func (v *Validator) Conflicts(polA, polB policies.Policy) bool {
	a, okA := polA.(*ngfAPI.ObservabilityPolicy)
	b, okB := polB.(*ngfAPI.ObservabilityPolicy)

	if !okA || !okB {
		panic(fmt.Sprintf("expected ObservabilityPolicies, got: %T, %T", polA, polB))
	}

	return a.Spec.Tracing != nil && b.Spec.Tracing != nil
}

func validateTargetRefs(refs []v1alpha2.LocalPolicyTargetReference) error {
	basePath := field.NewPath("spec").Child("targetRefs")

	for _, ref := range refs {
		if ref.Group != gatewayv1.GroupName {
			path := basePath.Child("group")

			return field.Invalid(
				path,
				ref.Group,
				fmt.Sprintf("unsupported targetRef Group %q; must be %s", ref.Group, gatewayv1.GroupName),
			)
		}

		supportedKinds := []gatewayv1.Kind{kinds.HTTPRoute, kinds.GRPCRoute}

		if !slices.Contains(supportedKinds, ref.Kind) {
			path := basePath.Child("kind")

			return field.Invalid(
				path,
				ref.Kind,
				fmt.Sprintf("unsupported targetRef Kind %q; Kind must be one of: %v", ref.Kind, supportedKinds),
			)
		}
	}

	return nil
}

func (v *Validator) validateSettings(spec ngfAPI.ObservabilityPolicySpec) error {
	var allErrs field.ErrorList
	fieldPath := field.NewPath("spec")

	if spec.Tracing != nil {
		tracePath := fieldPath.Child("tracing")

		switch spec.Tracing.Strategy {
		case ngfAPI.TraceStrategyRatio, ngfAPI.TraceStrategyParent:
		default:
			allErrs = append(
				allErrs,
				field.NotSupported(
					tracePath.Child("strategy"),
					spec.Tracing.Strategy,
					[]string{
						string(ngfAPI.TraceStrategyRatio),
						string(ngfAPI.TraceStrategyParent),
					}),
			)
		}

		if spec.Tracing.Context != nil {
			switch *spec.Tracing.Context {
			case ngfAPI.TraceContextExtract,
				ngfAPI.TraceContextInject,
				ngfAPI.TraceContextPropagate,
				ngfAPI.TraceContextIgnore:
			default:
				allErrs = append(
					allErrs,
					field.NotSupported(
						tracePath.Child("context"),
						spec.Tracing.Strategy,
						[]string{
							string(ngfAPI.TraceContextExtract),
							string(ngfAPI.TraceContextInject),
							string(ngfAPI.TraceContextPropagate),
							string(ngfAPI.TraceContextIgnore),
						}),
				)
			}
		}

		if spec.Tracing.SpanName != nil {
			if err := v.genericValidator.ValidateEscapedStringNoVarExpansion(*spec.Tracing.SpanName); err != nil {
				allErrs = append(allErrs, field.Invalid(tracePath.Child("spanName"), *spec.Tracing.SpanName, err.Error()))
			}
		}

		if spec.Tracing.SpanAttributes != nil {
			spanAttrPath := tracePath.Child("spanAttributes")
			for _, spanAttr := range spec.Tracing.SpanAttributes {
				if err := v.genericValidator.ValidateEscapedStringNoVarExpansion(spanAttr.Key); err != nil {
					allErrs = append(allErrs, field.Invalid(spanAttrPath.Child("key"), spanAttr.Key, err.Error()))
				}

				if err := v.genericValidator.ValidateEscapedStringNoVarExpansion(spanAttr.Value); err != nil {
					allErrs = append(allErrs, field.Invalid(spanAttrPath.Child("value"), spanAttr.Value, err.Error()))
				}
			}
		}
	}

	return allErrs.ToAggregate()
}
