package static

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestReadyCheck(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	nginxChecker := newNginxConfiguredOnStartChecker()
	g.Expect(nginxChecker.readyCheck(nil)).ToNot(Succeed())

	nginxChecker.ready = true
	g.Expect(nginxChecker.readyCheck(nil)).To(Succeed())
}
