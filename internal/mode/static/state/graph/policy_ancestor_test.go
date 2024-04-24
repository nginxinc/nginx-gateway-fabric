package graph

import (
	"testing"

	. "github.com/onsi/gomega"
	v1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/kinds"
)

func TestAncestorsFull(t *testing.T) {
	ns := "test"

	createCurStatus := func(numAncestors int, ctlrName string) []v1alpha2.PolicyAncestorStatus {
		statuses := make([]v1alpha2.PolicyAncestorStatus, 0, numAncestors)

		for i := 0; i < numAncestors; i++ {
			statuses = append(statuses, v1alpha2.PolicyAncestorStatus{
				AncestorRef: v1alpha2.ParentReference{
					Group:     helpers.GetPointer[v1.Group](v1.GroupName),
					Kind:      helpers.GetPointer[v1.Kind](kinds.Gateway),
					Namespace: (*v1.Namespace)(&ns),
					Name:      "name",
				},
				ControllerName: v1.GatewayController(ctlrName),
			})
		}

		return statuses
	}

	tests := []struct {
		newAncestor v1alpha2.ParentReference
		name        string
		curStatus   []v1alpha2.PolicyAncestorStatus
		expFull     bool
	}{
		{
			name:      "not full",
			curStatus: createCurStatus(15, "controller"),
			newAncestor: v1alpha2.ParentReference{
				Group:     helpers.GetPointer[v1.Group](v1.GroupName),
				Kind:      helpers.GetPointer[v1.Kind](kinds.Gateway),
				Namespace: (*v1.Namespace)(&ns),
				Name:      "name",
			},
			expFull: false,
		},
		{
			name:      "full; ancestor does not exist in current status",
			curStatus: createCurStatus(16, "controller"),
			newAncestor: v1alpha2.ParentReference{
				Group:     helpers.GetPointer[v1.Group](v1.GroupName),
				Kind:      helpers.GetPointer[v1.Kind](kinds.Gateway),
				Namespace: (*v1.Namespace)(&ns),
				Name:      "name",
			},
			expFull: true,
		},
		{
			name:      "full, but ancestor does exist in current status",
			curStatus: createCurStatus(16, "nginx-gateway"),
			newAncestor: v1alpha2.ParentReference{
				Group:     helpers.GetPointer[v1.Group](v1.GroupName),
				Kind:      helpers.GetPointer[v1.Kind](kinds.Gateway),
				Namespace: (*v1.Namespace)(&ns),
				Name:      "name",
			},
			expFull: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			full := ancestorsFull(test.curStatus, test.newAncestor, "nginx-gateway")
			g.Expect(full).To(Equal(test.expFull))
		})
	}
}

func TestAncestorStatusExists(t *testing.T) {
	getStatus := func() v1alpha2.PolicyAncestorStatus {
		ns := "test"

		return v1alpha2.PolicyAncestorStatus{
			AncestorRef: v1alpha2.ParentReference{
				Group:     helpers.GetPointer[v1.Group](v1.GroupName),
				Kind:      helpers.GetPointer[v1.Kind](kinds.Gateway),
				Namespace: (*v1.Namespace)(&ns),
				Name:      "name",
			},
			ControllerName: "nginx-gateway",
		}
	}

	type modFunc func(s v1alpha2.PolicyAncestorStatus) v1alpha2.PolicyAncestorStatus

	getModifiedStatus := func(mod modFunc) v1alpha2.PolicyAncestorStatus {
		return mod(getStatus())
	}

	tests := []struct {
		name      string
		curStatus v1alpha2.PolicyAncestorStatus
		newStatus v1alpha2.PolicyAncestorStatus
		expEqual  bool
	}{
		{
			name:      "equal",
			curStatus: getStatus(),
			newStatus: getStatus(),
			expEqual:  true,
		},
		{
			name:      "different controller name",
			curStatus: getStatus(),
			newStatus: getModifiedStatus(func(s v1alpha2.PolicyAncestorStatus) v1alpha2.PolicyAncestorStatus {
				s.ControllerName = "not-ours"
				return s
			}),
			expEqual: false,
		},
		{
			name:      "different groups; one nil",
			curStatus: getStatus(),
			newStatus: getModifiedStatus(func(s v1alpha2.PolicyAncestorStatus) v1alpha2.PolicyAncestorStatus {
				s.AncestorRef.Group = nil
				return s
			}),
			expEqual: false,
		},
		{
			name:      "different groups",
			curStatus: getStatus(),
			newStatus: getModifiedStatus(func(s v1alpha2.PolicyAncestorStatus) v1alpha2.PolicyAncestorStatus {
				s.AncestorRef.Group = helpers.GetPointer[v1.Group]("DiffGroup")
				return s
			}),
			expEqual: false,
		},
		{
			name:      "different kinds; one nil",
			curStatus: getStatus(),
			newStatus: getModifiedStatus(func(s v1alpha2.PolicyAncestorStatus) v1alpha2.PolicyAncestorStatus {
				s.AncestorRef.Kind = nil
				return s
			}),
			expEqual: false,
		},
		{
			name:      "different kinds",
			curStatus: getStatus(),
			newStatus: getModifiedStatus(func(s v1alpha2.PolicyAncestorStatus) v1alpha2.PolicyAncestorStatus {
				s.AncestorRef.Kind = helpers.GetPointer[v1.Kind](kinds.HTTPRoute)
				return s
			}),
			expEqual: false,
		},
		{
			name:      "different names",
			curStatus: getStatus(),
			newStatus: getModifiedStatus(func(s v1alpha2.PolicyAncestorStatus) v1alpha2.PolicyAncestorStatus {
				s.AncestorRef.Name = "diff-name"
				return s
			}),
			expEqual: false,
		},
		{
			name:      "different namespaces; one nil",
			curStatus: getStatus(),
			newStatus: getModifiedStatus(func(s v1alpha2.PolicyAncestorStatus) v1alpha2.PolicyAncestorStatus {
				s.AncestorRef.Namespace = nil
				return s
			}),
			expEqual: false,
		},
		{
			name: "different namespaces",
			newStatus: getModifiedStatus(func(s v1alpha2.PolicyAncestorStatus) v1alpha2.PolicyAncestorStatus {
				diffNs := "diff"
				s.AncestorRef.Namespace = (*v1.Namespace)(&diffNs)
				return s
			}),
			expEqual: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			equal := ancestorStatusEqual(test.curStatus, test.newStatus)
			g.Expect(equal).To(Equal(test.expEqual))
		})
	}
}
