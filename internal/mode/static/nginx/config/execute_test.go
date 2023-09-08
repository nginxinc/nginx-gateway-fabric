package config

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/mode/static/nginx/config/http"
)

func TestExecute(t *testing.T) {
	defer func() {
		g := NewGomegaWithT(t)
		g.Expect(recover()).Should(BeNil())
	}()
	g := NewGomegaWithT(t)
	bytes := execute(serversTemplate, []http.Server{})
	g.Expect(len(bytes)).ToNot(Equal(0))
}

func TestExecutePanics(t *testing.T) {
	defer func() {
		g := NewGomegaWithT(t)
		g.Expect(recover()).ShouldNot(BeNil())
	}()

	_ = execute(serversTemplate, "not-correct-data")
}
