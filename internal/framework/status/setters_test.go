package status

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

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

func TestNewNginxProxyStatusSetter(t *testing.T) {
	tests := []struct {
		name         string
		status       ngfAPI.NginxProxyStatus
		newStatus    NginxProxyStatus
		expStatusSet bool
	}{
		{
			name:         "NginxProxyStatus has no status",
			expStatusSet: true,
			newStatus: NginxProxyStatus{
				Conditions: []conditions.Condition{{Message: "new condition"}},
			},
		},
		{
			name:         "NginxProxyStatus has old status",
			expStatusSet: true,
			newStatus: NginxProxyStatus{
				Conditions: []conditions.Condition{{Message: "new condition"}},
			},
			status: ngfAPI.NginxProxyStatus{
				Conditions: []v1.Condition{{Message: "old condition"}},
			},
		},
		{
			name:         "NginxProxyStatus has same status",
			expStatusSet: false,
			newStatus: NginxProxyStatus{
				Conditions: []conditions.Condition{{Message: "same condition"}},
			},
			status: ngfAPI.NginxProxyStatus{
				Conditions: []v1.Condition{{Message: "same condition"}},
			},
		},
	}

	clock := &RealClock{}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			setter := newNginxProxyStatusSetter(clock, test.newStatus)

			statusSet := setter(&ngfAPI.NginxProxy{Status: test.status})
			g.Expect(statusSet).To(Equal(test.expStatusSet))
		})
	}
}

func TestNewGatewayClassStatusSetter(t *testing.T) {
	tests := []struct {
		name         string
		status       v1beta1.GatewayClassStatus
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
			status: v1beta1.GatewayClassStatus{
				Conditions: []v1.Condition{{Message: "old condition"}},
			},
			expStatusSet: true,
		},
		{
			name: "GatewayClass has same status",
			newStatus: GatewayClassStatus{
				Conditions: []conditions.Condition{{Message: "same condition"}},
			},
			status: v1beta1.GatewayClassStatus{
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
			statusSet := setter(&v1beta1.GatewayClass{Status: test.status})
			g.Expect(statusSet).To(Equal(test.expStatusSet))
		})
	}
}

func TestNewGatewayStatusSetter(t *testing.T) {
	expAddress := v1beta1.GatewayStatusAddress{
		Type:  helpers.GetPointer(v1beta1.IPAddressType),
		Value: "10.0.0.0",
	}

	tests := []struct {
		name         string
		status       v1beta1.GatewayStatus
		newStatus    GatewayStatus
		expStatusSet bool
	}{
		{
			name: "Gateway has no status",
			newStatus: GatewayStatus{
				Conditions: []conditions.Condition{{Message: "new condition"}},
				Addresses:  []v1beta1.GatewayStatusAddress{expAddress},
			},
			expStatusSet: true,
		},
		{
			name: "Gateway has old status",
			newStatus: GatewayStatus{
				Conditions: []conditions.Condition{{Message: "new condition"}},
				Addresses:  []v1beta1.GatewayStatusAddress{expAddress},
			},
			status: v1beta1.GatewayStatus{
				Conditions: []v1.Condition{{Message: "old condition"}},
				Addresses:  []v1beta1.GatewayStatusAddress{expAddress},
			},
			expStatusSet: true,
		},
		{
			name: "Gateway has same status",
			newStatus: GatewayStatus{
				Conditions: []conditions.Condition{{Message: "same condition"}},
				Addresses:  []v1beta1.GatewayStatusAddress{expAddress},
			},
			status: v1beta1.GatewayStatus{
				Conditions: []v1.Condition{{Message: "same condition"}},
				Addresses:  []v1beta1.GatewayStatusAddress{expAddress},
			},
			expStatusSet: false,
		},
	}

	clock := &RealClock{}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			setter := newGatewayStatusSetter(clock, test.newStatus)

			statusSet := setter(&v1beta1.Gateway{Status: test.status})
			g.Expect(statusSet).To(Equal(test.expStatusSet))
		})
	}
}

func TestNewHTTPRouteStatusSetter(t *testing.T) {
	controllerName := "controller"

	tests := []struct {
		name         string
		status       v1beta1.HTTPRouteStatus
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
			status: v1beta1.HTTPRouteStatus{
				RouteStatus: v1beta1.RouteStatus{
					Parents: []v1beta1.RouteParentStatus{
						{
							ParentRef:      v1beta1.ParentReference{},
							ControllerName: v1beta1.GatewayController(controllerName),
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
			status: v1beta1.HTTPRouteStatus{
				RouteStatus: v1beta1.RouteStatus{
					Parents: []v1beta1.RouteParentStatus{
						{
							ParentRef:      v1beta1.ParentReference{},
							ControllerName: v1beta1.GatewayController(controllerName),
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

			statusSet := setter(&v1beta1.HTTPRoute{Status: test.status})
			g.Expect(statusSet).To(Equal(test.expStatusSet))
		})
	}
}

func TestGWStatusEqual(t *testing.T) {
	getDefaultStatus := func() v1beta1.GatewayStatus {
		return v1beta1.GatewayStatus{
			Addresses: []v1beta1.GatewayStatusAddress{
				{
					Type:  helpers.GetPointer(v1beta1.IPAddressType),
					Value: "10.0.0.0",
				},
				{
					Type:  helpers.GetPointer(v1beta1.IPAddressType),
					Value: "11.0.0.0",
				},
			},
			Conditions: []v1.Condition{
				{
					Type: "type", /* conditions are covered by another test*/
				},
			},
			Listeners: []v1beta1.ListenerStatus{
				{
					Name: "listener1",
					SupportedKinds: []v1beta1.RouteGroupKind{
						{
							Group: helpers.GetPointer[v1beta1.Group](v1beta1.GroupName),
							Kind:  "HTTPRoute",
						},
						{
							Group: helpers.GetPointer[v1beta1.Group](v1beta1.GroupName),
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
					SupportedKinds: []v1beta1.RouteGroupKind{
						{
							Group: helpers.GetPointer[v1beta1.Group](v1beta1.GroupName),
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
					SupportedKinds: []v1beta1.RouteGroupKind{
						{
							Group: helpers.GetPointer[v1beta1.Group](v1beta1.GroupName),
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

	getModifiedStatus := func(mod func(v1beta1.GatewayStatus) v1beta1.GatewayStatus) v1beta1.GatewayStatus {
		return mod(getDefaultStatus())
	}

	tests := []struct {
		name       string
		prevStatus v1beta1.GatewayStatus
		curStatus  v1beta1.GatewayStatus
		expEqual   bool
	}{
		{
			name:       "different number of addresses",
			prevStatus: getDefaultStatus(),
			curStatus: getModifiedStatus(func(status v1beta1.GatewayStatus) v1beta1.GatewayStatus {
				status.Addresses = status.Addresses[:1]
				return status
			}),
			expEqual: false,
		},
		{
			name:       "different address type",
			prevStatus: getDefaultStatus(),
			curStatus: getModifiedStatus(func(status v1beta1.GatewayStatus) v1beta1.GatewayStatus {
				status.Addresses[1].Type = helpers.GetPointer(v1beta1.HostnameAddressType)
				return status
			}),
			expEqual: false,
		},
		{
			name:       "different address value",
			prevStatus: getDefaultStatus(),
			curStatus: getModifiedStatus(func(status v1beta1.GatewayStatus) v1beta1.GatewayStatus {
				status.Addresses[0].Value = "12.0.0.0"
				return status
			}),
			expEqual: false,
		},
		{
			name:       "different conditions",
			prevStatus: getDefaultStatus(),
			curStatus: getModifiedStatus(func(status v1beta1.GatewayStatus) v1beta1.GatewayStatus {
				status.Conditions[0].Type = "different"
				return status
			}),
			expEqual: false,
		},
		{
			name:       "different number of listener statuses",
			prevStatus: getDefaultStatus(),
			curStatus: getModifiedStatus(func(status v1beta1.GatewayStatus) v1beta1.GatewayStatus {
				status.Listeners = status.Listeners[:2]
				return status
			}),
			expEqual: false,
		},
		{
			name:       "different listener status name",
			prevStatus: getDefaultStatus(),
			curStatus: getModifiedStatus(func(status v1beta1.GatewayStatus) v1beta1.GatewayStatus {
				status.Listeners[2].Name = "different"
				return status
			}),
			expEqual: false,
		},
		{
			name:       "different listener status attached routes",
			prevStatus: getDefaultStatus(),
			curStatus: getModifiedStatus(func(status v1beta1.GatewayStatus) v1beta1.GatewayStatus {
				status.Listeners[1].AttachedRoutes++
				return status
			}),
			expEqual: false,
		},
		{
			name:       "different listener status conditions",
			prevStatus: getDefaultStatus(),
			curStatus: getModifiedStatus(func(status v1beta1.GatewayStatus) v1beta1.GatewayStatus {
				status.Listeners[0].Conditions[0].Type = "different"
				return status
			}),
			expEqual: false,
		},
		{
			name:       "different listener status supported kinds (different number)",
			prevStatus: getDefaultStatus(),
			curStatus: getModifiedStatus(func(status v1beta1.GatewayStatus) v1beta1.GatewayStatus {
				status.Listeners[0].SupportedKinds = status.Listeners[0].SupportedKinds[:1]
				return status
			}),
			expEqual: false,
		},
		{
			name:       "different listener status supported kinds (different kind)",
			prevStatus: getDefaultStatus(),
			curStatus: getModifiedStatus(func(status v1beta1.GatewayStatus) v1beta1.GatewayStatus {
				status.Listeners[1].SupportedKinds[0].Kind = "TCPRoute"
				return status
			}),
			expEqual: false,
		},
		{
			name:       "different listener status supported kinds (different group)",
			prevStatus: getDefaultStatus(),
			curStatus: getModifiedStatus(func(status v1beta1.GatewayStatus) v1beta1.GatewayStatus {
				status.Listeners[1].SupportedKinds[0].Group = helpers.GetPointer[v1beta1.Group]("different")
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

	previousStatus := v1beta1.HTTPRouteStatus{
		RouteStatus: v1beta1.RouteStatus{
			Parents: []v1beta1.RouteParentStatus{
				{
					ParentRef: v1beta1.ParentReference{
						Namespace:   helpers.GetPointer[v1beta1.Namespace]("test"),
						Name:        "our-parent",
						SectionName: helpers.GetPointer[v1beta1.SectionName]("section1"),
					},
					ControllerName: "ours",
					Conditions:     testConds,
				},
				{
					ParentRef: v1beta1.ParentReference{
						Namespace:   helpers.GetPointer[v1beta1.Namespace]("test"),
						Name:        "not-our-parent",
						SectionName: helpers.GetPointer[v1beta1.SectionName]("section1"),
					},
					ControllerName: "not-ours",
					Conditions:     testConds,
				},
				{
					ParentRef: v1beta1.ParentReference{
						Namespace:   helpers.GetPointer[v1beta1.Namespace]("test"),
						Name:        "our-parent",
						SectionName: helpers.GetPointer[v1beta1.SectionName]("section2"),
					},
					ControllerName: "ours",
					Conditions:     testConds,
				},
				{
					ParentRef: v1beta1.ParentReference{
						Namespace:   helpers.GetPointer[v1beta1.Namespace]("test"),
						Name:        "not-our-parent",
						SectionName: helpers.GetPointer[v1beta1.SectionName]("section2"),
					},
					ControllerName: "not-ours",
					Conditions:     testConds,
				},
			},
		},
	}

	getDefaultStatus := func() v1beta1.HTTPRouteStatus {
		return v1beta1.HTTPRouteStatus{
			RouteStatus: v1beta1.RouteStatus{
				Parents: []v1beta1.RouteParentStatus{
					{
						ParentRef: v1beta1.ParentReference{
							Namespace:   helpers.GetPointer[v1beta1.Namespace]("test"),
							Name:        "our-parent",
							SectionName: helpers.GetPointer[v1beta1.SectionName]("section1"),
						},
						ControllerName: "ours",
						Conditions:     testConds,
					},
					{
						ParentRef: v1beta1.ParentReference{
							Namespace:   helpers.GetPointer[v1beta1.Namespace]("test"),
							Name:        "our-parent",
							SectionName: helpers.GetPointer[v1beta1.SectionName]("section2"),
						},
						ControllerName: "ours",
						Conditions:     testConds,
					},
				},
			},
		}
	}

	newParentStatus := v1beta1.RouteParentStatus{
		ParentRef: v1beta1.ParentReference{
			Namespace:   helpers.GetPointer[v1beta1.Namespace]("test"),
			Name:        "our-parent",
			SectionName: helpers.GetPointer[v1beta1.SectionName]("section3"),
		},
		ControllerName: "ours",
		Conditions:     testConds,
	}

	getModifiedStatus := func(mod func(status v1beta1.HTTPRouteStatus) v1beta1.HTTPRouteStatus) v1beta1.HTTPRouteStatus {
		return mod(getDefaultStatus())
	}

	tests := []struct {
		name       string
		prevStatus v1beta1.HTTPRouteStatus
		curStatus  v1beta1.HTTPRouteStatus
		expEqual   bool
	}{
		{
			name:       "stale status",
			prevStatus: previousStatus,
			curStatus: getModifiedStatus(func(status v1beta1.HTTPRouteStatus) v1beta1.HTTPRouteStatus {
				// remove last parent status
				status.Parents = status.Parents[:1]
				return status
			}),
			expEqual: false,
		},
		{
			name:       "new status",
			prevStatus: previousStatus,
			curStatus: getModifiedStatus(func(status v1beta1.HTTPRouteStatus) v1beta1.HTTPRouteStatus {
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
	getDefaultStatus := func() v1beta1.RouteParentStatus {
		return v1beta1.RouteParentStatus{
			ParentRef: v1beta1.ParentReference{
				Namespace:   helpers.GetPointer[v1beta1.Namespace]("test"),
				Name:        "parent",
				SectionName: helpers.GetPointer[v1beta1.SectionName]("section"),
			},
			ControllerName: "controller",
			Conditions: []v1.Condition{
				{
					Type: "type", /* conditions are covered by another test*/
				},
			},
		}
	}

	getModifiedStatus := func(mod func(v1beta1.RouteParentStatus) v1beta1.RouteParentStatus) v1beta1.RouteParentStatus {
		return mod(getDefaultStatus())
	}

	tests := []struct {
		name     string
		p1       v1beta1.RouteParentStatus
		p2       v1beta1.RouteParentStatus
		expEqual bool
	}{
		{
			name: "different controller name",
			p1:   getDefaultStatus(),
			p2: getModifiedStatus(func(status v1beta1.RouteParentStatus) v1beta1.RouteParentStatus {
				status.ControllerName = "different"
				return status
			}),
			expEqual: false,
		},
		{
			name: "different parentRef name",
			p1:   getDefaultStatus(),
			p2: getModifiedStatus(func(status v1beta1.RouteParentStatus) v1beta1.RouteParentStatus {
				status.ParentRef.Name = "different"
				return status
			}),
			expEqual: false,
		},
		{
			name: "different parentRef namespace",
			p1:   getDefaultStatus(),
			p2: getModifiedStatus(func(status v1beta1.RouteParentStatus) v1beta1.RouteParentStatus {
				status.ParentRef.Namespace = helpers.GetPointer[v1beta1.Namespace]("different")
				return status
			}),
			expEqual: false,
		},
		{
			name: "different parentRef section name",
			p1:   getDefaultStatus(),
			p2: getModifiedStatus(func(status v1beta1.RouteParentStatus) v1beta1.RouteParentStatus {
				status.ParentRef.SectionName = helpers.GetPointer[v1beta1.SectionName]("different")
				return status
			}),
			expEqual: false,
		},
		{
			name: "different conditions",
			p1:   getDefaultStatus(),
			p2: getModifiedStatus(func(status v1beta1.RouteParentStatus) v1beta1.RouteParentStatus {
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
