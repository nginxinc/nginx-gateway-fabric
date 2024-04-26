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
func isNginxProxyReferenced(np *ngfAPI.NginxProxy, gc *GatewayClass) bool {
	return np != nil && gc != nil &&
		gcReferencesAnyNginxProxy(gc.Source) && gc.Source.Spec.ParametersRef.Name == np.Name
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

	validateStringValue := func(
		value,
		valueName string,
		path *field.Path,
		validator validation.GenericValidator,
	) *field.Error {
		if err := validator.ValidateEscapedStringNoVarExpansion(value); err != nil {
			return field.Invalid(path.Child(valueName), value, err.Error())
		}

		return nil
	}

	telemetry := npCfg.Spec.Telemetry
	if telemetry != nil {
		telPath := spec.Child("telemetry")
		if telemetry.ServiceName != nil {
			if err := validateStringValue(*telemetry.ServiceName, "serviceName", telPath, validator); err != nil {
				allErrs = append(allErrs, err)
			}
		}

		if telemetry.Exporter != nil {
			exp := telemetry.Exporter
			expPath := telPath.Child("exporter")

			if exp.Endpoint != "" {
				if err := validateStringValue(exp.Endpoint, "endpoint", expPath, validator); err != nil {
					allErrs = append(allErrs, err)
				}
			}

			if exp.Interval != nil {
				if err := validateStringValue(string(*exp.Interval), "interval", expPath, validator); err != nil {
					allErrs = append(allErrs, err)
				}
			}
		}

		if telemetry.SpanAttributes != nil {
			spanAttrPath := telPath.Child("spanAttributes")
			for _, spanAttr := range telemetry.SpanAttributes {
				if err := validateStringValue(spanAttr.Key, "key", spanAttrPath, validator); err != nil {
					allErrs = append(allErrs, err)
				}

				if err := validateStringValue(spanAttr.Value, "value", spanAttrPath, validator); err != nil {
					allErrs = append(allErrs, err)
				}
			}
		}
	}

	return allErrs
}
