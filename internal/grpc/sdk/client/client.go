package client

import (
	"context"
	"fmt"
	"log"
	"os"

	"google.golang.org/grpc"

	"github.com/nginxinc/nginx-gateway-fabric/internal/agent/file"
	pcontrolplane "github.com/nginxinc/nginx-gateway-fabric/internal/grpc/controlplane"
	"github.com/nginxinc/nginx-gateway-fabric/internal/grpc/sdk/server"
)

type Client struct {
	configCh      chan server.Config
	configApplyCh chan ApplyResult
	grpcClient    pcontrolplane.ControlPlaneClient
}

type ApplyResult struct {
	Generation uint32
	Success    bool
	Error      error
}

func NewClient(grpcClient pcontrolplane.ControlPlaneClient, configApplyCh chan ApplyResult) *Client {
	return &Client{
		configCh:      make(chan server.Config),
		configApplyCh: configApplyCh,
		grpcClient:    grpcClient,
	}
}

func (c *Client) ConfigCh() <-chan server.Config {
	return c.configCh
}

func (c *Client) Start(ctx context.Context) error {
	defer close(c.configCh)

	const hundredMb = 100 * 1024 * 1024

	stream, err := c.grpcClient.StreamMessages(ctx, grpc.MaxCallRecvMsgSize(2*hundredMb))
	if err != nil {
		return err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	request := &pcontrolplane.DataPlaneMessage{
		Message: &pcontrolplane.DataPlaneMessage_ConnectionRequest{
			ConnectionRequest: &pcontrolplane.ConnectionRequest{
				Host: hostname,
			},
		},
	}

	if err := stream.Send(request); err != nil {
		return err
	}

	for {
		response, err := stream.Recv()
		if err != nil {
			return err
		}

		cfg := response.GetConfiguration()
		if cfg == nil {
			return fmt.Errorf("received message is not a configuration")
		}

		log.Printf("Received message: %v", response)

		files := make([]file.File, 0, len(cfg.Files))
		for _, f := range cfg.Files {
			files = append(files, file.File{
				Path:    f.Path,
				Content: f.Content,
				Type:    file.Type(f.Type),
			})
		}

		select {
		case <-ctx.Done():
			return nil
		case c.configCh <- server.Config{
			Generation: cfg.Generation,
			Files:      files,
		}:
		}

		var applyResult ApplyResult

		select {
		case <-ctx.Done():
			return nil
		case applyResult = <-c.configApplyCh:
		}

		var errMsg string
		if applyResult.Error != nil {
			errMsg = applyResult.Error.Error()
		}

		cfgApplyResult := &pcontrolplane.DataPlaneMessage{
			Message: &pcontrolplane.DataPlaneMessage_ConfigurationApplyResult{
				ConfigurationApplyResult: &pcontrolplane.ConfigurationApplyResult{
					Generation: applyResult.Generation,
					Success:    applyResult.Success,
					Error:      errMsg,
				},
			},
		}

		if err := stream.Send(cfgApplyResult); err != nil {
			return err
		}
	}
}
