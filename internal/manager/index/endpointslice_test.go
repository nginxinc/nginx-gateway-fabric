package index

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/api/core/v1"
	discoveryV1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestServiceNameIndexFunc(t *testing.T) {
	testcases := []struct {
		msg       string
		obj       client.Object
		expOutput []string
	}{
		{
			msg: "normal case",
			obj: &discoveryV1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{KubernetesServiceNameLabel: "test-svc"},
				},
			},
			expOutput: []string{"test-svc"},
		},
		{
			msg:       "nil labels",
			obj:       &discoveryV1.EndpointSlice{},
			expOutput: nil,
		},
		{
			msg: "no service-name label",
			obj: &discoveryV1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					Labels: make(map[string]string),
				},
			},
			expOutput: nil,
		},
	}

	for _, tc := range testcases {
		output := serviceNameIndexFunc(tc.obj)
		if diff := cmp.Diff(tc.expOutput, output); diff != "" {
			t.Errorf("serviceNameIndexFunc() mismatch on %q (-want +got):\n%s", tc.msg, diff)
		}
	}
}

func TestServiceNameIndexFuncPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("serviceNameIndexFunc() did not panic")
		}
	}()

	serviceNameIndexFunc(&v1.Namespace{})
}
