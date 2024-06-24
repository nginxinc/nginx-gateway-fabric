package config

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestGetSocketName(t *testing.T) {
	res := getSocketName(800, "*.cafe:example.com")

	g := NewGomegaWithT(t)
	g.Expect(res).To(Equal("unix:/var/run/nginx/:s.cafe::example.com800.sock"))
}
