package validation

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
)

func ValidateGatewayClass(gc *v1alpha2.GatewayClass) error {
	allErrs := validateGatewayClassSpec(&gc.Spec, field.NewPath("spec"))

	return allErrs.ToAggregate()
}

const controllerName = "k8s-gateway.nginx.org/nginx-gateway/gateway"

func validateGatewayClassSpec(spec *v1alpha2.GatewayClassSpec, fieldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if spec.ControllerName != controllerName {
		allErrs = append(allErrs, field.Invalid(fieldPath.Child("controllerName"), spec.ControllerName,
			fmt.Sprintf("must be '%s'", controllerName)))
	}

	return allErrs
}