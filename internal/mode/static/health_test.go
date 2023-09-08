package static

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("health", func() {
	It("returns an error if not ready", func() {
		hc := healthChecker{}
		Expect(hc.readyCheck(nil)).ToNot(Succeed())

		hc.ready = true
		Expect(hc.readyCheck(nil)).To(Succeed())
	})
})
