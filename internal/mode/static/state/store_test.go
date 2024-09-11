package state

import (
	"testing"

	. "github.com/onsi/gomega"
	discoveryV1 "k8s.io/api/discovery/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	v1 "sigs.k8s.io/gateway-api/apis/v1"
)

//nolint:paralleltest,tparallel // Order matters for these tests.
func TestSetChangeType(t *testing.T) {
	t.Parallel()
	ctu := newChangeTrackingUpdater(nil, nil)

	// Order matters for these cases.
	tests := []struct {
		obj     client.Object
		name    string
		exp     ChangeType
		changed bool
	}{
		{
			name: "no change",
			exp:  NoChange,
		},
		{
			name:    "endpoint object",
			obj:     &discoveryV1.EndpointSlice{},
			changed: true,
			exp:     EndpointsOnlyChange,
		},
		{
			name:    "non-endpoint object",
			obj:     &v1.HTTPRoute{},
			changed: true,
			exp:     ClusterStateChange,
		},
		{
			name:    "changeType was previously set to ClusterStateChange",
			obj:     &discoveryV1.EndpointSlice{},
			changed: true,
			exp:     ClusterStateChange,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)
			ctu.setChangeType(test.obj, test.changed)
			g.Expect(ctu.changeType).To(Equal(test.exp))
		})
	}
}
