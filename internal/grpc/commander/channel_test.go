package commander_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/nginx/agent/sdk/v2/proto"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/grpc/commander"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/grpc/commander/commanderfakes"
)

func newTestCommand(msgID string) *proto.Command {
	return &proto.Command{Meta: &proto.Metadata{MessageId: msgID}}
}

func TestBidirectionalChannel(t *testing.T) {
	g := NewGomegaWithT(t)

	recvCommands := make(chan *proto.Command)
	sentCommands := make(chan *proto.Command)

	ctx, cancel := context.WithCancel(context.TODO())

	fakeServer := &commanderfakes.FakeCommander_CommandChannelServer{
		RecvStub: func() (*proto.Command, error) {
			select {
			case <-ctx.Done():
				return nil, nil
			case cmd := <-recvCommands:
				return cmd, nil
			}
		},
		SendStub: func(command *proto.Command) error {
			sentCommands <- command
			return nil
		},
	}

	ch := commander.NewBidirectionalChannel(fakeServer, zap.New())

	errCh := make(chan error)

	go func() {
		errCh <- ch.Run(ctx)
	}()

	testRecvCommand := func(msgID string) {
		recvCommands <- newTestCommand(msgID)
		cmd := <-ch.Out()
		g.Expect(cmd.GetMeta().GetMessageId()).To(Equal(msgID))
	}

	testSendCommand := func(msgID string) {
		ch.In() <- newTestCommand(msgID)
		cmd := <-sentCommands
		g.Expect(cmd.GetMeta().GetMessageId()).To(Equal(msgID))
	}

	for i := 0; i < 5; i++ {
		msgID := fmt.Sprintf("msg-%d", i)
		testRecvCommand(msgID)
		testSendCommand(msgID)
	}

	cancel()

	err := <-errCh
	g.Expect(err).Should(MatchError(context.Canceled))
}

func TestBidirectionalChannel_SendError(t *testing.T) {
	g := NewGomegaWithT(t)

	done := make(chan struct{})

	fakeServer := &commanderfakes.FakeCommander_CommandChannelServer{
		RecvStub: func() (*proto.Command, error) {
			return newTestCommand("msg-id"), nil
		},
		SendStub: func(command *proto.Command) error {
			<-done
			return errors.New("send error")
		},
	}

	ch := commander.NewBidirectionalChannel(fakeServer, zap.New())

	errCh := make(chan error)

	go func() {
		errCh <- ch.Run(context.Background())
	}()

	ch.In() <- newTestCommand("msg-id")
	close(done)

	err := <-errCh
	g.Expect(err).Should(MatchError("error sending command to CommandChannel: send error"))
}

func TestBidirectionalChannel_RecvError(t *testing.T) {
	g := NewGomegaWithT(t)

	done := make(chan struct{})

	fakeServer := &commanderfakes.FakeCommander_CommandChannelServer{
		RecvStub: func() (*proto.Command, error) {
			<-done
			return nil, errors.New("recv error")
		},
	}

	ch := commander.NewBidirectionalChannel(fakeServer, zap.New())

	errCh := make(chan error)

	go func() {
		errCh <- ch.Run(context.Background())
	}()

	close(done)

	err := <-errCh
	g.Expect(err).Should(MatchError("error receiving command from CommandChannel: recv error"))
}

func TestBidirectionalChannel_NilCommand(t *testing.T) {
	g := NewGomegaWithT(t)

	recvCommands := make(chan *proto.Command)
	ctx, cancel := context.WithCancel(context.Background())

	fakeServer := &commanderfakes.FakeCommander_CommandChannelServer{
		RecvStub: func() (*proto.Command, error) {
			select {
			case <-ctx.Done():
				return nil, nil
			case cmd := <-recvCommands:
				return cmd, nil
			}
		},
	}

	ch := commander.NewBidirectionalChannel(fakeServer, zap.New())

	errCh := make(chan error)

	go func() {
		errCh <- ch.Run(ctx)
	}()

	testRecvCommand := func(msgID string) {
		recvCommands <- newTestCommand(msgID)
		cmd := <-ch.Out()
		g.Expect(cmd.GetMeta().GetMessageId()).To(Equal(msgID))
	}

	testRecvCommand("msg-1")
	// add a nil command to the recv channel
	recvCommands <- nil
	// test that channel is still running and can receive non-nil commands
	testRecvCommand("msg-2")

	cancel()

	err := <-errCh
	g.Expect(err).Should(MatchError(context.Canceled))
}
