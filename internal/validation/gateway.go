package validation

import (
	"fmt"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
)

func ValidateGateway(gw *v1alpha2.Gateway) error {
	allErrs := validateGatewaySpec(&gw.Spec, field.NewPath("spec"))

	return allErrs.ToAggregate()
}

func validateGatewaySpec(spec *v1alpha2.GatewaySpec, fieldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	// for now, it must be nginx
	if spec.GatewayClassName != "nginx" {
		allErrs = append(allErrs, field.Invalid(fieldPath.Child("gatewayClassName"), spec.GatewayClassName, "must be 'nginx'"))
	}

	// for now, Gateway must have one HTTP listener with the name "http"

	listenersPath := fieldPath.Child("listeners")

	if len(spec.Listeners) < 1 {
		allErrs = append(allErrs, field.Required(listenersPath, "must define one HTTP listener with the name 'http'"))
		return allErrs
	} else if len(spec.Listeners) > 1 {
		allErrs = append(allErrs, field.TooMany(listenersPath, len(spec.Listeners), 1))
		return allErrs
	}

	listener := spec.Listeners[0]
	idxPath := listenersPath.Index(0)

	if listener.Name != "http" {
		allErrs = append(allErrs, field.Invalid(idxPath, listener.Name, "must be 'http'"))
	}
	if listener.Protocol != v1alpha2.HTTPProtocolType {
		allErrs = append(allErrs, field.Invalid(idxPath, listener.Protocol, fmt.Sprintf("must be '%s'", v1alpha2.HTTPProtocolType)))
	}

	return allErrs
}
