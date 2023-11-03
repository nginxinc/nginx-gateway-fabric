package relationship

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
)

func TestGetBackendServiceNamesFromRoute(t *testing.T) {
	getNormalRefs := func(svcName v1beta1.ObjectName) []v1beta1.HTTPBackendRef {
		return []v1beta1.HTTPBackendRef{
			{
				BackendRef: v1beta1.BackendRef{
					BackendObjectReference: v1beta1.BackendObjectReference{
						Kind:      (*v1beta1.Kind)(helpers.GetPointer("Service")),
						Name:      svcName,
						Namespace: (*v1beta1.Namespace)(helpers.GetPointer("test")),
						Port:      (*v1beta1.PortNumber)(helpers.GetPointer[int32](80)),
					},
				},
			},
		}
	}

	getModifiedRefs := func(
		svcName v1beta1.ObjectName,
		mod func([]v1beta1.HTTPBackendRef) []v1beta1.HTTPBackendRef,
	) []v1beta1.HTTPBackendRef {
		return mod(getNormalRefs(svcName))
	}

	hr := &v1beta1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{Namespace: "test"},
		Spec: v1beta1.HTTPRouteSpec{
			Rules: []v1beta1.HTTPRouteRule{
				{
					BackendRefs: getNormalRefs("svc1"),
				},
				{
					BackendRefs: getNormalRefs("svc1"), // duplicate
				},
				{
					BackendRefs: getModifiedRefs(
						"invalid-kind",
						func(refs []v1beta1.HTTPBackendRef) []v1beta1.HTTPBackendRef {
							refs[0].Kind = (*v1beta1.Kind)(helpers.GetPointer("Invalid"))
							return refs
						},
					),
				},
				{
					BackendRefs: getModifiedRefs(
						"nil-namespace",
						func(refs []v1beta1.HTTPBackendRef) []v1beta1.HTTPBackendRef {
							refs[0].Namespace = nil
							return refs
						},
					),
				},
				{
					BackendRefs: getModifiedRefs(
						"diff-namespace",
						func(refs []v1beta1.HTTPBackendRef) []v1beta1.HTTPBackendRef {
							refs[0].Namespace = (*v1beta1.Namespace)(
								helpers.GetPointer("not-test"),
							)
							return refs
						},
					),
				},
				{
					BackendRefs: nil,
				},
				{
					BackendRefs: getNormalRefs("svc2"),
				},
				{
					BackendRefs: getModifiedRefs(
						"multiple-refs",
						func(refs []v1beta1.HTTPBackendRef) []v1beta1.HTTPBackendRef {
							return append(refs, v1beta1.HTTPBackendRef{
								BackendRef: v1beta1.BackendRef{
									BackendObjectReference: v1beta1.BackendObjectReference{
										Kind: (*v1beta1.Kind)(
											helpers.GetPointer("Service"),
										),
										Name: "multiple-refs2",
										Namespace: (*v1beta1.Namespace)(
											helpers.GetPointer("test"),
										),
										Port: (*v1beta1.PortNumber)(
											helpers.GetPointer[int32](80),
										),
									},
								},
							})
						}),
				},
			},
		},
	}

	expNames := map[types.NamespacedName]struct{}{
		{Namespace: "test", Name: "svc1"}:               {},
		{Namespace: "test", Name: "nil-namespace"}:      {},
		{Namespace: "not-test", Name: "diff-namespace"}: {},
		{Namespace: "test", Name: "svc2"}:               {},
		{Namespace: "test", Name: "multiple-refs"}:      {},
		{Namespace: "test", Name: "multiple-refs2"}:     {},
	}

	g := NewWithT(t)
	names := getBackendServiceNamesFromRoute(hr)
	g.Expect(names).To(Equal(expNames))
}

func TestCapturerImpl_DecrementRouteCount(t *testing.T) {
	testcases := []struct {
		msg              string
		startingRefCount int
		expectedRefCount int
		exists           bool
	}{
		{
			msg:              "service does not exist in map",
			startingRefCount: 0,
			expectedRefCount: 0,
			exists:           false,
		},
		{
			msg:              "service has ref count of 1",
			startingRefCount: 1,
			expectedRefCount: 0,
			exists:           false,
		},
		{
			msg:              "service has ref count of 2",
			startingRefCount: 2,
			expectedRefCount: 1,
			exists:           true,
		},
	}

	capturer := NewCapturerImpl("") // gcName is not used in this test
	svc := types.NamespacedName{Namespace: "test", Name: "svc"}

	for _, tc := range testcases {
		g := NewWithT(t)
		if tc.startingRefCount > 0 {
			capturer.serviceRefCount[svc] = tc.startingRefCount
		}

		capturer.decrementRefCount(svc)

		count, exists := capturer.serviceRefCount[svc]
		g.Expect(exists).To(Equal(tc.exists))
		g.Expect(count).To(Equal(tc.expectedRefCount))
	}
}
