package graph

import (
	"fmt"

	"sigs.k8s.io/gateway-api/apis/v1beta1"
)

// GatewayClass represents the GatewayClass resource.
type GatewayClass struct {
	// Source is the source resource.
	Source *v1beta1.GatewayClass
	// ErrorMsg explains the error when the resource is invalid.
	ErrorMsg string
	// Valid shows whether the GatewayClass is valid.
	Valid bool
}

func buildGatewayClass(gc *v1beta1.GatewayClass, controllerName string) *GatewayClass {
	if gc == nil {
		return nil
	}

	var errorMsg string

	err := validateGatewayClass(gc, controllerName)
	if err != nil {
		errorMsg = err.Error()
	}

	return &GatewayClass{
		Source:   gc,
		Valid:    err == nil,
		ErrorMsg: errorMsg,
	}
}

func validateGatewayClass(gc *v1beta1.GatewayClass, controllerName string) error {
	if string(gc.Spec.ControllerName) != controllerName {
		return fmt.Errorf("Spec.ControllerName must be %s got %s", controllerName, gc.Spec.ControllerName)
	}

	return nil
}
