package grpc_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/nginx/agent/sdk/v2/client"
	agentGRPC "github.com/nginx/agent/sdk/v2/grpc"
	"github.com/nginx/agent/sdk/v2/proto"
	. "github.com/onsi/gomega"
	goGrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/grpc"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/grpc/commander"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/grpc/commander/commanderfakes"
)

func createTestClient(serverAddr string, clientUUID string) client.Commander {
	c := client.NewCommanderClient()
	c.WithServer(serverAddr)

	meta := agentGRPC.NewMeta(clientUUID, "", "")
	dialOptions := agentGRPC.DataplaneConnectionDialOptions("token", meta)
	dialOptions = append(dialOptions, goGrpc.WithTransportCredentials(insecure.NewCredentials()))
	c.WithDialOptions(dialOptions...)

	return c
}

func TestServer_ConcurrentConnections(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeMgr := new(commanderfakes.FakeConnectorManager)
	commanderService := commander.NewCommander(zap.New(), fakeMgr)

	server, err := grpc.NewServer(zap.New(), "localhost:0", commanderService)
	g.Expect(err).To(BeNil())
	g.Expect(server).ToNot(BeNil())

	ctx, cancel := context.WithCancel(context.TODO())

	errCh := make(chan error)
	go func() {
		errCh <- server.Start(ctx)
	}()

	verifySendAndRecv := func(c client.Commander, cmd *proto.Command, msgID string) {
		err := c.Send(ctx, client.MessageFromCommand(cmd))
		g.Expect(err).ToNot(HaveOccurred())

		select {
		case msg := <-c.Recv():
			g.Expect(msg.Meta().MessageId).To(Equal(msgID))
		case <-time.After(1 * time.Second):
			g.Fail("no message received from commander server")
		}
	}

	testClientServer := func(uuid string, wg *sync.WaitGroup) {
		defer wg.Done()

		c := createTestClient(server.Addr(), uuid)
		g.Expect(c.Connect(ctx)).To(Succeed())

		connectID := fmt.Sprintf("connect-%s", uuid)
		verifySendAndRecv(c, commander.CreateAgentConnectRequestCmd(connectID), connectID)
	}

	wg := &sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go testClientServer(fmt.Sprintf("client-%d", i), wg)
	}

	wg.Wait()

	cancel()

	err = <-errCh
	g.Expect(err).To(BeNil())
}
