package commander_test

import (
	"context"

	"github.com/nginx/agent/sdk/v2/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc/metadata"

	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/grpc/commander"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/grpc/commander/commanderfakes"
)

var _ = Describe("Commander", func() {
	Describe("CommandChannel", func() {
		It("adds and removes agents over its lifetime", func() {
			ctx, cancel := context.WithCancel(context.TODO())

			fakeServer := &commanderfakes.FakeCommander_CommandChannelServer{
				ContextStub: func() context.Context {
					return metadata.NewIncomingContext(ctx, metadata.New(map[string]string{"uuid": "uuid"}))
				},
				RecvStub: func() (*proto.Command, error) {
					<-ctx.Done()
					return nil, nil
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
						return context.TODO()
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
		When("server context metadata is missing UUID", func() {
			It("errors and does not get agent", func() {
				fakeMgr := new(commanderfakes.FakeAgentManager)

				fakeServer := &commanderfakes.FakeCommander_UploadServer{
					ContextStub: func() context.Context {
						return context.TODO()
					},
				}

				cmdr := commander.NewCommander(zap.New(), fakeMgr)

				err := cmdr.Upload(fakeServer)
				Expect(err).ToNot(BeNil())
				Expect(fakeMgr.GetAgentCallCount()).To(BeZero())
			})
		})
		When("agent does not exist for the server", func() {
			It("errors", func() {
				fakeMgr := new(commanderfakes.FakeAgentManager)

				fakeServer := &commanderfakes.FakeCommander_UploadServer{
					ContextStub: func() context.Context {
						return metadata.NewIncomingContext(
							context.TODO(),
							metadata.New(map[string]string{"uuid": "uuid"}),
						)
					},
				}

				cmdr := commander.NewCommander(zap.New(), fakeMgr)

				err := cmdr.Upload(fakeServer)
				Expect(err).ToNot(BeNil())
				Expect(fakeMgr.GetAgentCallCount()).To(Equal(1))
			})
		})
		When("agent for the server exists but is not registered", func() {
			It("errors", func() {
				fakeMgr := &commanderfakes.FakeAgentManager{
					GetAgentStub: func(s string) commander.Agent {
						return &commanderfakes.FakeAgent{
							StateStub: func() commander.State {
								return commander.StateConnected
							},
						}
					},
				}

				fakeServer := new(commanderfakes.FakeCommander_UploadServer)
				fakeServer.ContextStub = func() context.Context {
					return metadata.NewIncomingContext(context.TODO(), metadata.New(map[string]string{"uuid": "uuid"}))
				}

				cmdr := commander.NewCommander(zap.New(), fakeMgr)

				err := cmdr.Upload(fakeServer)
				Expect(err).ToNot(BeNil())
				Expect(fakeMgr.GetAgentCallCount()).To(Equal(1))
			})
		})
		When("agent for the server exists and is registered", func() {
			It("calls ReceiveFromUploadServer on agent", func() {
				fakeAgent := &commanderfakes.FakeAgent{
					StateStub: func() commander.State {
						return commander.StateRegistered
					},
				}

				fakeMgr := &commanderfakes.FakeAgentManager{
					GetAgentStub: func(s string) commander.Agent {
						return fakeAgent
					},
				}

				fakeServer := new(commanderfakes.FakeCommander_UploadServer)
				fakeServer.ContextStub = func() context.Context {
					return metadata.NewIncomingContext(context.TODO(), metadata.New(map[string]string{"uuid": "uuid"}))
				}

				cmdr := commander.NewCommander(zap.New(), fakeMgr)

				err := cmdr.Upload(fakeServer)
				Expect(err).To(BeNil())
				Expect(fakeMgr.GetAgentCallCount()).To(Equal(1))
				Expect(fakeAgent.ReceiveFromUploadServerCallCount()).To(Equal(1))
			})
		})
	})
})
