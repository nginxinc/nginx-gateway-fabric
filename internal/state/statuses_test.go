package state

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/types"
)

func TestBuildStatuses(t *testing.T) {
	g := &graph{
		Listeners: map[string]*listener{
			"listener-80-1": {
				Valid: true,
				Routes: map[types.NamespacedName]*route{
					{Namespace: "test", Name: "hr-1"}: {}},
			},
		},
		Routes: map[types.NamespacedName]*route{
			{Namespace: "test", Name: "hr-1"}: {
				ValidSectionNameRefs: map[string]struct{}{
					"listener-80-1": {},
				},
				InvalidSectionNameRefs: map[string]struct{}{
					"listener-80-2": {},
				},
			},
		},
	}

	expected := Statuses{
		ListenerStatuses: map[string]ListenerStatus{
			"listener-80-1": {
				Valid:          true,
				AttachedRoutes: 1,
			},
		},
		HTTPRouteStatuses: map[types.NamespacedName]HTTPRouteStatus{
			{Namespace: "test", Name: "hr-1"}: {
				ParentStatuses: map[string]ParentStatus{
					"listener-80-1": {
						Attached: true,
					},
					"listener-80-2": {
						Attached: false,
					},
				},
			},
		},
	}

	result := buildStatuses(g)
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("buildStatuses() mismatch (-want +got):\n%s", diff)
	}
}
