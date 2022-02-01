package state

import (
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func sortRoutes(routes []Route) {
	// stable sort is used so that the order of matches (as defined in each HTTPRoute rule) is preserved
	// this is important, because the winning match is the first match to win.
	sort.SliceStable(routes, func(i, j int) bool {
		return lessObjectMeta(&routes[i].Source.ObjectMeta, &routes[j].Source.ObjectMeta)
	})
}

func lessObjectMeta(meta1 *metav1.ObjectMeta, meta2 *metav1.ObjectMeta) bool {
	if meta1.CreationTimestamp.Equal(&meta2.CreationTimestamp) {
		return getResourceKey(meta1) < getResourceKey(meta2)
	}

	return meta1.CreationTimestamp.Before(&meta2.CreationTimestamp)
}

type mapper interface {
	Keys() []string
}

func getSortedKeys(m mapper) []string {
	keys := m.Keys()
	sort.Strings(keys)

	return keys
}
