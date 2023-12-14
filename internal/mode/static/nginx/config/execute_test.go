package config

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/config/http"
)

func TestExecute(t *testing.T) {
	g := NewWithT(t)
	defer func() {
		g.Expect(recover()).Should(BeNil())
	}()
	bytes := execute(serversTemplate, []http.Server{})
	g.Expect(bytes).ToNot(BeEmpty())
}

func TestExecutePanics(t *testing.T) {
	defer func() {
		g := NewWithT(t)
		g.Expect(recover()).ToNot(BeNil())
	}()

	_ = execute(serversTemplate, "not-correct-data")
}
