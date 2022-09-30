package sdk_test

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"

	"github.com/nginxinc/nginx-kubernetes-gateway/pkg/sdk"
)

func TestServicePortsChangedPredicate_Update(t *testing.T) {
	testcases := []struct {
		msg       string
		objectOld client.Object
		objectNew client.Object
		expUpdate bool
	}{
		{
			msg:       "nil objectOld",
			objectOld: nil,
			objectNew: &v1.Service{},
			expUpdate: false,
		},
		{
			msg:       "nil objectNew",
			objectOld: &v1.Service{},
			objectNew: nil,
			expUpdate: false,
		},
		{
			msg:       "non-Service objectOld",
			objectOld: &v1.Namespace{},
			objectNew: &v1.Service{},
			expUpdate: false,
		},
		{
			msg:       "non-Service objectNew",
			objectOld: &v1.Service{},
			objectNew: &v1.Namespace{},
			expUpdate: false,
		},
		{
			msg: "number of ports changed",
			objectOld: &v1.Service{
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Port:       80,
							TargetPort: intstr.FromInt(80),
						},
					},
				},
			},
			objectNew: &v1.Service{
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{},
				},
			},
			expUpdate: true,
		},
		{
			msg: "a target port changed",
			objectOld: &v1.Service{
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Port:       80,
							TargetPort: intstr.FromInt(80),
						},
						{
							Port:       81,
							TargetPort: intstr.FromInt(81),
						},
						{
							Port:       82,
							TargetPort: intstr.FromInt(82),
						},
					},
				},
			},
			objectNew: &v1.Service{
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Port:       80,
							TargetPort: intstr.FromInt(80),
						},
						{
							Port:       81,
							TargetPort: intstr.FromInt(81),
						},
						{
							Port:       82,
							TargetPort: intstr.FromInt(92), // this value changed
						},
					},
				},
			},
			expUpdate: true,
		},
		{
			msg: "a service port changed",
			objectOld: &v1.Service{
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Port:       80,
							TargetPort: intstr.FromInt(80),
						},
						{
							Port:       81,
							TargetPort: intstr.FromInt(81),
						},
						{
							Port:       82,
							TargetPort: intstr.FromInt(82),
						},
					},
				},
			},
			objectNew: &v1.Service{
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Port:       80,
							TargetPort: intstr.FromInt(80),
						},
						{
							Port:       91, // this value changed
							TargetPort: intstr.FromInt(81),
						},
						{
							Port:       82,
							TargetPort: intstr.FromInt(82),
						},
					},
				},
			},
			expUpdate: true,
		},
		{
			msg: "no ports changed",
			objectOld: &v1.Service{
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Port:       80,
							TargetPort: intstr.FromInt(80),
						},
						{
							Port:       81,
							TargetPort: intstr.FromInt(81),
						},
						{
							Port:       82,
							TargetPort: intstr.FromInt(82),
						},
					},
				},
			},
			objectNew: &v1.Service{
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Port:       80,
							TargetPort: intstr.FromInt(80),
						},
						{
							Port:       81,
							TargetPort: intstr.FromInt(81),
						},
						{
							Port:       82,
							TargetPort: intstr.FromInt(82),
						},
					},
				},
			},
			expUpdate: false,
		},
		{
			msg: "ports changed but service ports and target ports are the same",
			objectOld: &v1.Service{
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Port:       80,
							TargetPort: intstr.FromInt(80),
							Name:       "port",
						},
					},
				},
			},
			objectNew: &v1.Service{
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Port:       80,
							TargetPort: intstr.FromInt(80),
							Name:       "not-port", // name is different
						},
					},
				},
			},
			expUpdate: false,
		},
		{
			msg: "spec changed but ports are the same",
			objectOld: &v1.Service{
				Spec: v1.ServiceSpec{
					Type: v1.ServiceTypeClusterIP,
				},
			},
			objectNew: &v1.Service{
				Spec: v1.ServiceSpec{
					Type: v1.ServiceTypeNodePort,
				},
			},
			expUpdate: false,
		},
	}

	predicate := sdk.ServicePortsChangedPredicate{}

	for _, tc := range testcases {
		update := predicate.Update(event.UpdateEvent{
			ObjectOld: tc.objectOld,
			ObjectNew: tc.objectNew,
		})

		if update != tc.expUpdate {
			t.Errorf("ServicePortsChangedPredicate.Update() mismatch for %q; got %t, expected %t", tc.msg, update, tc.expUpdate)
		}
	}
}

func TestServicePortsChangedPredicate(t *testing.T) {
	predicate := sdk.ServicePortsChangedPredicate{}

	if !predicate.Delete(event.DeleteEvent{Object: &v1.Service{}}) {
		t.Errorf("ServicePortsChangedPredicate.Delete() returned false; expected true")
	}

	if !predicate.Create(event.CreateEvent{Object: &v1.Service{}}) {
		t.Errorf("ServicePortsChangedPredicate.Create() returned false; expected true")
	}

	if !predicate.Generic(event.GenericEvent{Object: &v1.Service{}}) {
		t.Errorf("ServicePortsChangedPredicate.Generic() returned false; expected true")
	}
}
