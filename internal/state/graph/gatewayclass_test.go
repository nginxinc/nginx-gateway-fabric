package graph

import (
	"testing"

	. "github.com/onsi/gomega"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/helpers"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/conditions"
)

func TestBuildGatewayClass(t *testing.T) {
	validGC := &v1beta1.GatewayClass{}

	invalidGC := &v1beta1.GatewayClass{
		Spec: v1beta1.GatewayClassSpec{
			ParametersRef: &v1beta1.ParametersReference{},
		},
	}

	tests := []struct {
		gc       *v1beta1.GatewayClass
		expected *GatewayClass
		name     string
	}{
		{
			gc: validGC,
			expected: &GatewayClass{
				Source: validGC,
				Valid:  true,
			},
			name: "valid gatewayclass",
		},
		{
			gc:       nil,
			expected: nil,
			name:     "no gatewayclass",
		},
		{
			gc: invalidGC,
			expected: &GatewayClass{
				Source: invalidGC,
				Valid:  false,
				Conditions: []conditions.Condition{
					conditions.NewGatewayClassInvalidParameters("spec.parametersRef: Forbidden: parametersRef is not supported"),
				},
			},
			name: "invalid gatewayclass",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			result := buildGatewayClass(test.gc)
			g.Expect(helpers.Diff(test.expected, result)).To(BeEmpty())
		})
	}
}

func TestGatewayClassBelongsToController(t *testing.T) {
	const controllerName = "my.controller"

	tests := []struct {
		gc       *v1beta1.GatewayClass
		name     string
		expected bool
	}{
		{
			gc: &v1beta1.GatewayClass{
				Spec: v1beta1.GatewayClassSpec{
					ControllerName: controllerName,
				},
			},
			expected: true,
			name:     "normal gatewayclass",
		},
		{
			gc:       nil,
			expected: true,
			name:     "no gatewayclass",
		},
		{
			gc: &v1beta1.GatewayClass{
				Spec: v1beta1.GatewayClassSpec{
					ControllerName: "some.controller",
				},
			},
			expected: false,
			name:     "wrong controller name",
		},
		{
			gc: &v1beta1.GatewayClass{
				Spec: v1beta1.GatewayClassSpec{
					ControllerName: "",
				},
			},
			expected: false,
			name:     "empty controller name",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			result := gatewayClassBelongsToController(test.gc, controllerName)
			g.Expect(result).To(Equal(test.expected))
		})
	}
}
