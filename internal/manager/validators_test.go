package manager

import (
	"testing"

	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
)

func TestCreateTypedValidator(t *testing.T) {
	tests := []struct {
		name        string
		obj         client.Object
		errorList   field.ErrorList
		expectPanic bool
		expectErr   bool
	}{
		{
			obj:         &v1beta1.HTTPRoute{},
			errorList:   field.ErrorList{},
			expectPanic: false,
			expectErr:   false,
			name:        "no errors",
		},
		{
			obj:         &v1beta1.HTTPRoute{},
			errorList:   []*field.Error{{Detail: "test"}},
			expectPanic: false,
			expectErr:   true,
			name:        "one error",
		},
		{
			obj:         nil,
			errorList:   field.ErrorList{},
			expectPanic: true,
			expectErr:   false,
			name:        "nil object",
		},
		{
			obj:         &v1beta1.Gateway{},
			errorList:   field.ErrorList{},
			expectPanic: true,
			expectErr:   false,
			name:        "wrong object type",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			v := createValidator(createValidateHTTPRouteThatReturns(test.errorList))

			if test.expectPanic {
				g.Expect(func() { _ = v(test.obj) }).To(Panic())
				return
			}

			result := v(test.obj)

			if test.expectErr {
				g.Expect(result).ToNot(BeNil())
				return
			}

			g.Expect(result).To(BeNil())
		})
	}
}

func createValidateHTTPRouteThatReturns(errorList field.ErrorList) func(*v1beta1.HTTPRoute) field.ErrorList {
	return func(*v1beta1.HTTPRoute) field.ErrorList {
		return errorList
	}
}
