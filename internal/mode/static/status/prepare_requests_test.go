package status

import (
	"context"
	"errors"
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	v1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	statusFramework "github.com/nginxinc/nginx-gateway-fabric/internal/framework/status"
	staticConds "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/graph"
)

func createK8sClientFor(resourceType client.Object) client.Client {
	scheme := runtime.NewScheme()

	// for simplicity, we add all used schemes here
	utilruntime.Must(v1.AddToScheme(scheme))
	utilruntime.Must(v1alpha2.AddToScheme(scheme))
	utilruntime.Must(ngfAPI.AddToScheme(scheme))

	k8sClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(
			resourceType,
		).
		Build()

	return k8sClient
}

const gatewayCtlrName = "controller"

var (
	gwNsName       = types.NamespacedName{Namespace: "test", Name: "gateway"}
	transitionTime = helpers.PrepareTimeForFakeClient(metav1.Now())

	invalidRouteCondition = conditions.Condition{
		Type:   "TestInvalidRoute",
		Status: metav1.ConditionTrue,
	}
	invalidAttachmentCondition = conditions.Condition{
		Type:   "TestInvalidAttachment",
		Status: metav1.ConditionTrue,
	}

	commonRouteSpecValid = v1.CommonRouteSpec{
		ParentRefs: []v1.ParentReference{
			{
				SectionName: helpers.GetPointer[v1.SectionName]("listener-80-1"),
			},
			{
				SectionName: helpers.GetPointer[v1.SectionName]("listener-80-2"),
			},
		},
	}

	commonRouteSpecInvalid = v1.CommonRouteSpec{
		ParentRefs: []v1.ParentReference{
			{
				SectionName: helpers.GetPointer[v1.SectionName]("listener-80-1"),
			},
		},
	}

	parentRefsValid = []graph.ParentRef{
		{
			Idx:     0,
			Gateway: gwNsName,
			Attachment: &graph.ParentRefAttachmentStatus{
				Attached: true,
			},
		},
		{
			Idx:     1,
			Gateway: gwNsName,
			Attachment: &graph.ParentRefAttachmentStatus{
				Attached:        false,
				FailedCondition: invalidAttachmentCondition,
			},
		},
	}

	parentRefsInvalid = []graph.ParentRef{
		{
			Idx:        0,
			Gateway:    gwNsName,
			Attachment: nil,
		},
	}

	routeStatusValid = v1.RouteStatus{
		Parents: []v1.RouteParentStatus{
			{
				ParentRef: v1.ParentReference{
					Namespace:   helpers.GetPointer(v1.Namespace(gwNsName.Namespace)),
					Name:        v1.ObjectName(gwNsName.Name),
					SectionName: helpers.GetPointer[v1.SectionName]("listener-80-1"),
				},
				ControllerName: gatewayCtlrName,
				Conditions: []metav1.Condition{
					{
						Type:               string(v1.RouteConditionAccepted),
						Status:             metav1.ConditionTrue,
						ObservedGeneration: 3,
						LastTransitionTime: transitionTime,
						Reason:             string(v1.RouteReasonAccepted),
						Message:            "The route is accepted",
					},
					{
						Type:               string(v1.RouteConditionResolvedRefs),
						Status:             metav1.ConditionTrue,
						ObservedGeneration: 3,
						LastTransitionTime: transitionTime,
						Reason:             string(v1.RouteReasonResolvedRefs),
						Message:            "All references are resolved",
					},
				},
			},
			{
				ParentRef: v1.ParentReference{
					Namespace:   helpers.GetPointer(v1.Namespace(gwNsName.Namespace)),
					Name:        v1.ObjectName(gwNsName.Name),
					SectionName: helpers.GetPointer[v1.SectionName]("listener-80-2"),
				},
				ControllerName: gatewayCtlrName,
				Conditions: []metav1.Condition{
					{
						Type:               string(v1.RouteConditionAccepted),
						Status:             metav1.ConditionTrue,
						ObservedGeneration: 3,
						LastTransitionTime: transitionTime,
						Reason:             string(v1.RouteReasonAccepted),
						Message:            "The route is accepted",
					},
					{
						Type:               string(v1.RouteConditionResolvedRefs),
						Status:             metav1.ConditionTrue,
						ObservedGeneration: 3,
						LastTransitionTime: transitionTime,
						Reason:             string(v1.RouteReasonResolvedRefs),
						Message:            "All references are resolved",
					},
					{
						Type:               invalidAttachmentCondition.Type,
						Status:             metav1.ConditionTrue,
						ObservedGeneration: 3,
						LastTransitionTime: transitionTime,
					},
				},
			},
		},
	}

	routeStatusInvalid = v1.RouteStatus{
		Parents: []v1.RouteParentStatus{
			{
				ParentRef: v1.ParentReference{
					Namespace:   helpers.GetPointer(v1.Namespace(gwNsName.Namespace)),
					Name:        v1.ObjectName(gwNsName.Name),
					SectionName: helpers.GetPointer[v1.SectionName]("listener-80-1"),
				},
				ControllerName: gatewayCtlrName,
				Conditions: []metav1.Condition{
					{
						Type:               string(v1.RouteConditionAccepted),
						Status:             metav1.ConditionTrue,
						ObservedGeneration: 3,
						LastTransitionTime: transitionTime,
						Reason:             string(v1.RouteReasonAccepted),
						Message:            "The route is accepted",
					},
					{
						Type:               string(v1.RouteConditionResolvedRefs),
						Status:             metav1.ConditionTrue,
						ObservedGeneration: 3,
						LastTransitionTime: transitionTime,
						Reason:             string(v1.RouteReasonResolvedRefs),
						Message:            "All references are resolved",
					},
					{
						Type:               invalidRouteCondition.Type,
						Status:             metav1.ConditionTrue,
						ObservedGeneration: 3,
						LastTransitionTime: transitionTime,
					},
				},
			},
		},
	}
)

func TestBuildHTTPRouteStatuses(t *testing.T) {
	routes := map[graph.RouteKey]*graph.L7Route{
		{NamespacedName: types.NamespacedName{Namespace: "test", Name: "hr-valid"}}: {
			Valid: true,
			Source: &v1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:  "test",
					Name:       "hr-valid",
					Generation: 3,
				},
				Spec: v1.HTTPRouteSpec{
					CommonRouteSpec: commonRouteSpecValid,
				},
			},
			ParentRefs:    parentRefsValid,
			RouteType:     graph.RouteTypeHTTP,
			SrcParentRefs: commonRouteSpecValid.ParentRefs,
		},
		{NamespacedName: types.NamespacedName{Namespace: "test", Name: "hr-invalid"}}: {
			Valid:      false,
			Conditions: []conditions.Condition{invalidRouteCondition},
			Source: &v1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:  "test",
					Name:       "hr-invalid",
					Generation: 3,
				},
				Spec: v1.HTTPRouteSpec{
					CommonRouteSpec: commonRouteSpecInvalid,
				},
			},
			ParentRefs:    parentRefsInvalid,
			RouteType:     graph.RouteTypeHTTP,
			SrcParentRefs: commonRouteSpecInvalid.ParentRefs,
		},
	}

	expectedStatuses := map[types.NamespacedName]v1.HTTPRouteStatus{
		{Namespace: "test", Name: "hr-valid"}: {
			RouteStatus: routeStatusValid,
		},
		{Namespace: "test", Name: "hr-invalid"}: {
			RouteStatus: routeStatusInvalid,
		},
	}

	g := NewWithT(t)

	k8sClient := createK8sClientFor(&v1.HTTPRoute{})

	for _, r := range routes {
		err := k8sClient.Create(context.Background(), r.Source)
		g.Expect(err).ToNot(HaveOccurred())
	}

	updater := statusFramework.NewUpdater(k8sClient, zap.New())

	reqs := PrepareRouteRequests(routes, transitionTime, NginxReloadResult{}, gatewayCtlrName)

	updater.Update(context.Background(), reqs...)

	g.Expect(reqs).To(HaveLen(len(expectedStatuses)))

	for nsname, expected := range expectedStatuses {
		var hr v1.HTTPRoute

		err := k8sClient.Get(context.Background(), nsname, &hr)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(helpers.Diff(expected, hr.Status)).To(BeEmpty())
	}
}

func TestBuildGRPCRouteStatuses(t *testing.T) {
	routes := map[graph.RouteKey]*graph.L7Route{
		{NamespacedName: types.NamespacedName{Namespace: "test", Name: "gr-valid"}}: {
			Valid: true,
			Source: &v1alpha2.GRPCRoute{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:  "test",
					Name:       "gr-valid",
					Generation: 3,
				},
				Spec: v1alpha2.GRPCRouteSpec{
					CommonRouteSpec: commonRouteSpecValid,
				},
			},
			ParentRefs:    parentRefsValid,
			RouteType:     graph.RouteTypeGRPC,
			SrcParentRefs: commonRouteSpecValid.ParentRefs,
		},
		{NamespacedName: types.NamespacedName{Namespace: "test", Name: "gr-invalid"}}: {
			Valid:      false,
			Conditions: []conditions.Condition{invalidRouteCondition},
			Source: &v1alpha2.GRPCRoute{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:  "test",
					Name:       "gr-invalid",
					Generation: 3,
				},
				Spec: v1alpha2.GRPCRouteSpec{
					CommonRouteSpec: commonRouteSpecInvalid,
				},
			},
			ParentRefs:    parentRefsInvalid,
			RouteType:     graph.RouteTypeGRPC,
			SrcParentRefs: commonRouteSpecInvalid.ParentRefs,
		},
	}

	expectedStatuses := map[types.NamespacedName]v1alpha2.GRPCRouteStatus{
		{Namespace: "test", Name: "gr-valid"}: {
			RouteStatus: routeStatusValid,
		},
		{Namespace: "test", Name: "gr-invalid"}: {
			RouteStatus: routeStatusInvalid,
		},
	}

	g := NewWithT(t)

	k8sClient := createK8sClientFor(&v1alpha2.GRPCRoute{})

	for _, r := range routes {
		err := k8sClient.Create(context.Background(), r.Source)
		g.Expect(err).ToNot(HaveOccurred())
	}

	updater := statusFramework.NewUpdater(k8sClient, zap.New())

	reqs := PrepareRouteRequests(routes, transitionTime, NginxReloadResult{}, gatewayCtlrName)

	updater.Update(context.Background(), reqs...)

	g.Expect(reqs).To(HaveLen(len(expectedStatuses)))

	for nsname, expected := range expectedStatuses {
		var hr v1alpha2.GRPCRoute

		err := k8sClient.Get(context.Background(), nsname, &hr)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(helpers.Diff(expected, hr.Status)).To(BeEmpty())
	}
}

func TestBuildRouteStatusesNginxErr(t *testing.T) {
	const gatewayCtlrName = "controller"

	gwNsName := types.NamespacedName{Namespace: "test", Name: "gateway"}
	routeKey := graph.RouteKey{NamespacedName: types.NamespacedName{Namespace: "test", Name: "hr-valid"}}

	routes := map[graph.RouteKey]*graph.L7Route{
		routeKey: {
			Valid:     true,
			RouteType: graph.RouteTypeHTTP,
			Source: &v1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:  routeKey.NamespacedName.Namespace,
					Name:       routeKey.NamespacedName.Name,
					Generation: 3,
				},
				Spec: v1.HTTPRouteSpec{
					CommonRouteSpec: commonRouteSpecValid,
				},
			},
			ParentRefs: []graph.ParentRef{
				{
					Idx:     0,
					Gateway: gwNsName,
					Attachment: &graph.ParentRefAttachmentStatus{
						Attached: true,
					},
				},
			},
			SrcParentRefs: commonRouteSpecValid.ParentRefs,
		},
	}

	transitionTime := helpers.PrepareTimeForFakeClient(metav1.Now())

	expectedStatus := v1.HTTPRouteStatus{
		RouteStatus: v1.RouteStatus{
			Parents: []v1.RouteParentStatus{
				{
					ParentRef: v1.ParentReference{
						Namespace:   helpers.GetPointer(v1.Namespace(gwNsName.Namespace)),
						Name:        v1.ObjectName(gwNsName.Name),
						SectionName: helpers.GetPointer[v1.SectionName]("listener-80-1"),
					},
					ControllerName: gatewayCtlrName,
					Conditions: []metav1.Condition{
						{
							Type:               string(v1.RouteConditionResolvedRefs),
							Status:             metav1.ConditionTrue,
							ObservedGeneration: 3,
							LastTransitionTime: transitionTime,
							Reason:             string(v1.RouteReasonResolvedRefs),
							Message:            "All references are resolved",
						},
						{
							Type:               string(v1.RouteConditionAccepted),
							Status:             metav1.ConditionFalse,
							ObservedGeneration: 3,
							LastTransitionTime: transitionTime,
							Reason:             string(staticConds.RouteReasonGatewayNotProgrammed),
							Message:            staticConds.RouteMessageFailedNginxReload,
						},
					},
				},
			},
		},
	}

	g := NewWithT(t)

	k8sClient := createK8sClientFor(&v1.HTTPRoute{})

	for _, r := range routes {
		err := k8sClient.Create(context.Background(), r.Source)
		g.Expect(err).ToNot(HaveOccurred())
	}

	updater := statusFramework.NewUpdater(k8sClient, zap.New())

	reqs := PrepareRouteRequests(
		routes,
		transitionTime,
		NginxReloadResult{Error: errors.New("test error")},
		gatewayCtlrName,
	)

	g.Expect(reqs).To(HaveLen(1))

	updater.Update(context.Background(), reqs...)

	var hr v1.HTTPRoute

	err := k8sClient.Get(context.Background(), routeKey.NamespacedName, &hr)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(helpers.Diff(expectedStatus, hr.Status)).To(BeEmpty())
}

func TestBuildGatewayClassStatuses(t *testing.T) {
	transitionTime := helpers.PrepareTimeForFakeClient(metav1.Now())

	tests := []struct {
		gc             *graph.GatewayClass
		ignoredClasses map[types.NamespacedName]*v1.GatewayClass
		expected       map[types.NamespacedName]v1.GatewayClassStatus
		name           string
	}{
		{
			name:     "nil gatewayclass and no ignored gatewayclasses",
			expected: map[types.NamespacedName]v1.GatewayClassStatus{},
		},
		{
			name: "nil gatewayclass and ignored gatewayclasses",
			ignoredClasses: map[types.NamespacedName]*v1.GatewayClass{
				{Name: "ignored-1"}: {
					ObjectMeta: metav1.ObjectMeta{
						Name:       "ignored-1",
						Generation: 1,
					},
				},
				{Name: "ignored-2"}: {
					ObjectMeta: metav1.ObjectMeta{
						Name:       "ignored-2",
						Generation: 2,
					},
				},
			},
			expected: map[types.NamespacedName]v1.GatewayClassStatus{
				{Name: "ignored-1"}: {
					Conditions: []metav1.Condition{
						{
							Type:               string(v1.GatewayClassConditionStatusAccepted),
							Status:             metav1.ConditionFalse,
							ObservedGeneration: 1,
							LastTransitionTime: transitionTime,
							Reason:             string(conditions.GatewayClassReasonGatewayClassConflict),
							Message:            conditions.GatewayClassMessageGatewayClassConflict,
						},
					},
				},
				{Name: "ignored-2"}: {
					Conditions: []metav1.Condition{
						{
							Type:               string(v1.GatewayClassConditionStatusAccepted),
							Status:             metav1.ConditionFalse,
							ObservedGeneration: 2,
							LastTransitionTime: transitionTime,
							Reason:             string(conditions.GatewayClassReasonGatewayClassConflict),
							Message:            conditions.GatewayClassMessageGatewayClassConflict,
						},
					},
				},
			},
		},
		{
			name: "valid gatewayclass",
			gc: &graph.GatewayClass{
				Source: &v1.GatewayClass{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "valid-gc",
						Generation: 1,
					},
				},
			},
			expected: map[types.NamespacedName]v1.GatewayClassStatus{
				{Name: "valid-gc"}: {
					Conditions: []metav1.Condition{
						{
							Type:               string(v1.GatewayClassConditionStatusAccepted),
							Status:             metav1.ConditionTrue,
							ObservedGeneration: 1,
							LastTransitionTime: transitionTime,
							Reason:             string(v1.GatewayClassReasonAccepted),
							Message:            "GatewayClass is accepted",
						},
						{
							Type:               string(v1.GatewayClassReasonSupportedVersion),
							Status:             metav1.ConditionTrue,
							ObservedGeneration: 1,
							LastTransitionTime: transitionTime,
							Reason:             string(v1.GatewayClassReasonSupportedVersion),
							Message:            "Gateway API CRD versions are supported",
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			k8sClient := createK8sClientFor(&v1.GatewayClass{})

			var expectedTotalReqs int

			if test.gc != nil {
				err := k8sClient.Create(context.Background(), test.gc.Source)
				g.Expect(err).ToNot(HaveOccurred())
				expectedTotalReqs++
			}

			for _, gc := range test.ignoredClasses {
				err := k8sClient.Create(context.Background(), gc)
				g.Expect(err).ToNot(HaveOccurred())
				expectedTotalReqs++
			}

			updater := statusFramework.NewUpdater(k8sClient, zap.New())

			reqs := PrepareGatewayClassRequests(test.gc, test.ignoredClasses, transitionTime)

			g.Expect(reqs).To(HaveLen(expectedTotalReqs))

			updater.Update(context.Background(), reqs...)

			for nsname, expected := range test.expected {
				var gc v1.GatewayClass

				err := k8sClient.Get(context.Background(), nsname, &gc)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(helpers.Diff(expected, gc.Status)).To(BeEmpty())
			}
		})
	}
}

func TestBuildGatewayStatuses(t *testing.T) {
	createGateway := func() *v1.Gateway {
		return &v1.Gateway{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:  "test",
				Name:       "gateway",
				Generation: 2,
			},
		}
	}

	transitionTime := helpers.PrepareTimeForFakeClient(metav1.Now())

	validListenerConditions := []metav1.Condition{
		{
			Type:               string(v1.ListenerConditionAccepted),
			Status:             metav1.ConditionTrue,
			ObservedGeneration: 2,
			LastTransitionTime: transitionTime,
			Reason:             string(v1.ListenerReasonAccepted),
			Message:            "Listener is accepted",
		},
		{
			Type:               string(v1.ListenerConditionProgrammed),
			Status:             metav1.ConditionTrue,
			ObservedGeneration: 2,
			LastTransitionTime: transitionTime,
			Reason:             string(v1.ListenerReasonProgrammed),
			Message:            "Listener is programmed",
		},
		{
			Type:               string(v1.ListenerConditionResolvedRefs),
			Status:             metav1.ConditionTrue,
			ObservedGeneration: 2,
			LastTransitionTime: transitionTime,
			Reason:             string(v1.ListenerReasonResolvedRefs),
			Message:            "All references are resolved",
		},
		{
			Type:               string(v1.ListenerConditionConflicted),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: 2,
			LastTransitionTime: transitionTime,
			Reason:             string(v1.ListenerReasonNoConflicts),
			Message:            "No conflicts",
		},
	}

	addr := []v1.GatewayStatusAddress{
		{
			Type:  helpers.GetPointer(v1.IPAddressType),
			Value: "1.2.3.4",
		},
	}

	tests := []struct {
		nginxReloadRes  NginxReloadResult
		gateway         *graph.Gateway
		ignoredGateways map[types.NamespacedName]*v1.Gateway
		expected        map[types.NamespacedName]v1.GatewayStatus
		name            string
	}{
		{
			name:     "nil gateway and no ignored gateways",
			expected: map[types.NamespacedName]v1.GatewayStatus{},
		},
		{
			name: "nil gateway and ignored gateways",
			ignoredGateways: map[types.NamespacedName]*v1.Gateway{
				{Namespace: "test", Name: "ignored-1"}: {
					ObjectMeta: metav1.ObjectMeta{
						Name:       "ignored-1",
						Namespace:  "test",
						Generation: 1,
					},
				},
				{Namespace: "test", Name: "ignored-2"}: {
					ObjectMeta: metav1.ObjectMeta{
						Name:       "ignored-2",
						Namespace:  "test",
						Generation: 2,
					},
				},
			},
			expected: map[types.NamespacedName]v1.GatewayStatus{
				{Namespace: "test", Name: "ignored-1"}: {
					Conditions: []metav1.Condition{
						{
							Type:               string(v1.GatewayConditionAccepted),
							Status:             metav1.ConditionFalse,
							ObservedGeneration: 1,
							LastTransitionTime: transitionTime,
							Reason:             string(staticConds.GatewayReasonGatewayConflict),
							Message:            staticConds.GatewayMessageGatewayConflict,
						},
						{
							Type:               string(v1.GatewayConditionProgrammed),
							Status:             metav1.ConditionFalse,
							ObservedGeneration: 1,
							LastTransitionTime: transitionTime,
							Reason:             string(staticConds.GatewayReasonGatewayConflict),
							Message:            staticConds.GatewayMessageGatewayConflict,
						},
					},
				},
				{Namespace: "test", Name: "ignored-2"}: {
					Conditions: []metav1.Condition{
						{
							Type:               string(v1.GatewayConditionAccepted),
							Status:             metav1.ConditionFalse,
							ObservedGeneration: 2,
							LastTransitionTime: transitionTime,
							Reason:             string(staticConds.GatewayReasonGatewayConflict),
							Message:            staticConds.GatewayMessageGatewayConflict,
						},
						{
							Type:               string(v1.GatewayConditionProgrammed),
							Status:             metav1.ConditionFalse,
							ObservedGeneration: 2,
							LastTransitionTime: transitionTime,
							Reason:             string(staticConds.GatewayReasonGatewayConflict),
							Message:            staticConds.GatewayMessageGatewayConflict,
						},
					},
				},
			},
		},
		{
			name: "valid gateway; all valid listeners",
			gateway: &graph.Gateway{
				Source: createGateway(),
				Listeners: []*graph.Listener{
					{
						Name:  "listener-valid-1",
						Valid: true,
						Routes: map[graph.RouteKey]*graph.L7Route{
							{NamespacedName: types.NamespacedName{Namespace: "test", Name: "hr-1"}}: {},
						},
					},
					{
						Name:  "listener-valid-2",
						Valid: true,
						Routes: map[graph.RouteKey]*graph.L7Route{
							{NamespacedName: types.NamespacedName{Namespace: "test", Name: "hr-1"}}: {},
						},
					},
				},
				Valid: true,
			},
			expected: map[types.NamespacedName]v1.GatewayStatus{
				{Namespace: "test", Name: "gateway"}: {
					Addresses: addr,
					Conditions: []metav1.Condition{
						{
							Type:               string(v1.GatewayConditionAccepted),
							Status:             metav1.ConditionTrue,
							ObservedGeneration: 2,
							LastTransitionTime: transitionTime,
							Reason:             string(v1.GatewayReasonAccepted),
							Message:            "Gateway is accepted",
						},
						{
							Type:               string(v1.GatewayConditionProgrammed),
							Status:             metav1.ConditionTrue,
							ObservedGeneration: 2,
							LastTransitionTime: transitionTime,
							Reason:             string(v1.GatewayReasonProgrammed),
							Message:            "Gateway is programmed",
						},
					},
					Listeners: []v1.ListenerStatus{
						{
							Name:           "listener-valid-1",
							AttachedRoutes: 1,
							Conditions:     validListenerConditions,
						},
						{
							Name:           "listener-valid-2",
							AttachedRoutes: 1,
							Conditions:     validListenerConditions,
						},
					},
				},
			},
		},
		{
			name: "valid gateway; some valid listeners",
			gateway: &graph.Gateway{
				Source: createGateway(),
				Listeners: []*graph.Listener{
					{
						Name:  "listener-valid",
						Valid: true,
						Routes: map[graph.RouteKey]*graph.L7Route{
							{NamespacedName: types.NamespacedName{Namespace: "test", Name: "hr-1"}}: {},
						},
					},
					{
						Name:       "listener-invalid",
						Valid:      false,
						Conditions: staticConds.NewListenerUnsupportedValue("unsupported value"),
					},
				},
				Valid: true,
			},
			expected: map[types.NamespacedName]v1.GatewayStatus{
				{Namespace: "test", Name: "gateway"}: {
					Addresses: addr,
					Conditions: []metav1.Condition{
						{
							Type:               string(v1.GatewayConditionProgrammed),
							Status:             metav1.ConditionTrue,
							ObservedGeneration: 2,
							LastTransitionTime: transitionTime,
							Reason:             string(v1.GatewayReasonProgrammed),
							Message:            "Gateway is programmed",
						},
						{
							// is it a bug?
							Type:               string(v1.GatewayReasonAccepted),
							Status:             metav1.ConditionTrue,
							ObservedGeneration: 2,
							LastTransitionTime: transitionTime,
							Reason:             string(v1.GatewayReasonListenersNotValid),
							Message:            "Gateway has at least one valid listener",
						},
					},
					Listeners: []v1.ListenerStatus{
						{
							Name:           "listener-valid",
							AttachedRoutes: 1,
							Conditions:     validListenerConditions,
						},
						{
							Name:           "listener-invalid",
							AttachedRoutes: 0,
							Conditions: []metav1.Condition{
								{
									Type:               string(v1.ListenerConditionAccepted),
									Status:             metav1.ConditionFalse,
									ObservedGeneration: 2,
									LastTransitionTime: transitionTime,
									Reason:             string(staticConds.ListenerReasonUnsupportedValue),
									Message:            "unsupported value",
								},
								{
									Type:               string(v1.ListenerConditionProgrammed),
									Status:             metav1.ConditionFalse,
									ObservedGeneration: 2,
									LastTransitionTime: transitionTime,
									Reason:             string(v1.ListenerReasonInvalid),
									Message:            "unsupported value",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "valid gateway; no valid listeners",
			gateway: &graph.Gateway{
				Source: createGateway(),
				Listeners: []*graph.Listener{
					{
						Name:       "listener-invalid-1",
						Valid:      false,
						Conditions: staticConds.NewListenerUnsupportedProtocol("unsupported protocol"),
					},
					{
						Name:       "listener-invalid-2",
						Valid:      false,
						Conditions: staticConds.NewListenerUnsupportedValue("unsupported value"),
					},
				},
				Valid: true,
			},
			expected: map[types.NamespacedName]v1.GatewayStatus{
				{Namespace: "test", Name: "gateway"}: {
					Addresses: addr,
					Conditions: []metav1.Condition{
						{
							Type:               string(v1.GatewayReasonAccepted),
							Status:             metav1.ConditionFalse,
							ObservedGeneration: 2,
							LastTransitionTime: transitionTime,
							Reason:             string(v1.GatewayReasonListenersNotValid),
							Message:            "Gateway has no valid listeners",
						},
						{
							Type:               string(v1.GatewayConditionProgrammed),
							Status:             metav1.ConditionFalse,
							ObservedGeneration: 2,
							LastTransitionTime: transitionTime,
							Reason:             string(v1.GatewayReasonInvalid),
							Message:            "Gateway has no valid listeners",
						},
					},
					Listeners: []v1.ListenerStatus{
						{
							Name:           "listener-invalid-1",
							AttachedRoutes: 0,
							Conditions: []metav1.Condition{
								{
									Type:               string(v1.ListenerConditionAccepted),
									Status:             metav1.ConditionFalse,
									ObservedGeneration: 2,
									LastTransitionTime: transitionTime,
									Reason:             string(v1.ListenerReasonUnsupportedProtocol),
									Message:            "unsupported protocol",
								},
								{
									Type:               string(v1.ListenerConditionProgrammed),
									Status:             metav1.ConditionFalse,
									ObservedGeneration: 2,
									LastTransitionTime: transitionTime,
									Reason:             string(v1.ListenerReasonInvalid),
									Message:            "unsupported protocol",
								},
							},
						},
						{
							Name:           "listener-invalid-2",
							AttachedRoutes: 0,
							Conditions: []metav1.Condition{
								{
									Type:               string(v1.ListenerConditionAccepted),
									Status:             metav1.ConditionFalse,
									ObservedGeneration: 2,
									LastTransitionTime: transitionTime,
									Reason:             string(staticConds.ListenerReasonUnsupportedValue),
									Message:            "unsupported value",
								},
								{
									Type:               string(v1.ListenerConditionProgrammed),
									Status:             metav1.ConditionFalse,
									ObservedGeneration: 2,
									LastTransitionTime: transitionTime,
									Reason:             string(v1.ListenerReasonInvalid),
									Message:            "unsupported value",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "invalid gateway",
			gateway: &graph.Gateway{
				Source:     createGateway(),
				Valid:      false,
				Conditions: staticConds.NewGatewayInvalid("no gateway class"),
			},
			expected: map[types.NamespacedName]v1.GatewayStatus{
				{Namespace: "test", Name: "gateway"}: {
					Conditions: []metav1.Condition{
						{
							Type:               string(v1.GatewayConditionAccepted),
							Status:             metav1.ConditionFalse,
							ObservedGeneration: 2,
							LastTransitionTime: transitionTime,
							Reason:             string(v1.GatewayReasonInvalid),
							Message:            "no gateway class",
						},
						{
							Type:               string(v1.GatewayConditionProgrammed),
							Status:             metav1.ConditionFalse,
							ObservedGeneration: 2,
							LastTransitionTime: transitionTime,
							Reason:             string(v1.GatewayReasonInvalid),
							Message:            "no gateway class",
						},
					},
				},
			},
		},
		{
			name: "error reloading nginx; gateway/listener not programmed",
			gateway: &graph.Gateway{
				Source:     createGateway(),
				Valid:      true,
				Conditions: staticConds.NewDefaultGatewayConditions(),
				Listeners: []*graph.Listener{
					{
						Name:  "listener-valid",
						Valid: true,
						Routes: map[graph.RouteKey]*graph.L7Route{
							{NamespacedName: types.NamespacedName{Namespace: "test", Name: "hr-1"}}: {},
						},
					},
				},
			},
			expected: map[types.NamespacedName]v1.GatewayStatus{
				{Namespace: "test", Name: "gateway"}: {
					Addresses: addr,
					Conditions: []metav1.Condition{
						{
							Type:               string(v1.GatewayConditionAccepted),
							Status:             metav1.ConditionTrue,
							ObservedGeneration: 2,
							LastTransitionTime: transitionTime,
							Reason:             string(v1.GatewayReasonAccepted),
							Message:            "Gateway is accepted",
						},
						{
							Type:               string(v1.GatewayConditionProgrammed),
							Status:             metav1.ConditionFalse,
							ObservedGeneration: 2,
							LastTransitionTime: transitionTime,
							Reason:             string(v1.GatewayReasonInvalid),
							Message:            staticConds.GatewayMessageFailedNginxReload,
						},
					},
					Listeners: []v1.ListenerStatus{
						{
							Name:           "listener-valid",
							AttachedRoutes: 1,
							Conditions: []metav1.Condition{
								{
									Type:               string(v1.ListenerConditionAccepted),
									Status:             metav1.ConditionTrue,
									ObservedGeneration: 2,
									LastTransitionTime: transitionTime,
									Reason:             string(v1.ListenerReasonAccepted),
									Message:            "Listener is accepted",
								},
								{
									Type:               string(v1.ListenerConditionResolvedRefs),
									Status:             metav1.ConditionTrue,
									ObservedGeneration: 2,
									LastTransitionTime: transitionTime,
									Reason:             string(v1.ListenerReasonResolvedRefs),
									Message:            "All references are resolved",
								},
								{
									Type:               string(v1.ListenerConditionConflicted),
									Status:             metav1.ConditionFalse,
									ObservedGeneration: 2,
									LastTransitionTime: transitionTime,
									Reason:             string(v1.ListenerReasonNoConflicts),
									Message:            "No conflicts",
								},
								{
									Type:               string(v1.ListenerConditionProgrammed),
									Status:             metav1.ConditionFalse,
									ObservedGeneration: 2,
									LastTransitionTime: transitionTime,
									Reason:             string(v1.ListenerReasonInvalid),
									Message:            staticConds.ListenerMessageFailedNginxReload,
								},
							},
						},
					},
				},
			},
			nginxReloadRes: NginxReloadResult{Error: errors.New("test error")},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			k8sClient := createK8sClientFor(&v1.Gateway{})

			var expectedTotalReqs int

			if test.gateway != nil {
				test.gateway.Source.ResourceVersion = ""
				err := k8sClient.Create(context.Background(), test.gateway.Source)
				g.Expect(err).ToNot(HaveOccurred())
				expectedTotalReqs++
			}

			for _, gw := range test.ignoredGateways {
				err := k8sClient.Create(context.Background(), gw)
				g.Expect(err).ToNot(HaveOccurred())
				expectedTotalReqs++
			}

			updater := statusFramework.NewUpdater(k8sClient, zap.New())

			reqs := PrepareGatewayRequests(test.gateway, test.ignoredGateways, transitionTime, addr, test.nginxReloadRes)

			g.Expect(reqs).To(HaveLen(expectedTotalReqs))

			updater.Update(context.Background(), reqs...)

			for nsname, expected := range test.expected {
				var gw v1.Gateway

				err := k8sClient.Get(context.Background(), nsname, &gw)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(helpers.Diff(expected, gw.Status)).To(BeEmpty())
			}
		})
	}
}

func TestBuildBackendTLSPolicyStatuses(t *testing.T) {
	const gatewayCtlrName = "controller"

	transitionTime := helpers.PrepareTimeForFakeClient(metav1.Now())

	type policyCfg struct {
		Name         string
		Conditions   []conditions.Condition
		Valid        bool
		Ignored      bool
		IsReferenced bool
	}

	getBackendTLSPolicy := func(policyCfg policyCfg) *graph.BackendTLSPolicy {
		return &graph.BackendTLSPolicy{
			Source: &v1alpha2.BackendTLSPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:  "test",
					Name:       policyCfg.Name,
					Generation: 1,
				},
			},
			Valid:        policyCfg.Valid,
			Ignored:      policyCfg.Ignored,
			IsReferenced: policyCfg.IsReferenced,
			Conditions:   policyCfg.Conditions,
			Gateway:      types.NamespacedName{Name: "gateway", Namespace: "test"},
		}
	}

	attachedConds := []conditions.Condition{staticConds.NewBackendTLSPolicyAccepted()}
	invalidConds := []conditions.Condition{staticConds.NewBackendTLSPolicyInvalid("invalid backendTLSPolicy")}

	validPolicyCfg := policyCfg{
		Name:         "valid-bt",
		Valid:        true,
		IsReferenced: true,
		Conditions:   attachedConds,
	}

	invalidPolicyCfg := policyCfg{
		Name:         "invalid-bt",
		IsReferenced: true,
		Conditions:   invalidConds,
	}

	ignoredPolicyCfg := policyCfg{
		Name:         "ignored-bt",
		Ignored:      true,
		IsReferenced: true,
	}

	notReferencedPolicyCfg := policyCfg{
		Name:  "not-referenced",
		Valid: true,
	}

	tests := []struct {
		backendTLSPolicies map[types.NamespacedName]*graph.BackendTLSPolicy
		expected           map[types.NamespacedName]v1alpha2.PolicyStatus
		name               string
		expectedReqs       int
	}{
		{
			name:         "nil backendTLSPolicies",
			expectedReqs: 0,
			expected:     map[types.NamespacedName]v1alpha2.PolicyStatus{},
		},
		{
			name: "valid backendTLSPolicy",
			backendTLSPolicies: map[types.NamespacedName]*graph.BackendTLSPolicy{
				{Namespace: "test", Name: "valid-bt"}: getBackendTLSPolicy(validPolicyCfg),
			},
			expectedReqs: 1,
			expected: map[types.NamespacedName]v1alpha2.PolicyStatus{
				{Name: "valid-bt", Namespace: "test"}: {
					Ancestors: []v1alpha2.PolicyAncestorStatus{
						{
							AncestorRef: v1.ParentReference{
								Namespace: helpers.GetPointer[v1.Namespace]("test"),
								Name:      "gateway",
							},
							ControllerName: gatewayCtlrName,
							Conditions: []metav1.Condition{
								{
									Type:               string(v1alpha2.PolicyConditionAccepted),
									Status:             metav1.ConditionTrue,
									ObservedGeneration: 1,
									LastTransitionTime: transitionTime,
									Reason:             string(v1alpha2.PolicyReasonAccepted),
									Message:            "BackendTLSPolicy is accepted by the Gateway",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "invalid backendTLSPolicy",
			backendTLSPolicies: map[types.NamespacedName]*graph.BackendTLSPolicy{
				{Namespace: "test", Name: "invalid-bt"}: getBackendTLSPolicy(invalidPolicyCfg),
			},
			expectedReqs: 1,
			expected: map[types.NamespacedName]v1alpha2.PolicyStatus{
				{Name: "invalid-bt", Namespace: "test"}: {
					Ancestors: []v1alpha2.PolicyAncestorStatus{
						{
							AncestorRef: v1.ParentReference{
								Namespace: helpers.GetPointer[v1.Namespace]("test"),
								Name:      "gateway",
							},
							ControllerName: gatewayCtlrName,
							Conditions: []metav1.Condition{
								{
									Type:               string(v1alpha2.PolicyConditionAccepted),
									Status:             metav1.ConditionFalse,
									ObservedGeneration: 1,
									LastTransitionTime: transitionTime,
									Reason:             string(v1alpha2.PolicyReasonInvalid),
									Message:            "invalid backendTLSPolicy",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "ignored or not referenced backendTLSPolicies",
			backendTLSPolicies: map[types.NamespacedName]*graph.BackendTLSPolicy{
				{Namespace: "test", Name: "ignored-bt"}:     getBackendTLSPolicy(ignoredPolicyCfg),
				{Namespace: "test", Name: "not-referenced"}: getBackendTLSPolicy(notReferencedPolicyCfg),
			},
			expectedReqs: 0,
			expected: map[types.NamespacedName]v1alpha2.PolicyStatus{
				{Name: "ignored-bt", Namespace: "test"}:     {},
				{Name: "not-referenced", Namespace: "test"}: {},
			},
		},
		{
			name: "mix valid and ignored backendTLSPolicies",
			backendTLSPolicies: map[types.NamespacedName]*graph.BackendTLSPolicy{
				{Namespace: "test", Name: "ignored-bt"}: getBackendTLSPolicy(ignoredPolicyCfg),
				{Namespace: "test", Name: "valid-bt"}:   getBackendTLSPolicy(validPolicyCfg),
			},
			expectedReqs: 1,
			expected: map[types.NamespacedName]v1alpha2.PolicyStatus{
				{Name: "ignored-bt", Namespace: "test"}: {},
				{Name: "valid-bt", Namespace: "test"}: {
					Ancestors: []v1alpha2.PolicyAncestorStatus{
						{
							AncestorRef: v1.ParentReference{
								Namespace: helpers.GetPointer[v1.Namespace]("test"),
								Name:      "gateway",
							},
							ControllerName: gatewayCtlrName,
							Conditions: []metav1.Condition{
								{
									Type:               string(v1alpha2.PolicyConditionAccepted),
									Status:             metav1.ConditionTrue,
									ObservedGeneration: 1,
									LastTransitionTime: transitionTime,
									Reason:             string(v1alpha2.PolicyReasonAccepted),
									Message:            "BackendTLSPolicy is accepted by the Gateway",
								},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			k8sClient := createK8sClientFor(&v1alpha2.BackendTLSPolicy{})

			for _, pol := range test.backendTLSPolicies {
				err := k8sClient.Create(context.Background(), pol.Source)
				g.Expect(err).ToNot(HaveOccurred())
			}

			updater := statusFramework.NewUpdater(k8sClient, zap.New())

			reqs := PrepareBackendTLSPolicyRequests(test.backendTLSPolicies, transitionTime, gatewayCtlrName)

			g.Expect(reqs).To(HaveLen(test.expectedReqs))

			updater.Update(context.Background(), reqs...)

			for nsname, expected := range test.expected {
				var pol v1alpha2.BackendTLSPolicy

				err := k8sClient.Get(context.Background(), nsname, &pol)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(helpers.Diff(expected, pol.Status)).To(BeEmpty())
			}
		})
	}
}

func TestBuildNginxGatewayStatus(t *testing.T) {
	transitionTime := helpers.PrepareTimeForFakeClient(metav1.Now())

	tests := []struct {
		cpUpdateResult ControlPlaneUpdateResult
		nginxGateway   *ngfAPI.NginxGateway
		expected       *ngfAPI.NginxGatewayStatus
		name           string
	}{
		{
			name: "nil NginxGateway",
		},
		{
			name: "NginxGateway with no update error",
			nginxGateway: &ngfAPI.NginxGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "nginx-gateway",
					Namespace:  "test",
					Generation: 3,
				},
			},
			cpUpdateResult: ControlPlaneUpdateResult{},
			expected: &ngfAPI.NginxGatewayStatus{
				Conditions: []metav1.Condition{
					{
						Type:               string(ngfAPI.NginxGatewayConditionValid),
						Status:             metav1.ConditionTrue,
						ObservedGeneration: 3,
						LastTransitionTime: transitionTime,
						Reason:             string(ngfAPI.NginxGatewayReasonValid),
						Message:            "NginxGateway is valid",
					},
				},
			},
		},
		{
			name: "NginxGateway with update error",
			nginxGateway: &ngfAPI.NginxGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "nginx-gateway",
					Namespace:  "test",
					Generation: 3,
				},
			},
			cpUpdateResult: ControlPlaneUpdateResult{
				Error: errors.New("test error"),
			},
			expected: &ngfAPI.NginxGatewayStatus{
				Conditions: []metav1.Condition{
					{
						Type:               string(ngfAPI.NginxGatewayConditionValid),
						Status:             metav1.ConditionFalse,
						ObservedGeneration: 3,
						LastTransitionTime: transitionTime,
						Reason:             string(ngfAPI.NginxGatewayReasonInvalid),
						Message:            "Failed to update control plane configuration: test error",
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			k8sClient := createK8sClientFor(&ngfAPI.NginxGateway{})

			if test.nginxGateway != nil {
				err := k8sClient.Create(context.Background(), test.nginxGateway)
				g.Expect(err).ToNot(HaveOccurred())
			}

			updater := statusFramework.NewUpdater(k8sClient, zap.New())

			req := PrepareNginxGatewayStatus(test.nginxGateway, transitionTime, test.cpUpdateResult)

			if test.nginxGateway == nil {
				g.Expect(req).To(BeNil())
			} else {
				g.Expect(req).ToNot(BeNil())
				updater.Update(context.Background(), *req)

				var ngw ngfAPI.NginxGateway

				err := k8sClient.Get(context.Background(), types.NamespacedName{Namespace: "test", Name: "nginx-gateway"}, &ngw)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(helpers.Diff(*test.expected, ngw.Status)).To(BeEmpty())
			}
		})
	}
}
