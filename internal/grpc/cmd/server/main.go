package main

import (
	"context"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"

	"github.com/nginxinc/nginx-gateway-fabric/internal/agent/file"
	"github.com/nginxinc/nginx-gateway-fabric/internal/grpc/controlplane"
	"github.com/nginxinc/nginx-gateway-fabric/internal/grpc/sdk/server"
)

func main() {
	log.Println("Starting server...")

	l, err := net.Listen("tcp", ":9001")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	cp := server.NewControlPlane()
	go func() {
		if err := cp.Start(context.Background()); err != nil {
			log.Fatalf("failed to start control plane: %v", err)
		}
	}()
	s := grpc.NewServer()
	controlplane.RegisterControlPlaneServer(s, cp)

	const hundredMb = 100 * 1024 * 1024
	largeBytes := make([]byte, hundredMb)

	go func() {
		for i := 1; i < 100; i++ {
			log.Printf("Updating config %d", i)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			err := cp.UpdateConfig(ctx, server.Config{
				Generation: uint32(i),
				Files: []file.File{
					{
						Path:    "/etc/nginx/conf.d/default.conf",
						Type:    file.TypeRegular,
						Content: []byte(`server { ... }`),
					},
					{
						Path: "/etc/nginx/conf.d/large.conf",
						Type: file.TypeRegular,
						// 100 MB
						Content: largeBytes,
					},
					{
						Path:    "/etc/nginx/secrets/something.pem",
						Type:    file.TypeSecret,
						Content: []byte(`-----BEGIN CERTIFICATE----- ...`),
					},
				},
			})
			if err != nil {
				log.Fatalf("failed to update config: %v", err)
			}
			cancel()
			time.Sleep(5 * time.Second)
		}
	}()

	if err := s.Serve(l); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
