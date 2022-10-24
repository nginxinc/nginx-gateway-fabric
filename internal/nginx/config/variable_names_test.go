package config

import "testing"

func TestConvertStringToSafeVariableName(t *testing.T) {
	tests := []struct {
		msg      string
		s        string
		expected string
	}{
		{
			msg:      "no hyphens",
			s:        "foo",
			expected: "foo",
		},
		{
			msg:      "hyphens",
			s:        "foo-bar-baz",
			expected: "foo_bar_baz",
		},
	}
	for _, test := range tests {
		if result := convertStringToSafeVariableName(test.s); result != test.expected {
			t.Errorf(
				"convertStringToSafeVariableName() mismatch for test %q; expected %s, got %s",
				test.msg,
				test.expected,
				result,
			)
		}
	}
}
