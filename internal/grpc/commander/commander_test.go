package commander_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	status2 "google.golang.org/grpc/status"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/grpc/commander"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/grpc/commander/commanderfakes"
)

var _ = Describe("Commander", func() {
	Describe("CommandChannel", func() {
		It("adds and removes agents over its lifetime", func() {
			ctx, cancel := context.WithCancel(context.Background())

			fakeServer := &commanderfakes.FakeCommander_CommandChannelServer{
				ContextStub: func() context.Context {
					return metadata.NewIncomingContext(ctx, metadata.New(map[string]string{"uuid": "uuid"}))
				},
			}

			added := make(chan struct{})
			fakeMgr := &commanderfakes.FakeAgentManager{
				AddAgentStub: func(_ commander.Agent) {
					close(added)
				},
			}

			cmdr := commander.NewCommander(zap.New(), fakeMgr)

			errCh := make(chan error)
			go func() {
				errCh <- cmdr.CommandChannel(fakeServer)
			}()

			<-added
			Expect(fakeMgr.AddAgentCallCount()).To(Equal(1))

			cancel()

			err := <-errCh
			Expect(err).Should(MatchError(context.Canceled))
			Expect(fakeMgr.RemoveAgentCallCount()).To(Equal(1))
		})
		When("server context metadata is missing UUID", func() {
			It("errors and does not add agent to manager", func() {
				fakeMgr := new(commanderfakes.FakeAgentManager)

				fakeServer := &commanderfakes.FakeCommander_CommandChannelServer{
					ContextStub: func() context.Context {
						return context.Background()
					},
				}

				cmdr := commander.NewCommander(zap.New(), fakeMgr)
				err := cmdr.CommandChannel(fakeServer)
				Expect(err).ToNot(BeNil())

				Expect(fakeMgr.AddAgentCallCount()).To(Equal(0))
			})
		})
	})
	Describe("Upload", func() {
		It("returns Unimplemented error code", func() {
			cmdr := commander.NewCommander(zap.New(), &commanderfakes.FakeAgentManager{})
			err := cmdr.Upload(&commanderfakes.FakeCommander_UploadServer{})
			status, ok := status2.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(status.Code()).To(Equal(codes.Unimplemented))
		})
	})
})
