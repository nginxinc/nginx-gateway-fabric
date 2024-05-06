package graph

import (
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	v1 "sigs.k8s.io/gateway-api/apis/v1"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/validation"
)

// getNginxProxy returns the NginxProxy associated with the GatewayClass (if it exists).
func getNginxProxy(
	nps map[types.NamespacedName]*ngfAPI.NginxProxy,
	gc *v1.GatewayClass,
) *ngfAPI.NginxProxy {
	if gcReferencesAnyNginxProxy(gc) {
		return nps[types.NamespacedName{Name: gc.Spec.ParametersRef.Name}]
	}

	return nil
}

// isNginxProxyReferenced returns whether or not a specific NginxProxy is referenced in the GatewayClass.
func isNginxProxyReferenced(npNSName types.NamespacedName, gc *GatewayClass) bool {
	return gc != nil && gcReferencesAnyNginxProxy(gc.Source) && gc.Source.Spec.ParametersRef.Name == npNSName.Name
}

// gcReferencesNginxProxy returns whether a GatewayClass references any NginxProxy resource.
func gcReferencesAnyNginxProxy(gc *v1.GatewayClass) bool {
	if gc != nil {
		ref := gc.Spec.ParametersRef
		return ref != nil && ref.Group == ngfAPI.GroupName && ref.Kind == v1.Kind("NginxProxy")
	}

	return false
}

// validateNginxProxy performs re-validation on string values in the case of CRD validation failure.
func validateNginxProxy(
	validator validation.GenericValidator,
	npCfg *ngfAPI.NginxProxy,
) field.ErrorList {
	var allErrs field.ErrorList
	spec := field.NewPath("spec")

	telemetry := npCfg.Spec.Telemetry
	if telemetry != nil {
		telPath := spec.Child("telemetry")
		if telemetry.ServiceName != nil {
			if err := validator.ValidateServiceName(*telemetry.ServiceName); err != nil {
				allErrs = append(allErrs, field.Invalid(telPath.Child("serviceName"), *telemetry.ServiceName, err.Error()))
			}
		}

		if telemetry.Exporter != nil {
			exp := telemetry.Exporter
			expPath := telPath.Child("exporter")

			if exp.Endpoint != "" {
				if err := validator.ValidateEndpoint(exp.Endpoint); err != nil {
					allErrs = append(allErrs, field.Invalid(expPath.Child("endpoint"), exp.Endpoint, err.Error()))
				}
			}

			if exp.Interval != nil {
				if err := validator.ValidateNginxDuration(string(*exp.Interval)); err != nil {
					allErrs = append(allErrs, field.Invalid(expPath.Child("interval"), *exp.Interval, err.Error()))
				}
			}
		}

		if telemetry.SpanAttributes != nil {
			spanAttrPath := telPath.Child("spanAttributes")
			for _, spanAttr := range telemetry.SpanAttributes {
				if err := validator.ValidateEscapedStringNoVarExpansion(spanAttr.Key); err != nil {
					allErrs = append(allErrs, field.Invalid(spanAttrPath.Child("key"), spanAttr.Key, err.Error()))
				}

				if err := validator.ValidateEscapedStringNoVarExpansion(spanAttr.Value); err != nil {
					allErrs = append(allErrs, field.Invalid(spanAttrPath.Child("value"), spanAttr.Value, err.Error()))
				}
			}
		}
	}

	return allErrs
}
