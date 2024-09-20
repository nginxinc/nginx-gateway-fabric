package config

import (
	"strings"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/dataplane"
)

func TestExecuteVersion(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	expSubStrings := map[string]int{
		"return 200 42;": 1,
	}

	cfg := dataplane.Configuration{
		Version: 42,
	}

	maps := string(executeVersion(cfg)[0].data)
	for expSubStr, expCount := range expSubStrings {
		g.Expect(expCount).To(Equal(strings.Count(maps, expSubStr)))
	}
}
