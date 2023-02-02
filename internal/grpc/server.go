package grpc

import (
	"context"
	"net"

	"github.com/go-logr/logr"
	sdkGrpc "github.com/nginx/agent/sdk/v2/grpc"
	"github.com/nginx/agent/sdk/v2/proto"
	"google.golang.org/grpc"
)

const protocol = "tcp"

// Server is the gRPC server that handles requests from nginx agents.
type Server struct {
	listener net.Listener
	server   *grpc.Server
	logger   logr.Logger
}

// NewServer accepts a logger, address, and CommandServer implementation. It creates a gRPC server listening on the
// given address and registers the CommandServer implementation with the gRPC server.
func NewServer(logger logr.Logger, address string, commander proto.CommanderServer) (*Server, error) {
	listener, err := net.Listen(protocol, address)
	if err != nil {
		return nil, err
	}

	grpcServer := grpc.NewServer(sdkGrpc.DefaultServerDialOptions...)

	proto.RegisterCommanderServer(grpcServer, commander)

	s := &Server{
		logger:   logger,
		listener: listener,
		server:   grpcServer,
	}

	return s, nil
}

// Addr returns the address that the server is listening on.
func (s *Server) Addr() string {
	return s.listener.Addr().String()
}

// Start starts the gRPC server. If the context is canceled, the server is stopped.
func (s *Server) Start(ctx context.Context) error {
	go func() {
		<-ctx.Done()

		s.server.Stop()
		s.logger.Info("gRPC server stopped")
	}()

	s.logger.Info("Starting gRPC Server", "addr", s.listener.Addr().String())
	return s.server.Serve(s.listener)
}
