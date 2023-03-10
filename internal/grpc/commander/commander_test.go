package commander_test

import (
	"context"

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

				fakeServer := &commanderfakes.FakeCommander_UploadServer{
					ContextStub: func() context.Context {
						return metadata.NewIncomingContext(context.Background(),
							metadata.New(map[string]string{"uuid": "uuid"}))
					},
				}

				cmdr := commander.NewCommander(zap.New(), fakeMgr)

				err := cmdr.Upload(fakeServer)
				Expect(err).To(BeNil())
				Expect(fakeMgr.GetAgentCallCount()).To(Equal(1))
				Expect(fakeAgent.ReceiveFromUploadServerCallCount()).To(Equal(1))
			})
		})
		Describe("error cases", func() {
			tests := []struct {
				getAgentStub      func(string) commander.Agent
				contextStub       func() context.Context
				expErrString      string
				name              string
				getAgentCallCount int
			}{
				{
					contextStub:       func() context.Context { return context.Background() },
					expErrString:      "metadata is not provided",
					getAgentCallCount: 0,
					name:              "server context has no metadata",
				},
				{
					contextStub: func() context.Context {
						return metadata.NewIncomingContext(
							context.Background(),
							metadata.Pairs("key", "value"),
						)
					},
					expErrString:      "uuid is not in metadata",
					getAgentCallCount: 0,
					name:              "server context has no uuid key in metadata",
				},
				{
					contextStub: func() context.Context {
						return metadata.NewIncomingContext(
							context.Background(),
							metadata.Pairs("uuid", "val1", "uuid", "val2"),
						)
					},
					expErrString:      "more than one value for uuid in metadata",
					getAgentCallCount: 0,
					name:              "server context has multiple values for uuid in metadata",
				},
				{
					getAgentStub: func(s string) commander.Agent {
						return nil
					},
					contextStub: func() context.Context {
						return metadata.NewIncomingContext(
							context.Background(),
							metadata.Pairs("uuid", "val1"),
						)
					},
					expErrString:      "agent with id: val1 not found",
					getAgentCallCount: 1,
					name:              "agent does not exist",
				},
				{
					getAgentStub: func(s string) commander.Agent {
						return &commanderfakes.FakeAgent{
							StateStub: func() commander.State {
								return commander.StateConnected
							},
						}
					},
					contextStub: func() context.Context {
						return metadata.NewIncomingContext(
							context.Background(),
							metadata.Pairs("uuid", "val1"),
						)
					},
					expErrString:      "agent with id: val1 is not registered",
					getAgentCallCount: 1,
					name:              "agent is not registered",
				},
			}

			for _, test := range tests {
				fakeMgr := &commanderfakes.FakeAgentManager{
					GetAgentStub: test.getAgentStub,
				}

				fakeServer := &commanderfakes.FakeCommander_UploadServer{
					ContextStub: test.contextStub,
				}

				cmdr := commander.NewCommander(zap.New(), fakeMgr)
				err := cmdr.Upload(fakeServer)
				Expect(err).To(MatchError(test.expErrString))
				Expect(fakeMgr.GetAgentCallCount()).To(Equal(test.getAgentCallCount))
			}
		})
	})
})
