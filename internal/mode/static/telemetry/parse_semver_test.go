package telemetry

import (
	"errors"
	"testing"

	. "github.com/onsi/gomega"
)

func TestParseKubeletVersion(t *testing.T) {
	tests := []struct {
		expError error
		input    string
		expected string
		name     string
	}{
		{
			input:    "v1.27.9",
			expected: "1.27.9",
			name:     "normal case",
			expError: nil,
		},
		{
			input:    "   v1.27.9  ",
			expected: "1.27.9",
			name:     "removes added whitespace",
			expError: nil,
		},
		{
			input:    "v1.27",
			expected: "1.27.0",
			name:     "adds appended 0's if missing semver patch number",
			expError: nil,
		},
		{
			input:    "v1.27.8-gke.1067004",
			expected: "1.27.8",
			name:     "removes trailing characters from semver version",
			expError: nil,
		},
		{
			input:    "v1.27.9+",
			expected: "1.27.9",
			name:     "removes trailing characters from semver version no following characters",
			expError: nil,
		},
		{
			input:    "v1.27-gke.1067004",
			expected: "1.27.0",
			name:     "removes trailing characters from semver version no patch version",
			expError: nil,
		},
		{
			input:    "v1.27.",
			expected: "1.27.0",
			name:     "edge case where patch version is missing but separating period is",
			expError: nil,
		},
		{
			input:    "v1.27.gke+323",
			expected: "1.27.0",
			name:     "edge case where patch version is missing but additional characters are present",
			expError: nil,
		},
		{
			input:    "",
			expected: "",
			name:     "error on empty string",
			expError: errors.New("string cannot be empty"),
		},
		{
			input:    "1",
			expected: "",
			name:     "errors when major and minor version are not present",
			expError: errors.New("string must have at least a major and minor version specified"),
		},
		{
			input:    "1.",
			expected: "",
			name:     "errors when major and minor version are not present",
			expError: errors.New("string must have at least a major and minor version specified"),
		},
		{
			input:    "123gke",
			expected: "",
			name:     "errors when string does not contain a number as the major version",
			expError: errors.New("string must have a number as the major version"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			result, err := parseSemver(test.input)
			g.Expect(result).To(Equal(test.expected))

			if test.expError != nil {
				g.Expect(err).To(MatchError(test.expError))
			} else {
				g.Expect(err).To(BeNil())
			}
		})
	}
}
