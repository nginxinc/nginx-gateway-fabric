package main

import (
	"context"
	"log"

	"google.golang.org/grpc"

	pcontrolplane "github.com/nginxinc/nginx-gateway-fabric/internal/grpc/controlplane"
	"github.com/nginxinc/nginx-gateway-fabric/internal/grpc/sdk/client"
)

func main() {
	conn, err := grpc.Dial("localhost:9001", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	grpcClient := pcontrolplane.NewControlPlaneClient(conn)

	applyResultCh := make(chan client.ApplyResult)
	defer close(applyResultCh)

	c := client.NewClient(grpcClient, applyResultCh)

	go func() {
		if err := c.Start(context.Background()); err != nil {
			log.Fatalf("failed to start c: %v", err)
		}
	}()

	for {
		select {
		case cfg := <-c.ConfigCh():
			log.Printf("Received config: %v", cfg)

			applyResultCh <- client.ApplyResult{
				Generation: cfg.Generation,
				Success:    true,
			}
		}
	}
}
