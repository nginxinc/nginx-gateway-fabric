package status_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/controller/controllerfakes"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/status"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/status/statusfakes"
)

func TestConditionWithContextFunc_GetFails(t *testing.T) {
	g := NewWithT(t)
	fakeStatusUpdater := &statusfakes.FakeStatusUpdater{}
	fakeGetter := &controllerfakes.FakeGetter{}

	fakeGetter.GetReturns(errors.New("failed to get resource"))
	f := status.ConditionWithContextFunc(
		fakeGetter,
		fakeStatusUpdater,
		types.NamespacedName{},
		&v1beta1.GatewayClass{},
		logr.New(nil),
		func(client.Object) {})
	boolean, err := f(context.Background())

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(boolean).To(BeFalse())
}

func TestConditionWithContextFunc_GetFailsIsNotFound(t *testing.T) {
	g := NewWithT(t)
	fakeStatusUpdater := &statusfakes.FakeStatusUpdater{}
	fakeGetter := &controllerfakes.FakeGetter{}

	fakeGetter.GetReturns(apierrors.NewNotFound(schema.GroupResource{}, "not found"))
	f := status.ConditionWithContextFunc(
		fakeGetter,
		fakeStatusUpdater,
		types.NamespacedName{},
		&v1beta1.GatewayClass{},
		logr.New(nil),
		func(client.Object) {})
	boolean, err := f(context.Background())

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(boolean).To(BeTrue())
}

func TestConditionWithContextFunc_UpdateFails(t *testing.T) {
	g := NewWithT(t)
	fakeStatusUpdater := &statusfakes.FakeStatusUpdater{}
	fakeGetter := &controllerfakes.FakeGetter{}

	fakeStatusUpdater.UpdateReturns(errors.New("failed to update resource"))
	f := status.ConditionWithContextFunc(
		fakeGetter,
		fakeStatusUpdater,
		types.NamespacedName{},
		&v1beta1.GatewayClass{},
		logr.New(nil),
		func(client.Object) {})
	boolean, err := f(context.Background())

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(boolean).To(BeFalse())
}

func TestConditionWithContextFunc_NothingFails(t *testing.T) {
	g := NewWithT(t)
	fakeStatusUpdater := &statusfakes.FakeStatusUpdater{}
	fakeGetter := &controllerfakes.FakeGetter{}

	f := status.ConditionWithContextFunc(
		fakeGetter,
		fakeStatusUpdater,
		types.NamespacedName{},
		&v1beta1.GatewayClass{},
		logr.New(nil),
		func(client.Object) {})
	boolean, err := f(context.Background())

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(boolean).To(BeTrue())
}
