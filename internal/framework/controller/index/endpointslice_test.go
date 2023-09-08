package index

import (
	"testing"

	. "github.com/onsi/gomega"
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
		g := NewGomegaWithT(t)
		output := ServiceNameIndexFunc(tc.obj)
		g.Expect(output).To(Equal(tc.expOutput))
	}
}

func TestServiceNameIndexFuncPanics(t *testing.T) {
	defer func() {
		g := NewGomegaWithT(t)
		g.Expect(recover()).ShouldNot(BeNil())
	}()

	ServiceNameIndexFunc(&v1.Namespace{})
}
