package config

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
)

func TestAddWarningf(t *testing.T) {
	warnings := newWarnings()
	obj := &v1alpha2.HTTPRoute{}

	expected := Warnings{
		obj: []string{
			"simple",
			"advanced 1",
		},
	}

	warnings.AddWarningf(obj, "simple")
	warnings.AddWarningf(obj, "advanced %d", 1)

	if diff := cmp.Diff(expected, warnings); diff != "" {
		t.Errorf("AddWarningf mismatch (-want +got):\n%s", diff)
	}
}

func TestAddWarning(t *testing.T) {
	warnings := newWarnings()
	obj := &v1alpha2.HTTPRoute{}

	expected := Warnings{
		obj: []string{
			"first",
			"second",
		},
	}

	warnings.AddWarning(obj, "first")
	warnings.AddWarning(obj, "second")

	if diff := cmp.Diff(expected, warnings); diff != "" {
		t.Errorf("AddWarning mismatch (-want +got):\n%s", diff)
	}
}
