package graph

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
)

func TestBuildGatewayClass(t *testing.T) {
	const controllerName = "my.controller"

	validGC := &v1beta1.GatewayClass{
		Spec: v1beta1.GatewayClassSpec{
			ControllerName: "my.controller",
		},
	}
	invalidGC := &v1beta1.GatewayClass{
		Spec: v1beta1.GatewayClassSpec{
			ControllerName: "wrong.controller",
		},
	}

	tests := []struct {
		gc       *v1beta1.GatewayClass
		expected *GatewayClass
		msg      string
	}{
		{
			gc:       nil,
			expected: nil,
			msg:      "no gatewayclass",
		},
		{
			gc: validGC,
			expected: &GatewayClass{
				Source:   validGC,
				Valid:    true,
				ErrorMsg: "",
			},
			msg: "valid gatewayclass",
		},
		{
			gc: invalidGC,
			expected: &GatewayClass{
				Source:   invalidGC,
				Valid:    false,
				ErrorMsg: "Spec.ControllerName must be my.controller got wrong.controller",
			},
			msg: "invalid gatewayclass",
		},
	}

	for _, test := range tests {
		result := buildGatewayClass(test.gc, controllerName)
		if diff := cmp.Diff(test.expected, result); diff != "" {
			t.Errorf("buildGatewayClass() '%s' mismatch (-want +got):\n%s", test.msg, diff)
		}
	}
}

func TestValidateGatewayClass(t *testing.T) {
	gc := &v1beta1.GatewayClass{
		Spec: v1beta1.GatewayClassSpec{
			ControllerName: "test.controller",
		},
	}

	err := validateGatewayClass(gc, "test.controller")
	if err != nil {
		t.Errorf("validateGatewayClass() returned unexpected error %v", err)
	}

	err = validateGatewayClass(gc, "unmatched.controller")
	if err == nil {
		t.Errorf("validateGatewayClass() didn't return an error")
	}
}
