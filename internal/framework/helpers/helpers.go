// Package helpers contains helper functions for unit tests.
package helpers

import (
	"fmt"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Diff prints the diff between two structs.
// It is useful in testing to compare two structs when they are large. In such a case, without Diff it will be difficult
// to pinpoint the difference between the two structs.
func Diff(want, got any) string {
	r := cmp.Diff(want, got)

	if r != "" {
		return "(-want +got)\n" + r
	}
	return r
}

// GetPointer takes a value of any type and returns a pointer to it.
func GetPointer[T any](v T) *T {
	return &v
}

// PrepareTimeForFakeClient processes the time similarly to the fake client
// from sigs.k8s.io/controller-runtime/pkg/client/fake
// making it is possible to use it in tests when comparing against values returned by the fake client.
// It panics if it fails to process the time.
func PrepareTimeForFakeClient(t metav1.Time) metav1.Time {
	bytes, err := t.Marshal()
	if err != nil {
		panic(fmt.Errorf("failed to marshal time: %w", err))
	}

	if err = t.Unmarshal(bytes); err != nil {
		panic(fmt.Errorf("failed to unmarshal time: %w", err))
	}

	return t
}
