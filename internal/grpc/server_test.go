package grpc_test

import (
	"context"
	"testing"

	"github.com/nginx/agent/sdk/v2/client"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	goGrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/grpc"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/grpc/service"
)

// This test is pretty simple at the moment. We are only verifying that the server can be started, stopped,
// and that the Commander implementation is registered with the server.
// Once we add more functionality this test may become more meaningful.
func TestServer(t *testing.T) {
	g := NewGomegaWithT(t)

	buf := gbytes.NewBuffer()
	logger := zap.New(zap.WriteTo(buf))

	server, err := grpc.NewServer(logger, "localhost:0", service.NewCommander(logger))
	g.Expect(err).To(BeNil())
	g.Expect(server).ToNot(BeNil())

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		g.Expect(server.Start(ctx)).To(Succeed())
	}()

	commanderClient := client.NewCommanderClient()
	commanderClient.WithServer(server.Addr())
	commanderClient.WithDialOptions(goGrpc.WithTransportCredentials(insecure.NewCredentials()))

	err = commanderClient.Connect(ctx)
	g.Expect(err).To(BeNil())

	g.Eventually(buf).Should(gbytes.Say("Commander CommandChannel"))

	cancel()
	g.Eventually(buf).Should(gbytes.Say("gRPC server stopped"))
}
