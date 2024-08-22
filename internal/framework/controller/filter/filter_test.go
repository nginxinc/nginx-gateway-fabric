package filter

import (
	"testing"

	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
)

func TestCreateSingleResourceFilter(t *testing.T) {
	t.Parallel()
	targetNsName := types.NamespacedName{Namespace: "test", Name: "resource"}

	g := NewWithT(t)
	filter := CreateSingleResourceFilter(targetNsName)
	g.Expect(filter).ToNot(BeNil())

	const expectedMsg = "Resource is ignored because this controller only supports a single resource " +
		"test/resource of that type"

	tests := []struct {
		name                  string
		nsname                types.NamespacedName
		expectedMsg           string
		expectedShouldProcess bool
	}{
		{
			name:                  "match",
			nsname:                targetNsName,
			expectedShouldProcess: true,
			expectedMsg:           "",
		},
		{
			name:                  "wrong namespace",
			nsname:                types.NamespacedName{Namespace: targetNsName.Namespace, Name: "other-name"},
			expectedShouldProcess: false,
			expectedMsg:           expectedMsg,
		},
		{
			name:                  "wrong name",
			nsname:                types.NamespacedName{Namespace: "other-ns", Name: targetNsName.Name},
			expectedShouldProcess: false,
			expectedMsg:           expectedMsg,
		},
		{
			name:                  "wrong namespace and name",
			nsname:                types.NamespacedName{Namespace: "other-ns", Name: "other-name"},
			expectedShouldProcess: false,
			expectedMsg:           expectedMsg,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)
			shouldProcess, msg := filter(test.nsname)
			g.Expect(shouldProcess).To(Equal(test.expectedShouldProcess))
			g.Expect(msg).To(Equal(test.expectedMsg))
		})
	}
}
