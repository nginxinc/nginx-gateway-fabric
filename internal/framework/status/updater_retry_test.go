package status_test

import (
	"context"
	"errors"
	"testing"

	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/controller/controllerfakes"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/status"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/status/statusfakes"
)

func TestNewRetryUpdateFunc(t *testing.T) {
	tests := []struct {
		getReturns         error
		updateReturns      error
		name               string
		expConditionPassed bool
	}{
		{
			getReturns:         errors.New("failed to get resource"),
			updateReturns:      nil,
			name:               "get fails",
			expConditionPassed: false,
		},
		{
			getReturns:         apierrors.NewNotFound(schema.GroupResource{}, "not found"),
			updateReturns:      nil,
			name:               "get fails and apierrors is not found",
			expConditionPassed: true,
		},
		{
			getReturns:         nil,
			updateReturns:      errors.New("failed to update resource"),
			name:               "update fails",
			expConditionPassed: false,
		},
		{
			getReturns:         nil,
			updateReturns:      nil,
			name:               "nothing fails",
			expConditionPassed: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)
			fakeStatusUpdater := &statusfakes.FakeK8sUpdater{}
			fakeGetter := &controllerfakes.FakeGetter{}
			fakeStatusUpdater.UpdateReturns(test.updateReturns)
			fakeGetter.GetReturns(test.getReturns)
			f := status.NewRetryUpdateFunc(
				fakeGetter,
				fakeStatusUpdater,
				types.NamespacedName{},
				&v1beta1.GatewayClass{},
				zap.New(),
				func(client.Object) {})
			conditionPassed, err := f(context.Background())

			// For now, the function should always return nil
			g.Expect(err).ToNot(HaveOccurred())
			if test.expConditionPassed {
				g.Expect(conditionPassed).To(BeTrue())
			} else {
				g.Expect(conditionPassed).To(BeFalse())
			}
		})
	}
}
