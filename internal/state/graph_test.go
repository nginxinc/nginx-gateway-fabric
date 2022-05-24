package state

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/helpers"
)

func TestBuildGraph(t *testing.T) {
	createRoute := func(name string, gatewayName string) *v1alpha2.HTTPRoute {
		return &v1alpha2.HTTPRoute{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test",
				Name:      name,
			},
			Spec: v1alpha2.HTTPRouteSpec{
				CommonRouteSpec: v1alpha2.CommonRouteSpec{
					ParentRefs: []v1alpha2.ParentRef{
						{
							Namespace:   (*v1alpha2.Namespace)(helpers.GetStringPointer("test")),
							Name:        v1alpha2.ObjectName(gatewayName),
							SectionName: (*v1alpha2.SectionName)(helpers.GetStringPointer("listener-80-1")),
						},
					},
				},
				Hostnames: []v1alpha2.Hostname{
					v1alpha2.Hostname("foo.example.com"),
				},
				Rules: []v1alpha2.HTTPRouteRule{
					{
						Matches: []v1alpha2.HTTPRouteMatch{
							{
								Path: &v1alpha2.HTTPPathMatch{
									Value: helpers.GetStringPointer("/"),
								},
							},
						},
					},
				},
			},
		}
	}
	hr1 := createRoute("hr-1", "gateway")
	hr2 := createRoute("hr-2", "wrong-gateway")

	store := &store{
		gw: &v1alpha2.Gateway{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test",
				Name:      "gateway",
			},
			Spec: v1alpha2.GatewaySpec{
				Listeners: []v1alpha2.Listener{
					{
						Name:     "listener-80-1",
						Hostname: nil,
						Port:     80,
						Protocol: v1alpha2.HTTPProtocolType,
					},
				},
			},
		},
		httpRoutes: map[types.NamespacedName]*v1alpha2.HTTPRoute{
			{Namespace: "test", Name: "hr-1"}: hr1,
			{Namespace: "test", Name: "hr-2"}: hr2,
		},
	}

	gwNsName := types.NamespacedName{Namespace: "test", Name: "gateway"}

	routeHR1 := &route{
		Source: hr1,
		ValidSectionNameRefs: map[string]struct{}{
			"listener-80-1": {},
		},
		InvalidSectionNameRefs: map[string]struct{}{},
	}
	expected := &graph{
		Listeners: map[string]*listener{
			"listener-80-1": {
				Source: store.gw.Spec.Listeners[0],
				Valid:  true,
				Routes: map[types.NamespacedName]*route{
					{Namespace: "test", Name: "hr-1"}: routeHR1,
				},
				AcceptedHostnames: map[string]struct{}{
					"foo.example.com": {},
				},
			},
		},
		Routes: map[types.NamespacedName]*route{
			{Namespace: "test", Name: "hr-1"}: routeHR1,
		},
	}

	result := buildGraph(store, gwNsName)
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("buildGraph() mismatch (-want +got):\n%s", diff)
	}
}

func TestBuildListeners(t *testing.T) {
	listener801 := v1alpha2.Listener{
		Name:     "listener-80-1",
		Hostname: (*v1alpha2.Hostname)(helpers.GetStringPointer("foo.example.com")),
		Port:     80,
		Protocol: v1alpha2.HTTPProtocolType,
	}
	listener802 := v1alpha2.Listener{
		Name:     "listener-80-2",
		Hostname: (*v1alpha2.Hostname)(helpers.GetStringPointer("bar.example.com")),
		Port:     80,
		Protocol: v1alpha2.TCPProtocolType, // invalid protocol
	}
	listener803 := v1alpha2.Listener{
		Name:     "listener-80-3",
		Hostname: (*v1alpha2.Hostname)(helpers.GetStringPointer("bar.example.com")),
		Port:     80,
		Protocol: v1alpha2.HTTPProtocolType,
	}
	listener804 := v1alpha2.Listener{
		Name:     "listener-80-4",
		Hostname: (*v1alpha2.Hostname)(helpers.GetStringPointer("foo.example.com")),
		Port:     80,
		Protocol: v1alpha2.HTTPProtocolType,
	}

	tests := []struct {
		gateway  *v1alpha2.Gateway
		expected map[string]*listener
		msg      string
	}{
		{
			gateway: &v1alpha2.Gateway{
				Spec: v1alpha2.GatewaySpec{
					Listeners: []v1alpha2.Listener{
						listener801,
					},
				},
			},
			expected: map[string]*listener{
				"listener-80-1": {
					Source:            listener801,
					Valid:             true,
					Routes:            map[types.NamespacedName]*route{},
					AcceptedHostnames: map[string]struct{}{},
				},
			},
			msg: "valid listener",
		},
		{
			gateway: &v1alpha2.Gateway{
				Spec: v1alpha2.GatewaySpec{
					Listeners: []v1alpha2.Listener{
						listener802,
					},
				},
			},
			expected: map[string]*listener{
				"listener-80-2": {
					Source:            listener802,
					Valid:             false,
					Routes:            map[types.NamespacedName]*route{},
					AcceptedHostnames: map[string]struct{}{},
				},
			},
			msg: "invalid listener",
		},
		{
			gateway: &v1alpha2.Gateway{
				Spec: v1alpha2.GatewaySpec{
					Listeners: []v1alpha2.Listener{
						listener801, listener803,
					},
				},
			},
			expected: map[string]*listener{
				"listener-80-1": {
					Source:            listener801,
					Valid:             true,
					Routes:            map[types.NamespacedName]*route{},
					AcceptedHostnames: map[string]struct{}{},
				},
				"listener-80-3": {
					Source:            listener803,
					Valid:             true,
					Routes:            map[types.NamespacedName]*route{},
					AcceptedHostnames: map[string]struct{}{},
				},
			},
			msg: "two valid Listeners",
		},
		{
			gateway: &v1alpha2.Gateway{
				Spec: v1alpha2.GatewaySpec{
					Listeners: []v1alpha2.Listener{
						listener801, listener804,
					},
				},
			},
			expected: map[string]*listener{
				"listener-80-1": {
					Source:            listener801,
					Valid:             false,
					Routes:            map[types.NamespacedName]*route{},
					AcceptedHostnames: map[string]struct{}{},
				},
				"listener-80-4": {
					Source:            listener804,
					Valid:             false,
					Routes:            map[types.NamespacedName]*route{},
					AcceptedHostnames: map[string]struct{}{},
				},
			},
			msg: "collision",
		},
		{
			gateway:  nil,
			expected: map[string]*listener{},
			msg:      "no gateway",
		},
	}

	for _, test := range tests {
		result := buildListeners(test.gateway)
		if diff := cmp.Diff(test.expected, result); diff != "" {
			t.Errorf("buildListeners() %q  mismatch (-want +got):\n%s", test.msg, diff)
		}
	}
}

func TestBindRouteToListeners(t *testing.T) {
	createRoute := func(hostname string, parentRefs ...v1alpha2.ParentRef) *v1alpha2.HTTPRoute {
		return &v1alpha2.HTTPRoute{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test",
				Name:      "hr-1",
			},
			Spec: v1alpha2.HTTPRouteSpec{
				CommonRouteSpec: v1alpha2.CommonRouteSpec{
					ParentRefs: parentRefs,
				},
				Hostnames: []v1alpha2.Hostname{
					v1alpha2.Hostname(hostname),
				},
			},
		}
	}

	hrNonExistingSectionName := createRoute("foo.example.com", v1alpha2.ParentRef{
		Namespace:   (*v1alpha2.Namespace)(helpers.GetStringPointer("test")),
		Name:        "gateway",
		SectionName: (*v1alpha2.SectionName)(helpers.GetStringPointer("listener-80-2")),
	})

	hrFoo := createRoute("foo.example.com", v1alpha2.ParentRef{
		Namespace:   (*v1alpha2.Namespace)(helpers.GetStringPointer("test")),
		Name:        "gateway",
		SectionName: (*v1alpha2.SectionName)(helpers.GetStringPointer("listener-80-1")),
	})

	hrBar := createRoute("bar.example.com", v1alpha2.ParentRef{
		Namespace:   (*v1alpha2.Namespace)(helpers.GetStringPointer("test")),
		Name:        "gateway",
		SectionName: (*v1alpha2.SectionName)(helpers.GetStringPointer("listener-80-1")),
	})

	// we create a new listener each time because the function under test can modify it
	createListener := func() *listener {
		return &listener{
			Source: v1alpha2.Listener{
				Hostname: (*v1alpha2.Hostname)(helpers.GetStringPointer("foo.example.com")),
			},
			Valid:             true,
			Routes:            map[types.NamespacedName]*route{},
			AcceptedHostnames: map[string]struct{}{},
		}
	}

	createModifiedListener := func(m func(*listener)) *listener {
		l := createListener()
		m(l)
		return l
	}

	tests := []struct {
		httpRoute         *v1alpha2.HTTPRoute
		gwNsName          types.NamespacedName
		listeners         map[string]*listener
		expectedIgnored   bool
		expectedRoute     *route
		expectedListeners map[string]*listener
		msg               string
	}{
		{
			httpRoute: createRoute("foo.example.com"),
			gwNsName:  types.NamespacedName{Namespace: "test", Name: "gateway"},
			listeners: map[string]*listener{
				"listener-80-1": createListener(),
			},
			expectedIgnored: true,
			expectedRoute:   nil,
			expectedListeners: map[string]*listener{
				"listener-80-1": createListener(),
			},
			msg: "HTTPRoute without parent refs",
		},
		{
			httpRoute: createRoute("foo.example.com", v1alpha2.ParentRef{
				Namespace:   (*v1alpha2.Namespace)(helpers.GetStringPointer("test")),
				Name:        "some-gateway", // wrong gateway
				SectionName: (*v1alpha2.SectionName)(helpers.GetStringPointer("listener-1")),
			}),
			gwNsName: types.NamespacedName{Namespace: "test", Name: "gateway"},
			listeners: map[string]*listener{
				"listener-80-1": createListener(),
			},
			expectedIgnored: true,
			expectedRoute:   nil,
			expectedListeners: map[string]*listener{
				"listener-80-1": createListener(),
			},
			msg: "HTTPRoute without good parent refs",
		},
		{
			httpRoute: hrNonExistingSectionName,
			gwNsName:  types.NamespacedName{Namespace: "test", Name: "gateway"},
			listeners: map[string]*listener{
				"listener-80-1": createListener(),
			},
			expectedIgnored: false,
			expectedRoute: &route{
				Source:               hrNonExistingSectionName,
				ValidSectionNameRefs: map[string]struct{}{},
				InvalidSectionNameRefs: map[string]struct{}{
					"listener-80-2": {},
				},
			},
			expectedListeners: map[string]*listener{
				"listener-80-1": createListener(),
			},
			msg: "HTTPRoute with non-existing section name",
		},
		{
			httpRoute: hrFoo,
			gwNsName:  types.NamespacedName{Namespace: "test", Name: "gateway"},
			listeners: map[string]*listener{
				"listener-80-1": createListener(),
			},
			expectedIgnored: false,
			expectedRoute: &route{
				Source: hrFoo,
				ValidSectionNameRefs: map[string]struct{}{
					"listener-80-1": {},
				},
				InvalidSectionNameRefs: map[string]struct{}{},
			},
			expectedListeners: map[string]*listener{
				"listener-80-1": createModifiedListener(func(l *listener) {
					l.Routes = map[types.NamespacedName]*route{
						{Namespace: "test", Name: "hr-1"}: {
							Source: hrFoo,
							ValidSectionNameRefs: map[string]struct{}{
								"listener-80-1": {},
							},
							InvalidSectionNameRefs: map[string]struct{}{},
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
			httpRoute: hrBar,
			gwNsName:  types.NamespacedName{Namespace: "test", Name: "gateway"},
			listeners: map[string]*listener{
				"listener-80-1": createListener(),
			},
			expectedIgnored: false,
			expectedRoute: &route{
				Source:               hrBar,
				ValidSectionNameRefs: map[string]struct{}{},
				InvalidSectionNameRefs: map[string]struct{}{
					"listener-80-1": {},
				},
			},
			expectedListeners: map[string]*listener{
				"listener-80-1": createListener(),
			},
			msg: "HTTPRoute with zero accepted hostnames",
		},
	}

	for _, test := range tests {
		ignored, route := bindHTTPRouteToListeners(test.httpRoute, test.gwNsName, test.listeners)
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
	var listenerHostnameFoo v1alpha2.Hostname = "foo.example.com"
	var listenerHostnameCafe v1alpha2.Hostname = "cafe.example.com"
	routeHostnames := []v1alpha2.Hostname{"foo.example.com", "bar.example.com"}

	tests := []struct {
		listenerHostname *v1alpha2.Hostname
		routeHostnames   []v1alpha2.Hostname
		expected         []string
		msg              string
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

func TestIgnoreParentRef(t *testing.T) {
	var gatewayNs v1alpha2.Namespace = "test"
	var sectionName v1alpha2.SectionName = "first"
	var emptySectionName v1alpha2.SectionName

	tests := []struct {
		parentRef   v1alpha2.ParentRef
		hrNamespace string
		gwNsName    types.NamespacedName
		expected    bool
		msg         string
	}{
		{
			parentRef: v1alpha2.ParentRef{
				Namespace:   &gatewayNs,
				Name:        "gateway",
				SectionName: &sectionName,
			},
			hrNamespace: "test2",
			gwNsName:    types.NamespacedName{Namespace: string(gatewayNs), Name: "gateway"},
			expected:    false,
			msg:         "normal case",
		},
		{
			parentRef: v1alpha2.ParentRef{
				Name:        "gateway",
				SectionName: &sectionName,
			},
			hrNamespace: "test",
			gwNsName:    types.NamespacedName{Namespace: string(gatewayNs), Name: "gateway"},
			expected:    false,
			msg:         "normal case with implicit namespace",
		},
		{
			parentRef: v1alpha2.ParentRef{
				Namespace:   &gatewayNs,
				Name:        "gateway",
				SectionName: &sectionName,
			},
			hrNamespace: "test2",
			gwNsName:    types.NamespacedName{Namespace: "test3", Name: "gateway"},
			expected:    true,
			msg:         "gateway namespace is different",
		},
		{
			parentRef: v1alpha2.ParentRef{
				Namespace:   &gatewayNs,
				Name:        "gateway",
				SectionName: &sectionName,
			},
			hrNamespace: "test2",
			gwNsName:    types.NamespacedName{Namespace: string(gatewayNs), Name: "gateway2"},
			expected:    true,
			msg:         "gateway name is different",
		},
		{
			parentRef: v1alpha2.ParentRef{
				Namespace:   &gatewayNs,
				Name:        "gateway",
				SectionName: nil,
			},
			hrNamespace: "test2",
			gwNsName:    types.NamespacedName{Namespace: string(gatewayNs), Name: "gateway"},
			expected:    true,
			msg:         "section name is nil",
		},
		{
			parentRef: v1alpha2.ParentRef{
				Namespace:   &gatewayNs,
				Name:        "gateway",
				SectionName: &emptySectionName,
			},
			hrNamespace: "test2",
			gwNsName:    types.NamespacedName{Namespace: string(gatewayNs), Name: "gateway"},
			expected:    true,
			msg:         "section name is empty",
		},
	}

	for _, test := range tests {
		result := ignoreParentRef(test.parentRef, test.hrNamespace, test.gwNsName)
		if result != test.expected {
			t.Errorf("ignoreParentRef() returned %v but expected %v for the case of %q", result, test.expected, test.msg)
		}
	}
}

func TestValidateListener(t *testing.T) {
	tests := []struct {
		l        v1alpha2.Listener
		expected bool
		msg      string
	}{
		{
			l: v1alpha2.Listener{
				Port:     80,
				Protocol: v1alpha2.HTTPProtocolType,
			},
			expected: true,
			msg:      "valid",
		},
		{
			l: v1alpha2.Listener{
				Port:     81,
				Protocol: v1alpha2.HTTPProtocolType,
			},
			expected: false,
			msg:      "invalid port",
		},
		{
			l: v1alpha2.Listener{
				Port:     80,
				Protocol: v1alpha2.TCPProtocolType,
			},
			expected: false,
			msg:      "invalid protocol",
		},
	}

	for _, test := range tests {
		result := validateListener(test.l)
		if result != test.expected {
			t.Errorf("validateListener() returned %v but expected %v for the case of %q", result, test.expected, test.msg)
		}
	}
}

func TestGetHostname(t *testing.T) {
	var emptyHostname v1alpha2.Hostname
	var hostname v1alpha2.Hostname = "example.com"

	tests := []struct {
		h        *v1alpha2.Hostname
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
