package predicate

import (
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func TestServicePortsChangedPredicate_Update(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		objectOld client.Object
		objectNew client.Object
		msg       string
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

	p := ServicePortsChangedPredicate{}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.msg, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)
			update := p.Update(event.UpdateEvent{
				ObjectOld: tc.objectOld,
				ObjectNew: tc.objectNew,
			})

			g.Expect(update).To(Equal(tc.expUpdate))
		})
	}
}

func TestServicePortsChangedPredicate(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	p := GatewayServicePredicate{}

	g.Expect(p.Delete(event.DeleteEvent{Object: &v1.Service{}})).To(BeTrue())
	g.Expect(p.Create(event.CreateEvent{Object: &v1.Service{}})).To(BeTrue())
	g.Expect(p.Generic(event.GenericEvent{Object: &v1.Service{}})).To(BeTrue())
}

func TestGatewayServicePredicate_Update(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		objectOld client.Object
		objectNew client.Object
		msg       string
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
			msg:       "Service not watched",
			objectOld: &v1.Service{},
			objectNew: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "nginx-gateway",
					Name:      "not-watched",
				},
			},
			expUpdate: false,
		},
		{
			msg: "something irrelevant changed",
			objectOld: &v1.Service{
				Spec: v1.ServiceSpec{
					ClusterIP: "1.2.3.4",
				},
			},
			objectNew: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "nginx-gateway",
					Name:      "nginx",
				},
				Spec: v1.ServiceSpec{
					ClusterIP: "5.6.7.8",
				},
			},
			expUpdate: false,
		},
		{
			msg: "type changed",
			objectOld: &v1.Service{
				Spec: v1.ServiceSpec{
					Type: v1.ServiceTypeLoadBalancer,
				},
			},
			objectNew: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "nginx-gateway",
					Name:      "nginx",
				},
				Spec: v1.ServiceSpec{
					Type: v1.ServiceTypeNodePort,
				},
			},
			expUpdate: true,
		},
		{
			msg: "ingress changed length",
			objectOld: &v1.Service{
				Spec: v1.ServiceSpec{
					Type: v1.ServiceTypeLoadBalancer,
				},
				Status: v1.ServiceStatus{
					LoadBalancer: v1.LoadBalancerStatus{
						Ingress: []v1.LoadBalancerIngress{
							{
								IP: "1.2.3.4",
							},
						},
					},
				},
			},
			objectNew: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "nginx-gateway",
					Name:      "nginx",
				},
				Spec: v1.ServiceSpec{
					Type: v1.ServiceTypeNodePort,
				}, Status: v1.ServiceStatus{
					LoadBalancer: v1.LoadBalancerStatus{
						Ingress: []v1.LoadBalancerIngress{
							{
								IP: "1.2.3.4",
							},
							{
								IP: "5.6.7.8",
							},
						},
					},
				},
			},
			expUpdate: true,
		},
		{
			msg: "IP address changed",
			objectOld: &v1.Service{
				Spec: v1.ServiceSpec{
					Type: v1.ServiceTypeLoadBalancer,
				},
				Status: v1.ServiceStatus{
					LoadBalancer: v1.LoadBalancerStatus{
						Ingress: []v1.LoadBalancerIngress{
							{
								IP: "1.2.3.4",
							},
						},
					},
				},
			},
			objectNew: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "nginx-gateway",
					Name:      "nginx",
				},
				Spec: v1.ServiceSpec{
					Type: v1.ServiceTypeNodePort,
				}, Status: v1.ServiceStatus{
					LoadBalancer: v1.LoadBalancerStatus{
						Ingress: []v1.LoadBalancerIngress{
							{
								IP: "5.6.7.8",
							},
						},
					},
				},
			},
			expUpdate: true,
		},
		{
			msg: "Hostname changed",
			objectOld: &v1.Service{
				Spec: v1.ServiceSpec{
					Type: v1.ServiceTypeLoadBalancer,
				},
				Status: v1.ServiceStatus{
					LoadBalancer: v1.LoadBalancerStatus{
						Ingress: []v1.LoadBalancerIngress{
							{
								Hostname: "one",
							},
						},
					},
				},
			},
			objectNew: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "nginx-gateway",
					Name:      "nginx",
				},
				Spec: v1.ServiceSpec{
					Type: v1.ServiceTypeNodePort,
				}, Status: v1.ServiceStatus{
					LoadBalancer: v1.LoadBalancerStatus{
						Ingress: []v1.LoadBalancerIngress{
							{
								Hostname: "two",
							},
						},
					},
				},
			},
			expUpdate: true,
		},
	}

	p := GatewayServicePredicate{NSName: types.NamespacedName{Namespace: "nginx-gateway", Name: "nginx"}}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.msg, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)
			update := p.Update(event.UpdateEvent{
				ObjectOld: tc.objectOld,
				ObjectNew: tc.objectNew,
			})

			g.Expect(update).To(Equal(tc.expUpdate))
		})
	}
}

func TestGatewayServicePredicate(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	p := GatewayServicePredicate{}

	g.Expect(p.Delete(event.DeleteEvent{Object: &v1.Service{}})).To(BeTrue())
	g.Expect(p.Create(event.CreateEvent{Object: &v1.Service{}})).To(BeTrue())
	g.Expect(p.Generic(event.GenericEvent{Object: &v1.Service{}})).To(BeTrue())
}
