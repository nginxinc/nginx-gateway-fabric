package graph

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
)

func TestBuildReferencedNamespaces(t *testing.T) {
	ns1 := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ns1",
		},
	}

	ns2 := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ns2",
			Labels: map[string]string{
				"apples": "oranges",
			},
		},
	}

	ns3 := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ns3",
			Labels: map[string]string{
				"peaches": "bananas",
			},
		},
	}

	clusterNamespaces := map[types.NamespacedName]*v1.Namespace{
		{Name: "ns1"}: ns1,
		{Name: "ns2"}: ns2,
		{Name: "ns3"}: ns3,
	}

	tests := []struct {
		gw            *Gateway
		expectedRefNS map[types.NamespacedName]*v1.Namespace
		name          string
	}{
		{
			gw: &Gateway{
				Listeners: map[string]*Listener{
					"listener-1": {
						Valid:                     true,
						AllowedRouteLabelSelector: labels.SelectorFromSet(map[string]string{"apples": "oranges"}),
					},
				},
				Valid: true,
			},
			expectedRefNS: map[types.NamespacedName]*v1.Namespace{
				{Name: "ns2"}: ns2,
			},
			name: "gateway matches labels with one namespace",
		},
		{
			gw: &Gateway{
				Listeners: map[string]*Listener{},
				Valid:     true,
			},
			expectedRefNS: nil,
			name:          "gateway has no Listeners",
		},
		{
			gw: &Gateway{
				Listeners: map[string]*Listener{
					"listener-1": {
						Valid:                     true,
						AllowedRouteLabelSelector: labels.SelectorFromSet(map[string]string{"apples": "oranges"}),
					},
					"listener-2": {
						Valid:                     true,
						AllowedRouteLabelSelector: labels.SelectorFromSet(map[string]string{"peaches": "bananas"}),
					},
				},
				Valid: true,
			},
			expectedRefNS: map[types.NamespacedName]*v1.Namespace{
				{Name: "ns2"}: ns2,
				{Name: "ns3"}: ns3,
			},
			name: "gateway matches labels with two namespaces",
		},
		{
			gw: &Gateway{
				Listeners: map[string]*Listener{
					"listener-1": {
						Valid:                     true,
						AllowedRouteLabelSelector: labels.SelectorFromSet(map[string]string{"not": "matching"}),
					},
				},
				Valid: true,
			},

			expectedRefNS: nil,
			name:          "gateway doesn't match labels with any namespace",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(buildReferencedNamespaces(clusterNamespaces, test.gw)).To(Equal(test.expectedRefNS))
		})
	}
}
