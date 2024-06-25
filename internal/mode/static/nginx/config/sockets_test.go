package config

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestGetSocketName(t *testing.T) {
	res := getSocketNameTLS(800, "*.cafe.example.com")

	g := NewGomegaWithT(t)
	g.Expect(res).To(Equal("unix:/var/run/nginx/*.cafe.example.com800.sock"))
}
