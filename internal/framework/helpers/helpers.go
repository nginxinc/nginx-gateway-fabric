// Package helpers contains helper functions for unit tests.
package helpers

import (
	"fmt"
	"regexp"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "sigs.k8s.io/gateway-api/apis/v1"
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

var catchAllNonRootPathRegex = regexp.MustCompile(fmt.Sprintf(`^(%s?)(.*)`, rootPath))

const rootPath = "/"

func CreateMirrorPathWithBackendRef(path *string, backendRef v1.BackendObjectReference) *string {
	svcName := string(backendRef.Name)
	if backendRef.Namespace == nil {
		return CreateMirrorBackendPath(path, nil, &svcName)
	}

	return CreateMirrorBackendPath(path, (*string)(backendRef.Namespace), &svcName)
}

func CreateMirrorBackendPath(path *string, namespace *string, svcName *string) *string {
	var mirrorPath string
	mirrorPathLabel := "mirror"
	matches := catchAllNonRootPathRegex.FindStringSubmatch(*path)

	if len(matches) > 2 {
		trailingPath := matches[2]
		var mirrorPathSuffix string
		if len(trailingPath) > 0 {
			mirrorPathSuffix = fmt.Sprintf("%s-%s-%s", *svcName, mirrorPathLabel, trailingPath)
		} else {
			mirrorPathSuffix = fmt.Sprintf("%s-%s", *svcName, mirrorPathLabel)
		}

		if namespace != nil {
			mirrorPath = fmt.Sprintf("%s%s-%s", rootPath, *namespace, mirrorPathSuffix)
		} else {
			mirrorPath = fmt.Sprintf("%s%s", rootPath, mirrorPathSuffix)
		}
	}

	return &mirrorPath
}
