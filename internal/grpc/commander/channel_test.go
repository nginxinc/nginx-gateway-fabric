package commander_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/nginx/agent/sdk/v2/proto"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/grpc/commander"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/grpc/commander/commanderfakes"
)

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
		recvCommands <- &proto.Command{Meta: &proto.Metadata{MessageId: msgID}}
		cmd := <-ch.Out()
		g.Expect(cmd.GetMeta().GetMessageId()).To(Equal(msgID))
	}

	testSendCommand := func(msgID string) {
		ch.In() <- &proto.Command{Meta: &proto.Metadata{MessageId: msgID}}
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
