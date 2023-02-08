package agent_test

import (
	"fmt"
	"sync"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/grpc/agent"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/grpc/commander/commanderfakes"
)

func newFakeConn(id string) *commanderfakes.FakeConnector {
	conn := new(commanderfakes.FakeConnector)
	conn.IDStub = func() string { return id }
	return conn
}

var _ = Describe("Agent Pool", func() {
	var (
		pool                *agent.Pool
		conn1, conn2, conn3 *commanderfakes.FakeConnector
	)

	BeforeEach(func() {
		pool = agent.NewPool(zap.New())
		conn1, conn2, conn3 = newFakeConn("1"), newFakeConn("2"), newFakeConn("3")
	})

	It("can add and get connectors", func() {
		pool.AddConnector(conn1)
		Expect(pool.Size()).To(Equal(1))
		Expect(pool.GetConnector("1")).To(Equal(conn1))

		pool.AddConnector(conn2)
		Expect(pool.Size()).To(Equal(2))
		Expect(pool.GetConnector("2")).To(Equal(conn2))

		pool.AddConnector(conn3)
		Expect(pool.Size()).To(Equal(3))
		Expect(pool.GetConnector("3")).To(Equal(conn3))
	})
	It("can remove connectors", func() {
		pool.AddConnector(conn1)
		pool.AddConnector(conn2)
		pool.AddConnector(conn3)
		Expect(pool.Size()).To(Equal(3))

		pool.RemoveConnector("2")
		Expect(pool.Size()).To(Equal(2))
		Expect(pool.GetConnector("1")).To(Equal(conn1))
		Expect(pool.GetConnector("3")).To(Equal(conn3))

		pool.RemoveConnector("1")
		Expect(pool.Size()).To(Equal(1))
		Expect(pool.GetConnector("3")).To(Equal(conn3))

		pool.RemoveConnector("3")
		Expect(pool.Size()).To(Equal(0))
	})
	When("a connector does not exist in pool", func() {
		It("remove connector does nothing", func() {
			Expect(pool.Size()).To(Equal(0))
			pool.RemoveConnector("dne")
			Expect(pool.Size()).To(Equal(0))
		})
	})
	It("can handle concurrent CRUD", func() {
		// populate pool with 5 connectors which will be removed.
		pool.AddConnector(conn1)
		pool.AddConnector(conn2)
		pool.AddConnector(conn3)
		pool.AddConnector(newFakeConn("4"))
		pool.AddConnector(newFakeConn("5"))

		addGetConn := func(id string, wg *sync.WaitGroup) {
			defer wg.Done()
			pool.AddConnector(newFakeConn(id))
			Expect(pool.GetConnector(id).ID()).To(Equal(id))
		}

		removeGetConn := func(id string, wg *sync.WaitGroup) {
			defer wg.Done()
			Expect(pool.GetConnector(id).ID()).To(Equal(id))
			pool.RemoveConnector(id)
			Expect(pool.GetConnector(id)).To(BeNil())
		}

		wg := &sync.WaitGroup{}
		for i := 0; i < 15; i++ {
			id := fmt.Sprintf("%d", i+1)

			wg.Add(1)
			// remove first five
			if i < 5 {
				go removeGetConn(id, wg)
			} else {
				// add 10 new
				go addGetConn(id, wg)
			}
		}

		wg.Wait()

		Expect(pool.Size()).To(Equal(10))
	})
})
