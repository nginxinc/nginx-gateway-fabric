package agent_test

import (
	"fmt"
	"sync"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/agent"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/grpc/commander"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/grpc/commander/commanderfakes"
)

func newFakeAgent(id string) commander.Agent {
	return &commanderfakes.FakeAgent{
		IDStub: func() string {
			return id
		},
	}
}

var _ = Describe("Agent Pool", func() {
	var (
		pool                   *agent.Pool
		agent1, agent2, agent3 commander.Agent
	)

	BeforeEach(func() {
		pool = agent.NewPool(zap.New())
		agent1, agent2, agent3 = newFakeAgent("1"), newFakeAgent("2"), newFakeAgent("3")
	})

	It("can add and get agents", func() {
		pool.AddAgent(agent1)
		Expect(pool.Size()).To(Equal(1))
		Expect(pool.GetAgent("1")).To(Equal(agent1))

		pool.AddAgent(agent2)
		Expect(pool.Size()).To(Equal(2))
		Expect(pool.GetAgent("2")).To(Equal(agent2))

		pool.AddAgent(agent3)
		Expect(pool.Size()).To(Equal(3))
		Expect(pool.GetAgent("3")).To(Equal(agent3))
	})
	It("can remove agents", func() {
		pool.AddAgent(agent1)
		pool.AddAgent(agent2)
		pool.AddAgent(agent3)
		Expect(pool.Size()).To(Equal(3))

		pool.RemoveAgent("2")
		Expect(pool.Size()).To(Equal(2))
		Expect(pool.GetAgent("1")).To(Equal(agent1))
		Expect(pool.GetAgent("3")).To(Equal(agent3))

		pool.RemoveAgent("1")
		Expect(pool.Size()).To(Equal(1))
		Expect(pool.GetAgent("3")).To(Equal(agent3))

		pool.RemoveAgent("3")
		Expect(pool.Size()).To(Equal(0))
	})
	When("an agent does not exist in pool", func() {
		It("remove agent does nothing", func() {
			Expect(pool.Size()).To(Equal(0))
			pool.RemoveAgent("dne")
			Expect(pool.Size()).To(Equal(0))
		})
	})
	It("can handle concurrent CRUD", func() {
		// populate pool with 5 agents which will be removed.
		for i := 1; i <= 5; i++ {
			pool.AddAgent(newFakeAgent(fmt.Sprintf("%d", i)))
		}

		addAndGetAgent := func(id string, wg *sync.WaitGroup) {
			defer wg.Done()
			pool.AddAgent(newFakeAgent(id))
			Expect(pool.GetAgent(id).ID()).To(Equal(id))
		}

		removeAndGetAgent := func(id string, wg *sync.WaitGroup) {
			defer wg.Done()
			Expect(pool.GetAgent(id).ID()).To(Equal(id))
			pool.RemoveAgent(id)
			Expect(pool.GetAgent(id)).To(BeNil())
		}

		wg := &sync.WaitGroup{}
		for i := 0; i < 15; i++ {
			id := fmt.Sprintf("%d", i+1)

			wg.Add(1)
			// remove first five
			if i < 5 {
				go removeAndGetAgent(id, wg)
			} else {
				// add 10 new
				go addAndGetAgent(id, wg)
			}
		}

		wg.Wait()

		Expect(pool.Size()).To(Equal(10))
	})
})
