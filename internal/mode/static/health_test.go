package static

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestReadyCheck(t *testing.T) {
	g := NewWithT(t)
	hc := healthChecker{}
	g.Expect(hc.readyCheck(nil)).ToNot(Succeed())

	hc.ready = true
	g.Expect(hc.readyCheck(nil)).To(Succeed())
}
