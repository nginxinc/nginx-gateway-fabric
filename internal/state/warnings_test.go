package state

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
)

func TestAddWarningf(t *testing.T) {
	warnings := newWarnings()
	obj := &v1beta1.HTTPRoute{}

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
	obj := &v1beta1.HTTPRoute{}

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

func TestAdd(t *testing.T) {
	obj1 := &v1beta1.HTTPRoute{}
	obj2 := &v1beta1.HTTPRoute{}
	obj3 := &v1beta1.HTTPRoute{}

	tests := []struct {
		warnings      Warnings
		addedWarnings Warnings
		expected      Warnings
		msg           string
	}{
		{
			warnings:      newWarnings(),
			addedWarnings: newWarnings(),
			expected:      newWarnings(),
			msg:           "empty warnings",
		},
		{
			warnings: Warnings{
				obj1: []string{
					"first",
				},
			},
			addedWarnings: newWarnings(),
			expected: Warnings{
				obj1: []string{
					"first",
				},
			},
			msg: "empty added warnings",
		},
		{
			warnings: newWarnings(),
			addedWarnings: Warnings{
				obj1: []string{
					"first",
				},
			},
			expected: Warnings{
				obj1: []string{
					"first",
				},
			},
			msg: "empty warnings",
		},
		{
			warnings: Warnings{
				obj1: []string{
					"first 1",
				},
				obj3: []string{
					"first 3",
				},
			},
			addedWarnings: Warnings{
				obj2: []string{
					"first 2",
				},
				obj3: []string{
					"second 3",
				},
			},
			expected: Warnings{
				obj1: []string{
					"first 1",
				},
				obj2: []string{
					"first 2",
				},
				obj3: []string{
					"first 3",
					"second 3",
				},
			},
			msg: "adding and merging",
		},
	}

	for _, test := range tests {
		test.warnings.Add(test.addedWarnings)
		if diff := cmp.Diff(test.expected, test.warnings); diff != "" {
			t.Errorf("Add() %q mismatch (-want +got):\n%s", test.msg, diff)
		}
	}
}
