package status

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/config"
)

func TestGetGatewayAddresses(t *testing.T) {
	g := NewWithT(t)

	fakeClient := fake.NewFakeClient()
	podConfig := config.GatewayPodConfig{
		PodIP:       "1.2.3.4",
		ServiceName: "my-service",
		Namespace:   "nginx-gateway",
	}

	// no Service exists yet, should get error and Pod Address
	addrs, err := GetGatewayAddresses(context.Background(), fakeClient, nil, podConfig)
	g.Expect(err).To(HaveOccurred())
	g.Expect(addrs).To(HaveLen(1))
	g.Expect(addrs[0].Value).To(Equal("1.2.3.4"))

	// Create NodePort Service and Nodes
	svc := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-service",
			Namespace: "nginx-gateway",
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeNodePort,
		},
	}
	node1 := v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node1",
		},
		Status: v1.NodeStatus{
			Addresses: []v1.NodeAddress{
				{
					Type:    v1.NodeExternalIP,
					Address: "172.0.0.1",
				},
			},
		},
	}

	node2 := v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node2",
		},
		Status: v1.NodeStatus{
			Addresses: []v1.NodeAddress{
				{
					Type:    v1.NodeInternalIP,
					Address: "10.10.10.10",
				},
			},
		},
	}

	g.Expect(fakeClient.Create(context.Background(), &svc)).To(Succeed())
	g.Expect(fakeClient.Create(context.Background(), &node1)).To(Succeed())
	g.Expect(fakeClient.Create(context.Background(), &node2)).To(Succeed())

	addrs, err = GetGatewayAddresses(context.Background(), fakeClient, nil, podConfig)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(addrs).To(HaveLen(2))
	g.Expect(addrs[0].Value).To(Equal("172.0.0.1"))
	g.Expect(addrs[1].Value).To(Equal("10.10.10.10"))

	// Change to LoadBalancer Service
	svc.Spec.Type = v1.ServiceTypeLoadBalancer
	svc.Status.LoadBalancer.Ingress = []v1.LoadBalancerIngress{
		{
			IP: "34.35.36.37",
		},
		{
			Hostname: "myhost",
		},
	}

	addrs, err = GetGatewayAddresses(context.Background(), fakeClient, &svc, podConfig)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(addrs).To(HaveLen(2))
	g.Expect(addrs[0].Value).To(Equal("34.35.36.37"))
	g.Expect(addrs[1].Value).To(Equal("myhost"))
}

func TestPrepareGatewayStatus(t *testing.T) {
	podIP := v1beta1.GatewayStatusAddress{
		Type:  helpers.GetPointer(v1beta1.IPAddressType),
		Value: "1.2.3.4",
	}
	status := GatewayStatus{
		Conditions: CreateTestConditions("GatewayTest"),
		ListenerStatuses: ListenerStatuses{
			"listener": {
				AttachedRoutes: 3,
				Conditions:     CreateTestConditions("ListenerTest"),
				SupportedKinds: []v1beta1.RouteGroupKind{
					{
						Kind: v1beta1.Kind("HTTPRoute"),
					},
				},
			},
		},
		Addresses:          []v1beta1.GatewayStatusAddress{podIP},
		ObservedGeneration: 1,
	}

	transitionTime := metav1.NewTime(time.Now())

	expected := v1beta1.GatewayStatus{
		Conditions: CreateExpectedAPIConditions("GatewayTest", 1, transitionTime),
		Listeners: []v1beta1.ListenerStatus{
			{
				Name: "listener",
				SupportedKinds: []v1beta1.RouteGroupKind{
					{
						Kind: "HTTPRoute",
					},
				},
				AttachedRoutes: 3,
				Conditions:     CreateExpectedAPIConditions("ListenerTest", 1, transitionTime),
			},
		},
		Addresses: []v1beta1.GatewayStatusAddress{podIP},
	}

	g := NewWithT(t)

	result := prepareGatewayStatus(status, transitionTime)
	g.Expect(helpers.Diff(expected, result)).To(BeEmpty())
}
