package filter

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"
)

func TestCreateFilterForGatewayClass(t *testing.T) {
	const gcName = "my-gc"

	filter := CreateFilterForGatewayClass(gcName)
	if filter == nil {
		t.Fatal("CreateFilterForGatewayClass() returned nil")
	}

	tests := []struct {
		nsname   types.NamespacedName
		expected bool
	}{
		{
			nsname:   types.NamespacedName{Name: gcName},
			expected: true,
		},
		{
			nsname:   types.NamespacedName{Name: gcName, Namespace: "doesn't matter"},
			expected: true,
		},
		{
			nsname:   types.NamespacedName{Name: "some-gc"},
			expected: false,
		},
	}

	for _, test := range tests {
		result, msg := filter(test.nsname)

		if result != test.expected {
			t.Errorf("filter(%#v) returned %v but expected %v", test.nsname, result, test.expected)
		}

		if result && msg != "" {
			t.Errorf("filter(%#v) returned a non-empty message %q", test.nsname, msg)
		}
		if !result && msg == "" {
			t.Errorf("filter(%#v) returned an empty message", test.nsname)
		}
	}
}
