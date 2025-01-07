package observability

import (
	"k8s.io/apimachinery/pkg/util/validation/field"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	ngfAPI "github.com/nginx/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginx/nginx-gateway-fabric/internal/framework/conditions"
	"github.com/nginx/nginx-gateway-fabric/internal/framework/helpers"
	"github.com/nginx/nginx-gateway-fabric/internal/framework/kinds"
	"github.com/nginx/nginx-gateway-fabric/internal/mode/static/nginx/config/policies"
	staticConds "github.com/nginx/nginx-gateway-fabric/internal/mode/static/state/conditions"
	"github.com/nginx/nginx-gateway-fabric/internal/mode/static/state/validation"
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
	globalSettings *policies.GlobalSettings,
) []conditions.Condition {
	obs := helpers.MustCastObject[*ngfAPI.ObservabilityPolicy](policy)

	if globalSettings == nil || !globalSettings.NginxProxyValid {
		return []conditions.Condition{
			staticConds.NewPolicyNotAcceptedNginxProxyNotSet(staticConds.PolicyMessageNginxProxyInvalid),
		}
	}

	if !globalSettings.TelemetryEnabled {
		return []conditions.Condition{
			staticConds.NewPolicyNotAcceptedNginxProxyNotSet(staticConds.PolicyMessageTelemetryNotEnabled),
		}
	}

	targetRefPath := field.NewPath("spec").Child("targetRefs")
	supportedKinds := []gatewayv1.Kind{kinds.HTTPRoute, kinds.GRPCRoute}
	supportedGroups := []gatewayv1.Group{gatewayv1.GroupName}

	for _, ref := range obs.Spec.TargetRefs {
		if err := policies.ValidateTargetRef(ref, targetRefPath, supportedGroups, supportedKinds); err != nil {
			return []conditions.Condition{staticConds.NewPolicyInvalid(err.Error())}
		}
	}

	if err := v.validateSettings(obs.Spec); err != nil {
		return []conditions.Condition{staticConds.NewPolicyInvalid(err.Error())}
	}

	return nil
}

// Conflicts returns true if the two ObservabilityPolicies conflict.
func (v *Validator) Conflicts(polA, polB policies.Policy) bool {
	a := helpers.MustCastObject[*ngfAPI.ObservabilityPolicy](polA)
	b := helpers.MustCastObject[*ngfAPI.ObservabilityPolicy](polB)

	return a.Spec.Tracing != nil && b.Spec.Tracing != nil
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
						spec.Tracing.Context,
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
				allErrs = append(
					allErrs,
					field.Invalid(tracePath.Child("spanName"), *spec.Tracing.SpanName, err.Error()),
				)
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
