package state

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/helpers"
)

func TestSort(t *testing.T) {
	// timestamps
	earlier := metav1.Now()
	later := metav1.NewTime(earlier.Add(1 * time.Second))

	// matches
	pathOnlyMatch := v1alpha2.HTTPRouteMatch{
		Path: &v1alpha2.HTTPPathMatch{
			Value: helpers.GetStringPointer("/path"), // path match only (low priority)
		},
	}
	twoHeaderMatch := v1alpha2.HTTPRouteMatch{
		Path: &v1alpha2.HTTPPathMatch{
			Value: helpers.GetStringPointer("/path"),
		},
		Headers: []v1alpha2.HTTPHeaderMatch{
			{
				Name:  "header1",
				Value: "value1",
			},
			{
				Name:  "header2",
				Value: "value2",
			},
		},
	}
	threeHeaderMatch := v1alpha2.HTTPRouteMatch{
		Path: &v1alpha2.HTTPPathMatch{
			Value: helpers.GetStringPointer("/path"),
		},
		Headers: []v1alpha2.HTTPHeaderMatch{
			{
				Name:  "header1",
				Value: "value1",
			},
			{
				Name:  "header2",
				Value: "value2",
			},
			{
				Name:  "header3",
				Value: "value3",
			},
		},
	}
	twoHeaderOneParamMatch := v1alpha2.HTTPRouteMatch{
		Path: &v1alpha2.HTTPPathMatch{
			Value: helpers.GetStringPointer("/path"),
		},
		Headers: []v1alpha2.HTTPHeaderMatch{
			{
				Name:  "header1",
				Value: "value1",
			},
			{
				Name:  "header2",
				Value: "value2",
			},
		},
		QueryParams: []v1alpha2.HTTPQueryParamMatch{
			{
				Name:  "key1",
				Value: "value1",
			},
		},
	}

	hr1 := v1alpha2.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "hr1",
			Namespace:         "test",
			CreationTimestamp: earlier,
		},
		Spec: v1alpha2.HTTPRouteSpec{
			Rules: []v1alpha2.HTTPRouteRule{
				{
					Matches: []v1alpha2.HTTPRouteMatch{pathOnlyMatch},
				},
				{
					Matches: []v1alpha2.HTTPRouteMatch{twoHeaderMatch},
				},
				{
					Matches: []v1alpha2.HTTPRouteMatch{
						twoHeaderOneParamMatch, // tie decided on params
						threeHeaderMatch,       // tie decided on headers
					},
				},
			},
		},
	}

	hr2 := v1alpha2.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "hr2",
			Namespace:         "test",
			CreationTimestamp: later,
		},
		Spec: v1alpha2.HTTPRouteSpec{
			Rules: []v1alpha2.HTTPRouteRule{
				{
					Matches: []v1alpha2.HTTPRouteMatch{twoHeaderMatch}, // tie decided on creation timestamp
				},
			},
		},
	}

	hr3 := v1alpha2.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "hr3",
			Namespace:         "a-test", // tie decided by namespace name
			CreationTimestamp: later,
		},
		Spec: v1alpha2.HTTPRouteSpec{
			Rules: []v1alpha2.HTTPRouteRule{
				{
					Matches: []v1alpha2.HTTPRouteMatch{twoHeaderMatch},
				},
			},
		},
	}

	routes := []MatchRule{
		{
			MatchIdx: 0, // pathOnlyMatch
			RuleIdx:  0,
			Source:   &hr1,
		},
		{
			MatchIdx: 0, // twoHeaderMatch / earlier timestamp
			RuleIdx:  1,
			Source:   &hr1,
		},
		{
			MatchIdx: 0, // twoHeaderOneParamMatch
			RuleIdx:  2,
			Source:   &hr1,
		},
		{
			MatchIdx: 1, // threeHeaderMatch
			RuleIdx:  2,
			Source:   &hr1,
		},
		{
			MatchIdx: 0, // twoHeaderMatch / later timestamp / test/hr2
			RuleIdx:  0,
			Source:   &hr2,
		},
		{
			MatchIdx: 0, // twoHeaderMatch / later timestamp / a-test/hr3
			RuleIdx:  0,
			Source:   &hr3,
		},
	}

	sortedRoutes := []MatchRule{
		{
			MatchIdx: 1, // threeHeaderMatch
			RuleIdx:  2,
			Source:   &hr1,
		},
		{
			MatchIdx: 0, // twoHeaderOneParamMatch
			RuleIdx:  2,
			Source:   &hr1,
		},
		{
			MatchIdx: 0, // twoHeaderMatch / earlier timestamp
			RuleIdx:  1,
			Source:   &hr1,
		},
		{
			MatchIdx: 0, // twoHeaderMatch / later timestamp / a-test/hr3
			RuleIdx:  0,
			Source:   &hr3,
		},
		{
			MatchIdx: 0, // twoHeaderMatch / later timestamp / test/hr2
			RuleIdx:  0,
			Source:   &hr2,
		},
		{
			MatchIdx: 0, // pathOnlyMatch
			RuleIdx:  0,
			Source:   &hr1,
		},
	}

	sortMatchRules(routes)

	if diff := cmp.Diff(sortedRoutes, routes); diff != "" {
		t.Errorf("sortMatchRules() mismatch (-want +got):\n%s", diff)
	}
}
