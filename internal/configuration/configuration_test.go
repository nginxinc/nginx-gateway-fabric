package configuration

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
)

func TestHTTPRoutes(t *testing.T) {
	config := NewConfiguration()

	gc := &v1alpha2.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "nginx",
		},
		Spec: v1alpha2.GatewayClassSpec{
			ControllerName: "k8s-gateway.nginx.org/nginx-gateway/gateway",
		},
	}

	var expectedChanges []Change

	changes := config.UpsertGatewayClass(gc)
	if diff := cmp.Diff(expectedChanges, changes); diff != "" {
		t.Errorf("UpsertGatewayClass() returned unexpected result (-want +got):\n%s", diff)
	}

	gw := &v1alpha2.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "nginx",
		},
		Spec: v1alpha2.GatewaySpec{
			GatewayClassName: "nginx",
			Listeners: []v1alpha2.Listener{
				{
					Name:     "http",
					Protocol: v1alpha2.HTTPProtocolType,
				},
			},
		},
	}

	expectedChanges = nil
	changes = config.UpsertGateway(gw)
	if diff := cmp.Diff(expectedChanges, changes); diff != "" {
		t.Errorf("UpsertGateway() returned unexpected result (-want +got):\n%s", diff)
	}

	httpRoute1 := &v1alpha2.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "route1",
		},
		Spec: v1alpha2.HTTPRouteSpec{
			Hostnames: []v1alpha2.Hostname{
				"cafe.example.com",
			},
			Rules: []v1alpha2.HTTPRouteRule{
				{
					// mo matches -> "/"
				},
			},
		},
	}

	expectedChanges = []Change{
		{
			Op: Upsert,
			Host: &Host{
				Value: "cafe.example.com",
				PathRouteGroups: []*PathRouteGroup{
					{
						Path: "/",
						Routes: []Route{
							{
								MatchIdx: -1,
								Rule:     &httpRoute1.Spec.Rules[0],
								Source:   httpRoute1,
							},
						},
					},
				},
			},
		},
	}
	changes = config.UpsertHTTPRoute(httpRoute1)
	if diff := cmp.Diff(expectedChanges, changes); diff != "" {
		t.Errorf("UpsertHTTPRoute() returned unexpected result (-want +got):\n%s", diff)
	}

	httpRoute2 := &v1alpha2.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "route2",
		},
		Spec: v1alpha2.HTTPRouteSpec{
			Hostnames: []v1alpha2.Hostname{
				"cafe.example.com",
			},
			Rules: []v1alpha2.HTTPRouteRule{
				{
					Matches: []v1alpha2.HTTPRouteMatch{
						{
							Path: &v1alpha2.HTTPPathMatch{
								Value: getStringPointer("/coffee"),
							},
						},
					},
				},
			},
		},
	}

	expectedChanges = []Change{
		{
			Op: Upsert,
			Host: &Host{
				Value: "cafe.example.com",
				PathRouteGroups: []*PathRouteGroup{
					{
						Path: "/",
						Routes: []Route{
							{
								MatchIdx: -1,
								Rule:     &httpRoute1.Spec.Rules[0],
								Source:   httpRoute1,
							},
						},
					},
					{
						Path: "/coffee",
						Routes: []Route{
							{
								MatchIdx: 0,
								Rule:     &httpRoute2.Spec.Rules[0],
								Source:   httpRoute2,
							},
						},
					},
				},
			},
		},
	}
	changes = config.UpsertHTTPRoute(httpRoute2)
	if diff := cmp.Diff(expectedChanges, changes); diff != "" {
		t.Errorf("UpsertHTTPRoute() returned unexpected result (-want +got):\n%s", diff)
	}

	httpRoute2Updated := &v1alpha2.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "route2",
		},
		Spec: v1alpha2.HTTPRouteSpec{
			Hostnames: []v1alpha2.Hostname{
				"cafe.example.com",
			},
			Rules: []v1alpha2.HTTPRouteRule{
				{
					Matches: []v1alpha2.HTTPRouteMatch{
						{
							Path: &v1alpha2.HTTPPathMatch{
								Value: getStringPointer("/tea"),
							},
						},
					},
				},
			},
		},
	}

	expectedChanges = []Change{
		{
			Op: Upsert,
			Host: &Host{
				Value: "cafe.example.com",
				PathRouteGroups: []*PathRouteGroup{
					{
						Path: "/",
						Routes: []Route{
							{
								MatchIdx: -1,
								Rule:     &httpRoute1.Spec.Rules[0],
								Source:   httpRoute1,
							},
						},
					},
					{
						Path: "/tea",
						Routes: []Route{
							{
								MatchIdx: 0,
								Rule:     &httpRoute2Updated.Spec.Rules[0],
								Source:   httpRoute2Updated,
							},
						},
					},
				},
			},
		},
	}
	changes = config.UpsertHTTPRoute(httpRoute2Updated)
	if diff := cmp.Diff(expectedChanges, changes); diff != "" {
		t.Errorf("UpsertHTTPRoute() returned unexpected result (-want +got):\n%s", diff)

	}

	expectedChanges = []Change{
		{
			Op: Upsert,
			Host: &Host{
				Value: "cafe.example.com",
				PathRouteGroups: []*PathRouteGroup{
					{
						Path: "/",
						Routes: []Route{
							{
								MatchIdx: -1,
								Rule:     &httpRoute1.Spec.Rules[0],
								Source:   httpRoute1,
							},
						},
					},
				},
			},
		},
	}
	changes = config.DeleteHTTPRoute("test/route2")
	if diff := cmp.Diff(expectedChanges, changes); diff != "" {
		t.Errorf("DeleteHTTPRoute() returned unexpected result (-want +got):\n%s", diff)

	}

	expectedChanges = []Change{
		{
			Op: Delete,
			Host: &Host{
				Value: "cafe.example.com",
				PathRouteGroups: []*PathRouteGroup{
					{
						Path: "/",
						Routes: []Route{
							{
								MatchIdx: -1,
								Rule:     &httpRoute1.Spec.Rules[0],
								Source:   httpRoute1,
							},
						},
					},
				},
			},
		},
	}
	changes = config.DeleteHTTPRoute("test/route1")
	if diff := cmp.Diff(expectedChanges, changes); diff != "" {
		t.Errorf("DeleteHTTPRoute() returned unexpected result (-want +got):\n%s", diff)

	}

	expectedChanges = nil
	changes = config.DeleteGateway()
	if diff := cmp.Diff(expectedChanges, changes); diff != "" {
		t.Errorf("DeleteGateway() returned unexpected result (-want +got):\n%s", diff)
	}

	expectedChanges = nil
	changes = config.DeleteGatewayClass()
	if diff := cmp.Diff(expectedChanges, changes); diff != "" {
		t.Errorf("DeleteGatewayClass() returned unexpected result (-want +got):\n%s", diff)
	}
}

func getStringPointer(s string) *string {
	return &s
}
