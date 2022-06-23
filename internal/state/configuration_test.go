package state

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/helpers"
)

func TestBuildConfiguration(t *testing.T) {
	createRoute := func(name string, hostname string, paths ...string) *v1alpha2.HTTPRoute {
		rules := make([]v1alpha2.HTTPRouteRule, 0, len(paths))
		for _, p := range paths {
			rules = append(rules, v1alpha2.HTTPRouteRule{
				Matches: []v1alpha2.HTTPRouteMatch{
					{
						Path: &v1alpha2.HTTPPathMatch{
							Value: helpers.GetStringPointer(p),
						},
					},
				},
			})
		}
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
							Name:        "gateway",
							SectionName: (*v1alpha2.SectionName)(helpers.GetStringPointer("listener-80-1")),
						},
					},
				},
				Hostnames: []v1alpha2.Hostname{
					v1alpha2.Hostname(hostname),
				},
				Rules: rules,
			},
		}
	}

	hr1 := createRoute("hr-1", "foo.example.com", "/")

	routeHR1 := &route{
		Source: hr1,
		ValidSectionNameRefs: map[string]struct{}{
			"listener-80-1": {},
		},
		InvalidSectionNameRefs: map[string]struct{}{},
	}

	hr2 := createRoute("hr-2", "bar.example.com", "/")

	routeHR2 := &route{
		Source: hr2,
		ValidSectionNameRefs: map[string]struct{}{
			"listener-80-1": {},
		},
		InvalidSectionNameRefs: map[string]struct{}{},
	}

	hr3 := createRoute("hr-3", "foo.example.com", "/", "/third")

	routeHR3 := &route{
		Source: hr3,
		ValidSectionNameRefs: map[string]struct{}{
			"listener-80-1": {},
		},
		InvalidSectionNameRefs: map[string]struct{}{},
	}

	hr4 := createRoute("hr-4", "foo.example.com", "/fourth", "/")

	routeHR4 := &route{
		Source: hr4,
		ValidSectionNameRefs: map[string]struct{}{
			"listener-80-1": {},
		},
		InvalidSectionNameRefs: map[string]struct{}{},
	}

	tests := []struct {
		graph    *graph
		expected Configuration
		msg      string
	}{
		{
			graph: &graph{
				Listeners: map[string]*listener{},
				Routes:    map[types.NamespacedName]*route{},
			},
			expected: Configuration{
				HTTPServers: []HTTPServer{},
			},
			msg: "empty graph",
		},
		{
			graph: &graph{
				Listeners: map[string]*listener{
					"listener-80-1": {
						Valid:             true,
						Routes:            map[types.NamespacedName]*route{},
						AcceptedHostnames: map[string]struct{}{},
					},
				},
				Routes: map[types.NamespacedName]*route{},
			},
			expected: Configuration{
				HTTPServers: []HTTPServer{},
			},
			msg: "listener with no routes",
		},
		{
			graph: &graph{
				Listeners: map[string]*listener{
					"listener-80-1": {
						Valid: true,
						Routes: map[types.NamespacedName]*route{
							{Namespace: "test", Name: "hr-1"}: routeHR1,
							{Namespace: "test", Name: "hr-2"}: routeHR2,
						},
						AcceptedHostnames: map[string]struct{}{
							"foo.example.com": {},
							"bar.example.com": {},
						},
					},
				},
				Routes: map[types.NamespacedName]*route{
					{Namespace: "test", Name: "hr-1"}: routeHR1,
					{Namespace: "test", Name: "hr-2"}: routeHR2,
				},
			},
			expected: Configuration{
				HTTPServers: []HTTPServer{
					{
						Hostname: "bar.example.com",
						PathRules: []PathRule{
							{
								Path: "/",
								MatchRules: []MatchRule{
									{
										MatchIdx: 0,
										RuleIdx:  0,
										Source:   hr2,
									},
								},
							},
						},
					},
					{
						Hostname: "foo.example.com",
						PathRules: []PathRule{
							{
								Path: "/",
								MatchRules: []MatchRule{
									{
										MatchIdx: 0,
										RuleIdx:  0,
										Source:   hr1,
									},
								},
							},
						},
					},
				},
			},
			msg: "one listener with two routes for different hostnames",
		},
		{
			graph: &graph{
				Listeners: map[string]*listener{
					"listener-80-1": {
						Valid: true,
						Routes: map[types.NamespacedName]*route{
							{Namespace: "test", Name: "hr-3"}: routeHR3,
							{Namespace: "test", Name: "hr-4"}: routeHR4,
						},
						AcceptedHostnames: map[string]struct{}{
							"foo.example.com": {},
						},
					},
				},
				Routes: map[types.NamespacedName]*route{
					{Namespace: "test", Name: "hr-3"}: routeHR3,
					{Namespace: "test", Name: "hr-4"}: routeHR4,
				},
			},
			expected: Configuration{
				HTTPServers: []HTTPServer{
					{
						Hostname: "foo.example.com",
						PathRules: []PathRule{
							{
								Path: "/",
								MatchRules: []MatchRule{
									{
										MatchIdx: 0,
										RuleIdx:  0,
										Source:   hr3,
									},
									{
										MatchIdx: 0,
										RuleIdx:  1,
										Source:   hr4,
									},
								},
							},
							{
								Path: "/fourth",
								MatchRules: []MatchRule{
									{
										MatchIdx: 0,
										RuleIdx:  0,
										Source:   hr4,
									},
								},
							},
							{
								Path: "/third",
								MatchRules: []MatchRule{
									{
										MatchIdx: 0,
										RuleIdx:  1,
										Source:   hr3,
									},
								},
							},
						},
					},
				},
			},
			msg: "one listener with two routes with the same hostname with and without collisions",
		},
	}

	for _, test := range tests {
		result := buildConfiguration(test.graph)
		if diff := cmp.Diff(test.expected, result); diff != "" {
			t.Errorf("buildConfiguration() %q mismatch (-want +got):\n%s", test.msg, diff)
		}
	}
}

func TestLessObjectMeta(t *testing.T) {
	sooner := metav1.Now()
	later := metav1.NewTime(sooner.Add(10 * time.Millisecond))

	tests := []struct {
		meta1, meta2 *metav1.ObjectMeta
		expected     bool
		msg          string
	}{
		{
			meta1: &metav1.ObjectMeta{
				CreationTimestamp: sooner,
				Namespace:         "test",
				Name:              "myname",
			},
			meta2: &metav1.ObjectMeta{
				CreationTimestamp: sooner,
				Namespace:         "test",
				Name:              "myname",
			},
			expected: false,
			msg:      "equal",
		},
		{
			meta1: &metav1.ObjectMeta{
				CreationTimestamp: sooner,
				Namespace:         "test",
				Name:              "myname",
			},
			meta2: &metav1.ObjectMeta{
				CreationTimestamp: later,
				Namespace:         "test",
				Name:              "myname",
			},
			expected: true,
			msg:      "less by timestamp",
		},
		{
			meta1: &metav1.ObjectMeta{
				CreationTimestamp: later,
				Namespace:         "test",
				Name:              "myname",
			},
			meta2: &metav1.ObjectMeta{
				CreationTimestamp: sooner,
				Namespace:         "test",
				Name:              "myname",
			},
			expected: false,
			msg:      "greater by timestamp",
		},
		{
			meta1: &metav1.ObjectMeta{
				CreationTimestamp: sooner,
				Namespace:         "atest",
				Name:              "myname",
			},
			meta2: &metav1.ObjectMeta{
				CreationTimestamp: sooner,
				Namespace:         "test",
				Name:              "myname",
			},
			expected: true,
			msg:      "less by namespace",
		},
		{
			meta1: &metav1.ObjectMeta{
				CreationTimestamp: sooner,
				Namespace:         "test",
				Name:              "myname",
			},
			meta2: &metav1.ObjectMeta{
				CreationTimestamp: sooner,
				Namespace:         "atest",
				Name:              "myname",
			},
			expected: false,
			msg:      "greater by namespace",
		},
		{
			meta1: &metav1.ObjectMeta{
				CreationTimestamp: sooner,
				Namespace:         "test",
				Name:              "amyname",
			},
			meta2: &metav1.ObjectMeta{
				CreationTimestamp: sooner,
				Namespace:         "test",
				Name:              "myname",
			},
			expected: true,
			msg:      "less by name",
		},
		{
			meta1: &metav1.ObjectMeta{
				CreationTimestamp: sooner,
				Namespace:         "test",
				Name:              "myname",
			},
			meta2: &metav1.ObjectMeta{
				CreationTimestamp: sooner,
				Namespace:         "test",
				Name:              "amyname",
			},
			expected: false,
			msg:      "greater by name",
		},
	}

	for _, test := range tests {
		result := lessObjectMeta(test.meta1, test.meta2)
		if result != test.expected {
			t.Errorf("lessObjectMeta() returned %v but expected %v for the case of %q", result, test.expected, test.msg)
		}
	}
}

func TestGetPath(t *testing.T) {
	tests := []struct {
		path     *v1alpha2.HTTPPathMatch
		expected string
		msg      string
	}{
		{
			path:     &v1alpha2.HTTPPathMatch{Value: helpers.GetStringPointer("/abc")},
			expected: "/abc",
			msg:      "normal case",
		},
		{
			path:     nil,
			expected: "/",
			msg:      "nil path",
		},
		{
			path:     &v1alpha2.HTTPPathMatch{Value: nil},
			expected: "/",
			msg:      "nil value",
		},
		{
			path:     &v1alpha2.HTTPPathMatch{Value: helpers.GetStringPointer("")},
			expected: "/",
			msg:      "empty value",
		},
	}

	for _, test := range tests {
		result := getPath(test.path)
		if result != test.expected {
			t.Errorf("getPath() returned %q but expected %q for the case of %q", result, test.expected, test.msg)
		}
	}
}

func TestMatchRuleGetMatch(t *testing.T) {
	var hr = &v1alpha2.HTTPRoute{
		Spec: v1alpha2.HTTPRouteSpec{
			Rules: []v1alpha2.HTTPRouteRule{
				{
					Matches: []v1alpha2.HTTPRouteMatch{
						{
							Path: &v1alpha2.HTTPPathMatch{
								Value: helpers.GetStringPointer("/path-1"),
							},
						},
						{
							Path: &v1alpha2.HTTPPathMatch{
								Value: helpers.GetStringPointer("/path-2"),
							},
						},
					},
				},
				{
					Matches: []v1alpha2.HTTPRouteMatch{
						{
							Path: &v1alpha2.HTTPPathMatch{
								Value: helpers.GetStringPointer("/path-3"),
							},
						},
						{
							Path: &v1alpha2.HTTPPathMatch{
								Value: helpers.GetStringPointer("/path-4"),
							},
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name,
		expPath string
		rule MatchRule
	}{
		{
			name:    "first match in first rule",
			expPath: "/path-1",
			rule:    MatchRule{MatchIdx: 0, RuleIdx: 0, Source: hr},
		},
		{
			name:    "second match in first rule",
			expPath: "/path-2",
			rule:    MatchRule{MatchIdx: 1, RuleIdx: 0, Source: hr},
		},
		{
			name:    "second match in second rule",
			expPath: "/path-4",
			rule:    MatchRule{MatchIdx: 1, RuleIdx: 1, Source: hr},
		},
	}

	for _, tc := range tests {
		actual := tc.rule.GetMatch()
		if *actual.Path.Value != tc.expPath {
			t.Errorf("MatchRule.GetMatch() returned incorrect match with path: %s, expected path: %s for test case: %q", *actual.Path.Value, tc.expPath, tc.name)
		}
	}
}
