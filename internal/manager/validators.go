package manager

import (
	"errors"
	"fmt"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/reconciler"
)

// createValidator creates a reconciler.ValidatorFunc from a function that validates a resource of type R.
func createValidator[R client.Object](validate func(R) field.ErrorList) reconciler.ValidatorFunc {
	return func(obj client.Object) error {
		if obj == nil {
			panic(errors.New("obj is nil"))
		}

		r, ok := obj.(R)
		if !ok {
			panic(fmt.Errorf("obj type mismatch: got %T, expected %T", obj, r))
		}

		return validate(r).ToAggregate()
	}
}
