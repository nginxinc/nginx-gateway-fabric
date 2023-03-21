package graph

import (
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/conditions"
)

// GatewayClass represents the GatewayClass resource.
type GatewayClass struct {
	// Source is the source resource.
	Source *v1beta1.GatewayClass
	// Conditions include Conditions for the GatewayClass.
	Conditions []conditions.Condition
	// Valid shows whether the GatewayClass is valid.
	Valid bool
}

func gatewayClassBelongsToController(gc *v1beta1.GatewayClass, controllerName string) bool {
	// if GatewayClass doesn't exist, we assume it belongs to the controller
	if gc == nil {
		return true
	}

	return string(gc.Spec.ControllerName) == controllerName
}

func buildGatewayClass(gc *v1beta1.GatewayClass) *GatewayClass {
	if gc == nil {
		return nil
	}

	var conds []conditions.Condition

	valErr := validateGatewayClass(gc)
	if valErr != nil {
		conds = append(conds, conditions.NewGatewayClassInvalidParameters(valErr.Error()))
	}

	return &GatewayClass{
		Source:     gc,
		Valid:      valErr == nil,
		Conditions: conds,
	}
}

func validateGatewayClass(gc *v1beta1.GatewayClass) error {
	if gc.Spec.ParametersRef != nil {
		path := field.NewPath("spec").Child("parametersRef")
		return field.Forbidden(path, "parametersRef is not supported")
	}

	return nil
}
