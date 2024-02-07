package status

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
)

func TestNewNginxGatewayStatusSetter(t *testing.T) {
	tests := []struct {
		name         string
		status       ngfAPI.NginxGatewayStatus
		newStatus    NginxGatewayStatus
		expStatusSet bool
	}{
		{
			name:         "NginxGateway has no status",
			expStatusSet: true,
			newStatus: NginxGatewayStatus{
				Conditions: []conditions.Condition{{Message: "new condition"}},
			},
		},
		{
			name:         "NginxGateway has old status",
			expStatusSet: true,
			newStatus: NginxGatewayStatus{
				Conditions: []conditions.Condition{{Message: "new condition"}},
			},
			status: ngfAPI.NginxGatewayStatus{
				Conditions: []v1.Condition{{Message: "old condition"}},
			},
		},
		{
			name:         "NginxGateway has same status",
			expStatusSet: false,
			newStatus: NginxGatewayStatus{
				Conditions: []conditions.Condition{{Message: "same condition"}},
			},
			status: ngfAPI.NginxGatewayStatus{
				Conditions: []v1.Condition{{Message: "same condition"}},
			},
		},
	}

	clock := &RealClock{}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			setter := newNginxGatewayStatusSetter(clock, test.newStatus)

			statusSet := setter(&ngfAPI.NginxGateway{Status: test.status})
			g.Expect(statusSet).To(Equal(test.expStatusSet))
		})
	}
}

func TestNewGatewayClassStatusSetter(t *testing.T) {
	tests := []struct {
		name         string
		status       gatewayv1.GatewayClassStatus
		newStatus    GatewayClassStatus
		expStatusSet bool
	}{
		{
			name: "GatewayClass has no status",
			newStatus: GatewayClassStatus{
				Conditions: []conditions.Condition{{Message: "new condition"}},
			},
			expStatusSet: true,
		},
		{
			name: "GatewayClass has old status",
			newStatus: GatewayClassStatus{
				Conditions: []conditions.Condition{{Message: "new condition"}},
			},
			status: gatewayv1.GatewayClassStatus{
				Conditions: []v1.Condition{{Message: "old condition"}},
			},
			expStatusSet: true,
		},
		{
			name: "GatewayClass has same status",
			newStatus: GatewayClassStatus{
				Conditions: []conditions.Condition{{Message: "same condition"}},
			},
			status: gatewayv1.GatewayClassStatus{
				Conditions: []v1.Condition{{Message: "same condition"}},
			},
			expStatusSet: false,
		},
	}

	clock := &RealClock{}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			setter := newGatewayClassStatusSetter(clock, test.newStatus)
			statusSet := setter(&gatewayv1.GatewayClass{Status: test.status})
			g.Expect(statusSet).To(Equal(test.expStatusSet))
		})
	}
}

func TestNewGatewayStatusSetter(t *testing.T) {
	expAddress := gatewayv1.GatewayStatusAddress{
		Type:  helpers.GetPointer(gatewayv1.IPAddressType),
		Value: "10.0.0.0",
	}

	tests := []struct {
		name         string
		status       gatewayv1.GatewayStatus
		newStatus    GatewayStatus
		expStatusSet bool
	}{
		{
			name: "Gateway has no status",
			newStatus: GatewayStatus{
				Conditions: []conditions.Condition{{Message: "new condition"}},
				Addresses:  []gatewayv1.GatewayStatusAddress{expAddress},
			},
			expStatusSet: true,
		},
		{
			name: "Gateway has old status",
			newStatus: GatewayStatus{
				Conditions: []conditions.Condition{{Message: "new condition"}},
				Addresses:  []gatewayv1.GatewayStatusAddress{expAddress},
			},
			status: gatewayv1.GatewayStatus{
				Conditions: []v1.Condition{{Message: "old condition"}},
				Addresses:  []gatewayv1.GatewayStatusAddress{expAddress},
			},
			expStatusSet: true,
		},
		{
			name: "Gateway has same status",
			newStatus: GatewayStatus{
				Conditions: []conditions.Condition{{Message: "same condition"}},
				Addresses:  []gatewayv1.GatewayStatusAddress{expAddress},
			},
			status: gatewayv1.GatewayStatus{
				Conditions: []v1.Condition{{Message: "same condition"}},
				Addresses:  []gatewayv1.GatewayStatusAddress{expAddress},
			},
			expStatusSet: false,
		},
	}

	clock := &RealClock{}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			setter := newGatewayStatusSetter(clock, test.newStatus)

			statusSet := setter(&gatewayv1.Gateway{Status: test.status})
			g.Expect(statusSet).To(Equal(test.expStatusSet))
		})
	}
}

func TestNewHTTPRouteStatusSetter(t *testing.T) {
	controllerName := "controller"

	tests := []struct {
		name         string
		status       gatewayv1.HTTPRouteStatus
		newStatus    HTTPRouteStatus
		expStatusSet bool
	}{
		{
			name: "HTTPRoute has no status",
			newStatus: HTTPRouteStatus{
				ParentStatuses: []ParentStatus{
					{
						Conditions: []conditions.Condition{{Message: "new condition"}},
					},
				},
			},
			expStatusSet: true,
		},
		{
			name: "HTTPRoute has old status",
			newStatus: HTTPRouteStatus{
				ParentStatuses: []ParentStatus{
					{
						Conditions: []conditions.Condition{{Message: "new condition"}},
					},
				},
			},
			status: gatewayv1.HTTPRouteStatus{
				RouteStatus: gatewayv1.RouteStatus{
					Parents: []gatewayv1.RouteParentStatus{
						{
							ParentRef:      gatewayv1.ParentReference{},
							ControllerName: gatewayv1.GatewayController(controllerName),
							Conditions:     []v1.Condition{{Message: "old condition"}},
						},
					},
				},
			},
			expStatusSet: true,
		},
		{
			name: "HTTPRoute has same status",
			newStatus: HTTPRouteStatus{
				ParentStatuses: []ParentStatus{
					{
						Conditions: []conditions.Condition{{Message: "same condition"}},
					},
				},
			},
			status: gatewayv1.HTTPRouteStatus{
				RouteStatus: gatewayv1.RouteStatus{
					Parents: []gatewayv1.RouteParentStatus{
						{
							ParentRef:      gatewayv1.ParentReference{},
							ControllerName: gatewayv1.GatewayController(controllerName),
							Conditions:     []v1.Condition{{Message: "same condition"}},
						},
					},
				},
			},
			expStatusSet: false,
		},
	}

	clock := &RealClock{}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			setter := newHTTPRouteStatusSetter(controllerName, clock, test.newStatus)

			statusSet := setter(&gatewayv1.HTTPRoute{Status: test.status})
			g.Expect(statusSet).To(Equal(test.expStatusSet))
		})
	}
}

func TestGWStatusEqual(t *testing.T) {
	getDefaultStatus := func() gatewayv1.GatewayStatus {
		return gatewayv1.GatewayStatus{
			Addresses: []gatewayv1.GatewayStatusAddress{
				{
					Type:  helpers.GetPointer(gatewayv1.IPAddressType),
					Value: "10.0.0.0",
				},
				{
					Type:  helpers.GetPointer(gatewayv1.IPAddressType),
					Value: "11.0.0.0",
				},
			},
			Conditions: []v1.Condition{
				{
					Type: "type", /* conditions are covered by another test*/
				},
			},
			Listeners: []gatewayv1.ListenerStatus{
				{
					Name: "listener1",
					SupportedKinds: []gatewayv1.RouteGroupKind{
						{
							Group: helpers.GetPointer[gatewayv1.Group](gatewayv1.GroupName),
							Kind:  "HTTPRoute",
						},
						{
							Group: helpers.GetPointer[gatewayv1.Group](gatewayv1.GroupName),
							Kind:  "TCPRoute",
						},
					},
					AttachedRoutes: 1,
					Conditions: []v1.Condition{
						{
							Type: "type", /* conditions are covered by another test*/
						},
					},
				},
				{
					Name: "listener2",
					SupportedKinds: []gatewayv1.RouteGroupKind{
						{
							Group: helpers.GetPointer[gatewayv1.Group](gatewayv1.GroupName),
							Kind:  "HTTPRoute",
						},
					},
					AttachedRoutes: 1,
					Conditions: []v1.Condition{
						{
							Type: "type", /* conditions are covered by another test*/
						},
					},
				},
				{
					Name: "listener3",
					SupportedKinds: []gatewayv1.RouteGroupKind{
						{
							Group: helpers.GetPointer[gatewayv1.Group](gatewayv1.GroupName),
							Kind:  "HTTPRoute",
						},
					},
					AttachedRoutes: 1,
					Conditions: []v1.Condition{
						{
							Type: "type", /* conditions are covered by another test*/
						},
					},
				},
			},
		}
	}

	getModifiedStatus := func(mod func(gatewayv1.GatewayStatus) gatewayv1.GatewayStatus) gatewayv1.GatewayStatus {
		return mod(getDefaultStatus())
	}

	tests := []struct {
		name       string
		prevStatus gatewayv1.GatewayStatus
		curStatus  gatewayv1.GatewayStatus
		expEqual   bool
	}{
		{
			name:       "different number of addresses",
			prevStatus: getDefaultStatus(),
			curStatus: getModifiedStatus(func(status gatewayv1.GatewayStatus) gatewayv1.GatewayStatus {
				status.Addresses = status.Addresses[:1]
				return status
			}),
			expEqual: false,
		},
		{
			name:       "different address type",
			prevStatus: getDefaultStatus(),
			curStatus: getModifiedStatus(func(status gatewayv1.GatewayStatus) gatewayv1.GatewayStatus {
				status.Addresses[1].Type = helpers.GetPointer(gatewayv1.HostnameAddressType)
				return status
			}),
			expEqual: false,
		},
		{
			name:       "different address value",
			prevStatus: getDefaultStatus(),
			curStatus: getModifiedStatus(func(status gatewayv1.GatewayStatus) gatewayv1.GatewayStatus {
				status.Addresses[0].Value = "12.0.0.0"
				return status
			}),
			expEqual: false,
		},
		{
			name:       "different conditions",
			prevStatus: getDefaultStatus(),
			curStatus: getModifiedStatus(func(status gatewayv1.GatewayStatus) gatewayv1.GatewayStatus {
				status.Conditions[0].Type = "different"
				return status
			}),
			expEqual: false,
		},
		{
			name:       "different number of listener statuses",
			prevStatus: getDefaultStatus(),
			curStatus: getModifiedStatus(func(status gatewayv1.GatewayStatus) gatewayv1.GatewayStatus {
				status.Listeners = status.Listeners[:2]
				return status
			}),
			expEqual: false,
		},
		{
			name:       "different listener status name",
			prevStatus: getDefaultStatus(),
			curStatus: getModifiedStatus(func(status gatewayv1.GatewayStatus) gatewayv1.GatewayStatus {
				status.Listeners[2].Name = "different"
				return status
			}),
			expEqual: false,
		},
		{
			name:       "different listener status attached routes",
			prevStatus: getDefaultStatus(),
			curStatus: getModifiedStatus(func(status gatewayv1.GatewayStatus) gatewayv1.GatewayStatus {
				status.Listeners[1].AttachedRoutes++
				return status
			}),
			expEqual: false,
		},
		{
			name:       "different listener status conditions",
			prevStatus: getDefaultStatus(),
			curStatus: getModifiedStatus(func(status gatewayv1.GatewayStatus) gatewayv1.GatewayStatus {
				status.Listeners[0].Conditions[0].Type = "different"
				return status
			}),
			expEqual: false,
		},
		{
			name:       "different listener status supported kinds (different number)",
			prevStatus: getDefaultStatus(),
			curStatus: getModifiedStatus(func(status gatewayv1.GatewayStatus) gatewayv1.GatewayStatus {
				status.Listeners[0].SupportedKinds = status.Listeners[0].SupportedKinds[:1]
				return status
			}),
			expEqual: false,
		},
		{
			name:       "different listener status supported kinds (different kind)",
			prevStatus: getDefaultStatus(),
			curStatus: getModifiedStatus(func(status gatewayv1.GatewayStatus) gatewayv1.GatewayStatus {
				status.Listeners[1].SupportedKinds[0].Kind = "TCPRoute"
				return status
			}),
			expEqual: false,
		},
		{
			name:       "different listener status supported kinds (different group)",
			prevStatus: getDefaultStatus(),
			curStatus: getModifiedStatus(func(status gatewayv1.GatewayStatus) gatewayv1.GatewayStatus {
				status.Listeners[1].SupportedKinds[0].Group = helpers.GetPointer[gatewayv1.Group]("different")
				return status
			}),
			expEqual: false,
		},
		{
			name:       "equal",
			prevStatus: getDefaultStatus(),
			curStatus:  getDefaultStatus(),
			expEqual:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			equal := gwStatusEqual(test.prevStatus, test.curStatus)
			g.Expect(equal).To(Equal(test.expEqual))
		})
	}
}

func TestHRStatusEqual(t *testing.T) {
	testConds := []v1.Condition{
		{
			Type: "type", /* conditions are covered by another test*/
		},
	}

	previousStatus := gatewayv1.HTTPRouteStatus{
		RouteStatus: gatewayv1.RouteStatus{
			Parents: []gatewayv1.RouteParentStatus{
				{
					ParentRef: gatewayv1.ParentReference{
						Namespace:   helpers.GetPointer[gatewayv1.Namespace]("test"),
						Name:        "our-parent",
						SectionName: helpers.GetPointer[gatewayv1.SectionName]("section1"),
					},
					ControllerName: "ours",
					Conditions:     testConds,
				},
				{
					ParentRef: gatewayv1.ParentReference{
						Namespace:   helpers.GetPointer[gatewayv1.Namespace]("test"),
						Name:        "not-our-parent",
						SectionName: helpers.GetPointer[gatewayv1.SectionName]("section1"),
					},
					ControllerName: "not-ours",
					Conditions:     testConds,
				},
				{
					ParentRef: gatewayv1.ParentReference{
						Namespace:   helpers.GetPointer[gatewayv1.Namespace]("test"),
						Name:        "our-parent",
						SectionName: helpers.GetPointer[gatewayv1.SectionName]("section2"),
					},
					ControllerName: "ours",
					Conditions:     testConds,
				},
				{
					ParentRef: gatewayv1.ParentReference{
						Namespace:   helpers.GetPointer[gatewayv1.Namespace]("test"),
						Name:        "not-our-parent",
						SectionName: helpers.GetPointer[gatewayv1.SectionName]("section2"),
					},
					ControllerName: "not-ours",
					Conditions:     testConds,
				},
			},
		},
	}

	getDefaultStatus := func() gatewayv1.HTTPRouteStatus {
		return gatewayv1.HTTPRouteStatus{
			RouteStatus: gatewayv1.RouteStatus{
				Parents: []gatewayv1.RouteParentStatus{
					{
						ParentRef: gatewayv1.ParentReference{
							Namespace:   helpers.GetPointer[gatewayv1.Namespace]("test"),
							Name:        "our-parent",
							SectionName: helpers.GetPointer[gatewayv1.SectionName]("section1"),
						},
						ControllerName: "ours",
						Conditions:     testConds,
					},
					{
						ParentRef: gatewayv1.ParentReference{
							Namespace:   helpers.GetPointer[gatewayv1.Namespace]("test"),
							Name:        "our-parent",
							SectionName: helpers.GetPointer[gatewayv1.SectionName]("section2"),
						},
						ControllerName: "ours",
						Conditions:     testConds,
					},
				},
			},
		}
	}

	newParentStatus := gatewayv1.RouteParentStatus{
		ParentRef: gatewayv1.ParentReference{
			Namespace:   helpers.GetPointer[gatewayv1.Namespace]("test"),
			Name:        "our-parent",
			SectionName: helpers.GetPointer[gatewayv1.SectionName]("section3"),
		},
		ControllerName: "ours",
		Conditions:     testConds,
	}

	getModifiedStatus := func(
		mod func(status gatewayv1.HTTPRouteStatus) gatewayv1.HTTPRouteStatus,
	) gatewayv1.HTTPRouteStatus {
		return mod(getDefaultStatus())
	}

	tests := []struct {
		name       string
		prevStatus gatewayv1.HTTPRouteStatus
		curStatus  gatewayv1.HTTPRouteStatus
		expEqual   bool
	}{
		{
			name:       "stale status",
			prevStatus: previousStatus,
			curStatus: getModifiedStatus(func(status gatewayv1.HTTPRouteStatus) gatewayv1.HTTPRouteStatus {
				// remove last parent status
				status.Parents = status.Parents[:1]
				return status
			}),
			expEqual: false,
		},
		{
			name:       "new status",
			prevStatus: previousStatus,
			curStatus: getModifiedStatus(func(status gatewayv1.HTTPRouteStatus) gatewayv1.HTTPRouteStatus {
				// add another parent status
				status.Parents = append(status.Parents, newParentStatus)
				return status
			}),
			expEqual: false,
		},
		{
			name:       "equal",
			prevStatus: previousStatus,
			curStatus:  getDefaultStatus(),
			expEqual:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)
			equal := hrStatusEqual("ours", test.prevStatus, test.curStatus)
			g.Expect(equal).To(Equal(test.expEqual))
		})
	}
}

func TestRouteParentStatusEqual(t *testing.T) {
	getDefaultStatus := func() gatewayv1.RouteParentStatus {
		return gatewayv1.RouteParentStatus{
			ParentRef: gatewayv1.ParentReference{
				Namespace:   helpers.GetPointer[gatewayv1.Namespace]("test"),
				Name:        "parent",
				SectionName: helpers.GetPointer[gatewayv1.SectionName]("section"),
			},
			ControllerName: "controller",
			Conditions: []v1.Condition{
				{
					Type: "type", /* conditions are covered by another test*/
				},
			},
		}
	}

	getModifiedStatus := func(
		mod func(gatewayv1.RouteParentStatus) gatewayv1.RouteParentStatus,
	) gatewayv1.RouteParentStatus {
		return mod(getDefaultStatus())
	}

	tests := []struct {
		name     string
		p1       gatewayv1.RouteParentStatus
		p2       gatewayv1.RouteParentStatus
		expEqual bool
	}{
		{
			name: "different controller name",
			p1:   getDefaultStatus(),
			p2: getModifiedStatus(func(status gatewayv1.RouteParentStatus) gatewayv1.RouteParentStatus {
				status.ControllerName = "different"
				return status
			}),
			expEqual: false,
		},
		{
			name: "different parentRef name",
			p1:   getDefaultStatus(),
			p2: getModifiedStatus(func(status gatewayv1.RouteParentStatus) gatewayv1.RouteParentStatus {
				status.ParentRef.Name = "different"
				return status
			}),
			expEqual: false,
		},
		{
			name: "different parentRef namespace",
			p1:   getDefaultStatus(),
			p2: getModifiedStatus(func(status gatewayv1.RouteParentStatus) gatewayv1.RouteParentStatus {
				status.ParentRef.Namespace = helpers.GetPointer[gatewayv1.Namespace]("different")
				return status
			}),
			expEqual: false,
		},
		{
			name: "different parentRef section name",
			p1:   getDefaultStatus(),
			p2: getModifiedStatus(func(status gatewayv1.RouteParentStatus) gatewayv1.RouteParentStatus {
				status.ParentRef.SectionName = helpers.GetPointer[gatewayv1.SectionName]("different")
				return status
			}),
			expEqual: false,
		},
		{
			name: "different conditions",
			p1:   getDefaultStatus(),
			p2: getModifiedStatus(func(status gatewayv1.RouteParentStatus) gatewayv1.RouteParentStatus {
				status.Conditions[0].Type = "different"
				return status
			}),
			expEqual: false,
		},
		{
			name:     "equal",
			p1:       getDefaultStatus(),
			p2:       getDefaultStatus(),
			expEqual: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)
			equal := routeParentStatusEqual(test.p1, test.p2)
			g.Expect(equal).To(Equal(test.expEqual))
		})
	}
}

func TestConditionsEqual(t *testing.T) {
	getDefaultConds := func() []v1.Condition {
		return []v1.Condition{
			{
				Type:               "type1",
				Status:             "status1",
				ObservedGeneration: 1,
				LastTransitionTime: v1.Time{Time: time.Now()},
				Reason:             "reason1",
				Message:            "message1",
			},
			{
				Type:               "type2",
				Status:             "status2",
				ObservedGeneration: 1,
				LastTransitionTime: v1.Time{Time: time.Now()},
				Reason:             "reason2",
				Message:            "message2",
			},
			{
				Type:               "type3",
				Status:             "status3",
				ObservedGeneration: 1,
				LastTransitionTime: v1.Time{Time: time.Now()},
				Reason:             "reason3",
				Message:            "message3",
			},
		}
	}

	getModifiedConds := func(mod func([]v1.Condition) []v1.Condition) []v1.Condition {
		return mod(getDefaultConds())
	}

	tests := []struct {
		name      string
		prevConds []v1.Condition
		curConds  []v1.Condition
		expEqual  bool
	}{
		{
			name:      "different observed gen",
			prevConds: getDefaultConds(),
			curConds: getModifiedConds(func(conds []v1.Condition) []v1.Condition {
				conds[2].ObservedGeneration++
				return conds
			}),
			expEqual: false,
		},
		{
			name:      "different status",
			prevConds: getDefaultConds(),
			curConds: getModifiedConds(func(conds []v1.Condition) []v1.Condition {
				conds[1].Status = "different"
				return conds
			}),
			expEqual: false,
		},
		{
			name:      "different type",
			prevConds: getDefaultConds(),
			curConds: getModifiedConds(func(conds []v1.Condition) []v1.Condition {
				conds[0].Type = "different"
				return conds
			}),
			expEqual: false,
		},
		{
			name:      "different message",
			prevConds: getDefaultConds(),
			curConds: getModifiedConds(func(conds []v1.Condition) []v1.Condition {
				conds[2].Message = "different"
				return conds
			}),
			expEqual: false,
		},
		{
			name:      "different reason",
			prevConds: getDefaultConds(),
			curConds: getModifiedConds(func(conds []v1.Condition) []v1.Condition {
				conds[1].Reason = "different"
				return conds
			}),
			expEqual: false,
		},
		{
			name:      "different number of conditions",
			prevConds: getDefaultConds(),
			curConds: getModifiedConds(func(conds []v1.Condition) []v1.Condition {
				return conds[:2]
			}),
			expEqual: false,
		},
		{
			name:      "equal",
			prevConds: getDefaultConds(),
			curConds:  getDefaultConds(),
			expEqual:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)
			equal := conditionsEqual(test.prevConds, test.curConds)
			g.Expect(equal).To(Equal(test.expEqual))
		})
	}
}

func TestEqualPointers(t *testing.T) {
	tests := []struct {
		p1       *string
		p2       *string
		name     string
		expEqual bool
	}{
		{
			name:     "first pointer nil; second has non-empty value",
			p1:       nil,
			p2:       helpers.GetPointer("test"),
			expEqual: false,
		},
		{
			name:     "second pointer nil; first has non-empty value",
			p1:       helpers.GetPointer("test"),
			p2:       nil,
			expEqual: false,
		},
		{
			name:     "different values",
			p1:       helpers.GetPointer("test"),
			p2:       helpers.GetPointer("different"),
			expEqual: false,
		},
		{
			name:     "both pointers nil",
			p1:       nil,
			p2:       nil,
			expEqual: true,
		},
		{
			name:     "first pointer nil; second empty",
			p1:       nil,
			p2:       helpers.GetPointer(""),
			expEqual: true,
		},
		{
			name:     "second pointer nil; first empty",
			p1:       helpers.GetPointer(""),
			p2:       nil,
			expEqual: true,
		},
		{
			name:     "same value",
			p1:       helpers.GetPointer("test"),
			p2:       helpers.GetPointer("test"),
			expEqual: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			val := equalPointers(test.p1, test.p2)
			g.Expect(val).To(Equal(test.expEqual))
		})
	}
}

func TestBtpStatusEqual(t *testing.T) {
	getPolicyStatus := func(ancestorName, ancestorNs, ctlrName string) gatewayv1alpha2.PolicyStatus {
		return gatewayv1alpha2.PolicyStatus{
			Ancestors: []gatewayv1alpha2.PolicyAncestorStatus{
				{
					AncestorRef: gatewayv1.ParentReference{
						Namespace: helpers.GetPointer[gatewayv1.Namespace]((gatewayv1.Namespace)(ancestorNs)),
						Name:      gatewayv1alpha2.ObjectName(ancestorName),
					},
					ControllerName: gatewayv1alpha2.GatewayController(ctlrName),
					Conditions:     []v1.Condition{{Type: "otherType", Status: "otherStatus"}},
				},
			},
		}
	}
	prevMultiple := getPolicyStatus("ancestor1", "ns1", "ctlr1")
	prevMultiple.Ancestors = append(prevMultiple.Ancestors, getPolicyStatus("ancestor2", "ns2", "ctlr2").Ancestors...)

	currMultiple := getPolicyStatus("ancestor1", "ns1", "ctlr1")
	currMultiple.Ancestors = append(currMultiple.Ancestors, getPolicyStatus("ancestor3", "ns3", "ctlr2").Ancestors...)

	tests := []struct {
		name           string
		controllerName string
		previous       gatewayv1alpha2.PolicyStatus
		current        gatewayv1alpha2.PolicyStatus
		expEqual       bool
	}{
		{
			name:           "status equal",
			previous:       getPolicyStatus("ancestor1", "ns1", "ctlr1"),
			current:        getPolicyStatus("ancestor1", "ns1", "ctlr1"),
			controllerName: "ctlr1",
			expEqual:       true,
		},
		{
			name:           "status not equal, different ancestor name",
			previous:       getPolicyStatus("ancestor1", "ns1", "ctlr1"),
			current:        getPolicyStatus("ancestor2", "ns1", "ctlr1"),
			controllerName: "ctlr1",
			expEqual:       false,
		},
		{
			name:           "status not equal, different ancestor namespace",
			previous:       getPolicyStatus("ancestor1", "ns1", "ctlr1"),
			current:        getPolicyStatus("ancestor1", "ns2", "ctlr1"),
			controllerName: "ctlr1",
			expEqual:       false,
		},
		{
			name:           "status not equal, different controller name on current",
			previous:       getPolicyStatus("ancestor1", "ns1", "ctlr1"),
			current:        getPolicyStatus("ancestor1", "ns1", "ctlr2"),
			controllerName: "ctlr1",
			expEqual:       false,
		},
		{
			name:           "status not equal, different controller name on previous",
			previous:       getPolicyStatus("ancestor1", "ns1", "ctlr2"),
			current:        getPolicyStatus("ancestor1", "ns1", "ctlr1"),
			controllerName: "ctlr1",
			expEqual:       false,
		},
		{
			name:           "status not equal, different controller ancestor changed",
			previous:       prevMultiple,
			current:        currMultiple,
			controllerName: "ctlr1",
			expEqual:       false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)
			equal := btpStatusEqual(test.controllerName, test.previous, test.current)
			g.Expect(equal).To(Equal(test.expEqual))
		})
	}
}
