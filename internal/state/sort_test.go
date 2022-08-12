package state

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/helpers"
)

func TestSort(t *testing.T) {
	// timestamps
	earlier := metav1.Now()
	later := metav1.NewTime(earlier.Add(1 * time.Second))

	// matches
	pathOnlyMatch := v1beta1.HTTPRouteMatch{
		Path: &v1beta1.HTTPPathMatch{
			Value: helpers.GetStringPointer("/path"), // path match only (low priority)
		},
	}
	twoHeaderMatch := v1beta1.HTTPRouteMatch{
		Path: &v1beta1.HTTPPathMatch{
			Value: helpers.GetStringPointer("/path"),
		},
		Headers: []v1beta1.HTTPHeaderMatch{
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
	threeHeaderMatch := v1beta1.HTTPRouteMatch{
		Path: &v1beta1.HTTPPathMatch{
			Value: helpers.GetStringPointer("/path"),
		},
		Headers: []v1beta1.HTTPHeaderMatch{
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
	twoHeaderOneParamMatch := v1beta1.HTTPRouteMatch{
		Path: &v1beta1.HTTPPathMatch{
			Value: helpers.GetStringPointer("/path"),
		},
		Headers: []v1beta1.HTTPHeaderMatch{
			{
				Name:  "header1",
				Value: "value1",
			},
			{
				Name:  "header2",
				Value: "value2",
			},
		},
		QueryParams: []v1beta1.HTTPQueryParamMatch{
			{
				Name:  "key1",
				Value: "value1",
			},
		},
	}

	hr1 := v1beta1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "hr1",
			Namespace:         "test",
			CreationTimestamp: earlier,
		},
		Spec: v1beta1.HTTPRouteSpec{
			Rules: []v1beta1.HTTPRouteRule{
				{
					Matches: []v1beta1.HTTPRouteMatch{pathOnlyMatch},
				},
				{
					Matches: []v1beta1.HTTPRouteMatch{twoHeaderMatch},
				},
				{
					Matches: []v1beta1.HTTPRouteMatch{
						twoHeaderOneParamMatch, // tie decided on params
						threeHeaderMatch,       // tie decided on headers
					},
				},
			},
		},
	}

	hr2 := v1beta1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "hr2",
			Namespace:         "test",
			CreationTimestamp: later,
		},
		Spec: v1beta1.HTTPRouteSpec{
			Rules: []v1beta1.HTTPRouteRule{
				{
					Matches: []v1beta1.HTTPRouteMatch{twoHeaderMatch}, // tie decided on creation timestamp
				},
			},
		},
	}

	hr3 := v1beta1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "hr3",
			Namespace:         "a-test", // tie decided by namespace name
			CreationTimestamp: later,
		},
		Spec: v1beta1.HTTPRouteSpec{
			Rules: []v1beta1.HTTPRouteRule{
				{
					Matches: []v1beta1.HTTPRouteMatch{twoHeaderMatch},
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
