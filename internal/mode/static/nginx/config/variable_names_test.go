package config

import (
	"testing"

	. "github.com/onsi/gomega"
)

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
		g := NewWithT(t)
		g.Expect(convertStringToSafeVariableName(test.s)).To(Equal(test.expected))
	}
}

func TestGenerateAddHeaderMapVariableName(t *testing.T) {
	g := NewWithT(t)
	tests := []struct {
		msg        string
		headerName string
		expected   string
	}{
		{
			msg:        "no hyphens",
			headerName: "MyCoolHeader",
			expected:   "mycoolheader_header_var",
		},
		{
			msg:        "with hyphens",
			headerName: "My-Cool-Header",
			expected:   "my_cool_header_header_var",
		},
	}
	for _, tc := range tests {
		actual := generateAddHeaderMapVariableName(tc.headerName)
		g.Expect(actual).To(Equal(tc.expected))
	}
}
