package graph

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/helpers"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/conditions"
)

func TestBindRouteToListeners(t *testing.T) {
	createRoute := func(hostname string, parentRefs ...v1beta1.ParentReference) *v1beta1.HTTPRoute {
		return &v1beta1.HTTPRoute{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test",
				Name:      "hr-1",
			},
			Spec: v1beta1.HTTPRouteSpec{
				CommonRouteSpec: v1beta1.CommonRouteSpec{
					ParentRefs: parentRefs,
				},
				Hostnames: []v1beta1.Hostname{
					v1beta1.Hostname(hostname),
				},
			},
		}
	}

	hrNonExistingSectionName := createRoute("foo.example.com", v1beta1.ParentReference{
		Namespace:   (*v1beta1.Namespace)(helpers.GetStringPointer("test")),
		Name:        "gateway",
		SectionName: (*v1beta1.SectionName)(helpers.GetStringPointer("listener-80-2")),
	})

	hrEmptySectionName := createRoute("foo.example.com", v1beta1.ParentReference{
		Namespace: (*v1beta1.Namespace)(helpers.GetStringPointer("test")),
		Name:      "gateway",
	})

	hrIgnoredGateway := createRoute("foo.example.com", v1beta1.ParentReference{
		Namespace:   (*v1beta1.Namespace)(helpers.GetStringPointer("test")),
		Name:        "ignored-gateway",
		SectionName: (*v1beta1.SectionName)(helpers.GetStringPointer("listener-80-1")),
	})

	hrFoo := createRoute("foo.example.com", v1beta1.ParentReference{
		Namespace:   (*v1beta1.Namespace)(helpers.GetStringPointer("test")),
		Name:        "gateway",
		SectionName: (*v1beta1.SectionName)(helpers.GetStringPointer("listener-80-1")),
	})

	hrFooImplicitNamespace := createRoute("foo.example.com", v1beta1.ParentReference{
		Name:        "gateway",
		SectionName: (*v1beta1.SectionName)(helpers.GetStringPointer("listener-80-1")),
	})

	hrBar := createRoute("bar.example.com", v1beta1.ParentReference{
		Namespace:   (*v1beta1.Namespace)(helpers.GetStringPointer("test")),
		Name:        "gateway",
		SectionName: (*v1beta1.SectionName)(helpers.GetStringPointer("listener-80-1")),
	})

	// we create a new listener each time because the function under test can modify it
	createListener := func() *Listener {
		return &Listener{
			Source: v1beta1.Listener{
				Hostname: (*v1beta1.Hostname)(helpers.GetStringPointer("foo.example.com")),
			},
			Valid:             true,
			Routes:            map[types.NamespacedName]*Route{},
			AcceptedHostnames: map[string]struct{}{},
		}
	}

	createModifiedListener := func(m func(*Listener)) *Listener {
		l := createListener()
		m(l)
		return l
	}

	gw := &v1beta1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "gateway",
		},
	}

	tests := []struct {
		httpRoute         *v1beta1.HTTPRoute
		gw                *v1beta1.Gateway
		ignoredGws        map[types.NamespacedName]*v1beta1.Gateway
		listeners         map[string]*Listener
		expectedRoute     *Route
		expectedListeners map[string]*Listener
		msg               string
		expectedIgnored   bool
	}{
		{
			httpRoute:  createRoute("foo.example.com"),
			gw:         gw,
			ignoredGws: nil,
			listeners: map[string]*Listener{
				"listener-80-1": createListener(),
			},
			expectedIgnored: true,
			expectedRoute:   nil,
			expectedListeners: map[string]*Listener{
				"listener-80-1": createListener(),
			},
			msg: "HTTPRoute without parent refs",
		},
		{
			httpRoute: createRoute("foo.example.com", v1beta1.ParentReference{
				Namespace:   (*v1beta1.Namespace)(helpers.GetStringPointer("test")),
				Name:        "some-gateway", // wrong gateway
				SectionName: (*v1beta1.SectionName)(helpers.GetStringPointer("listener-1")),
			}),
			gw:         gw,
			ignoredGws: nil,
			listeners: map[string]*Listener{
				"listener-80-1": createListener(),
			},
			expectedIgnored: true,
			expectedRoute:   nil,
			expectedListeners: map[string]*Listener{
				"listener-80-1": createListener(),
			},
			msg: "HTTPRoute without good parent refs",
		},
		{
			httpRoute:  hrNonExistingSectionName,
			gw:         gw,
			ignoredGws: nil,
			listeners: map[string]*Listener{
				"listener-80-1": createListener(),
			},
			expectedIgnored: false,
			expectedRoute: &Route{
				Source:               hrNonExistingSectionName,
				ValidSectionNameRefs: map[string]struct{}{},
				InvalidSectionNameRefs: map[string]conditions.Condition{
					"listener-80-2": conditions.NewTODO("listener is not found"),
				},
			},
			expectedListeners: map[string]*Listener{
				"listener-80-1": createListener(),
			},
			msg: "HTTPRoute with non-existing section name",
		},
		{
			httpRoute:  hrEmptySectionName,
			gw:         gw,
			ignoredGws: nil,
			listeners: map[string]*Listener{
				"listener-80-1": createListener(),
			},
			expectedIgnored: true,
			expectedRoute:   nil,
			expectedListeners: map[string]*Listener{
				"listener-80-1": createListener(),
			},
			msg: "HTTPRoute with empty section name",
		},
		{
			httpRoute:  hrFoo,
			gw:         gw,
			ignoredGws: nil,
			listeners: map[string]*Listener{
				"listener-80-1": createListener(),
			},
			expectedIgnored: false,
			expectedRoute: &Route{
				Source: hrFoo,
				ValidSectionNameRefs: map[string]struct{}{
					"listener-80-1": {},
				},
				InvalidSectionNameRefs: map[string]conditions.Condition{},
			},
			expectedListeners: map[string]*Listener{
				"listener-80-1": createModifiedListener(func(l *Listener) {
					l.Routes = map[types.NamespacedName]*Route{
						{Namespace: "test", Name: "hr-1"}: {
							Source: hrFoo,
							ValidSectionNameRefs: map[string]struct{}{
								"listener-80-1": {},
							},
							InvalidSectionNameRefs: map[string]conditions.Condition{},
						},
					}
					l.AcceptedHostnames = map[string]struct{}{
						"foo.example.com": {},
					}
				}),
			},
			msg: "HTTPRoute with one accepted hostname",
		},
		{
			httpRoute:  hrFooImplicitNamespace,
			gw:         gw,
			ignoredGws: nil,
			listeners: map[string]*Listener{
				"listener-80-1": createListener(),
			},
			expectedIgnored: false,
			expectedRoute: &Route{
				Source: hrFooImplicitNamespace,
				ValidSectionNameRefs: map[string]struct{}{
					"listener-80-1": {},
				},
				InvalidSectionNameRefs: map[string]conditions.Condition{},
			},
			expectedListeners: map[string]*Listener{
				"listener-80-1": createModifiedListener(func(l *Listener) {
					l.Routes = map[types.NamespacedName]*Route{
						{Namespace: "test", Name: "hr-1"}: {
							Source: hrFooImplicitNamespace,
							ValidSectionNameRefs: map[string]struct{}{
								"listener-80-1": {},
							},
							InvalidSectionNameRefs: map[string]conditions.Condition{},
						},
					}
					l.AcceptedHostnames = map[string]struct{}{
						"foo.example.com": {},
					}
				}),
			},
			msg: "HTTPRoute with one accepted hostname with implicit namespace in parentRef",
		},
		{
			httpRoute:  hrBar,
			gw:         gw,
			ignoredGws: nil,
			listeners: map[string]*Listener{
				"listener-80-1": createListener(),
			},
			expectedIgnored: false,
			expectedRoute: &Route{
				Source:               hrBar,
				ValidSectionNameRefs: map[string]struct{}{},
				InvalidSectionNameRefs: map[string]conditions.Condition{
					"listener-80-1": conditions.NewRouteNoMatchingListenerHostname(),
				},
			},
			expectedListeners: map[string]*Listener{
				"listener-80-1": createListener(),
			},
			msg: "HTTPRoute with zero accepted hostnames",
		},
		{
			httpRoute: hrIgnoredGateway,
			gw:        gw,
			ignoredGws: map[types.NamespacedName]*v1beta1.Gateway{
				{Namespace: "test", Name: "ignored-gateway"}: {},
			},
			listeners: map[string]*Listener{
				"listener-80-1": createListener(),
			},
			expectedIgnored: false,
			expectedRoute: &Route{
				Source:               hrIgnoredGateway,
				ValidSectionNameRefs: map[string]struct{}{},
				InvalidSectionNameRefs: map[string]conditions.Condition{
					"listener-80-1": conditions.NewTODO("Gateway is ignored"),
				},
			},
			expectedListeners: map[string]*Listener{
				"listener-80-1": createListener(),
			},
			msg: "HTTPRoute with ignored gateway reference",
		},
		{
			httpRoute:         hrFoo,
			gw:                nil,
			ignoredGws:        nil,
			listeners:         nil,
			expectedIgnored:   true,
			expectedRoute:     nil,
			expectedListeners: nil,
			msg:               "HTTPRoute when no gateway exists",
		},
		{
			httpRoute:  hrFoo,
			gw:         gw,
			ignoredGws: nil,
			listeners: map[string]*Listener{
				"listener-80-1": createModifiedListener(func(l *Listener) {
					l.Valid = false
				}),
			},
			expectedIgnored: false,
			expectedRoute: &Route{
				Source:               hrFoo,
				ValidSectionNameRefs: map[string]struct{}{},
				InvalidSectionNameRefs: map[string]conditions.Condition{
					"listener-80-1": conditions.NewRouteInvalidListener(),
				},
			},
			expectedListeners: map[string]*Listener{
				"listener-80-1": createModifiedListener(func(l *Listener) {
					l.Valid = false
				}),
			},
			msg: "HTTPRoute with invalid listener parentRef",
		},
	}

	for _, test := range tests {
		ignored, route := bindHTTPRouteToListeners(test.httpRoute, test.gw, test.ignoredGws, test.listeners)
		if diff := cmp.Diff(test.expectedIgnored, ignored); diff != "" {
			t.Errorf("bindHTTPRouteToListeners() %q  mismatch on ignored (-want +got):\n%s", test.msg, diff)
		}
		if diff := cmp.Diff(test.expectedRoute, route); diff != "" {
			t.Errorf("bindHTTPRouteToListeners() %q  mismatch on route (-want +got):\n%s", test.msg, diff)
		}
		if diff := cmp.Diff(test.expectedListeners, test.listeners); diff != "" {
			t.Errorf("bindHTTPRouteToListeners() %q  mismatch on listeners (-want +got):\n%s", test.msg, diff)
		}
	}
}

func TestFindAcceptedHostnames(t *testing.T) {
	var listenerHostnameFoo v1beta1.Hostname = "foo.example.com"
	var listenerHostnameCafe v1beta1.Hostname = "cafe.example.com"
	routeHostnames := []v1beta1.Hostname{"foo.example.com", "bar.example.com"}

	tests := []struct {
		listenerHostname *v1beta1.Hostname
		msg              string
		routeHostnames   []v1beta1.Hostname
		expected         []string
	}{
		{
			listenerHostname: &listenerHostnameFoo,
			routeHostnames:   routeHostnames,
			expected:         []string{"foo.example.com"},
			msg:              "one match",
		},
		{
			listenerHostname: &listenerHostnameCafe,
			routeHostnames:   routeHostnames,
			expected:         nil,
			msg:              "no match",
		},
		{
			listenerHostname: nil,
			routeHostnames:   routeHostnames,
			expected:         []string{"foo.example.com", "bar.example.com"},
			msg:              "nil listener hostname",
		},
	}

	for _, test := range tests {
		result := findAcceptedHostnames(test.listenerHostname, test.routeHostnames)
		if diff := cmp.Diff(test.expected, result); diff != "" {
			t.Errorf("findAcceptedHostnames() %q  mismatch (-want +got):\n%s", test.msg, diff)
		}
	}
}

func TestGetHostname(t *testing.T) {
	var emptyHostname v1beta1.Hostname
	var hostname v1beta1.Hostname = "example.com"

	tests := []struct {
		h        *v1beta1.Hostname
		expected string
		msg      string
	}{
		{
			h:        nil,
			expected: "",
			msg:      "nil hostname",
		},
		{
			h:        &emptyHostname,
			expected: "",
			msg:      "empty hostname",
		},
		{
			h:        &hostname,
			expected: string(hostname),
			msg:      "normal hostname",
		},
	}

	for _, test := range tests {
		result := getHostname(test.h)
		if result != test.expected {
			t.Errorf("getHostname() returned %q but expected %q for the case of %q", result, test.expected, test.msg)
		}
	}
}
