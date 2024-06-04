package config

import (
	"strings"
	"testing"

	. "github.com/onsi/gomega"
)

func TestExecuteVersion(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	expSubStrings := map[string]int{
		"return 200 42;": 1,
	}

	maps := string(executeVersion(42))
	for expSubStr, expCount := range expSubStrings {
		g.Expect(expCount).To(Equal(strings.Count(maps, expSubStr)))
	}
}
