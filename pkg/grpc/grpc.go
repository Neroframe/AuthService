package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

type Config struct {
	Addr string

	CertFile string
	KeyFile  string

	KeepaliveEnforcement struct {
		MinTime             time.Duration
		PermitWithoutStream bool
	}

	KeepaliveParams struct {
		MaxConnectionAge      time.Duration
		MaxConnectionAgeGrace time.Duration
		MaxRecvMsgSizeMiB     int
	}
}

type Server struct {
	server   *grpc.Server
	listener net.Listener
}

func New(cfg Config, registerSrv func(*grpc.Server), unaryInts []grpc.UnaryServerInterceptor) (*Server, error) {
	opts := []grpc.ServerOption{
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             cfg.KeepaliveEnforcement.MinTime,
			PermitWithoutStream: cfg.KeepaliveEnforcement.PermitWithoutStream,
		}),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionAge:      cfg.KeepaliveParams.MaxConnectionAge,
			MaxConnectionAgeGrace: cfg.KeepaliveParams.MaxConnectionAgeGrace,
		}),
		grpc.MaxRecvMsgSize(cfg.KeepaliveParams.MaxRecvMsgSizeMiB << 20),
	}

	// If TLS certs are provided, enable TLS
	if cfg.CertFile != "" && cfg.KeyFile != "" {
		creds, err := credentials.NewServerTLSFromFile(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS credentials: %w", err)
		}
		opts = append(opts, grpc.Creds(creds))
	}

	// Append interceptors in order
	if len(unaryInts) > 0 {
		opts = append(opts, grpc.ChainUnaryInterceptor(unaryInts...))
	}

	//  Build options
	srv := grpc.NewServer(opts...)

	// Register service handlers
	registerSrv(srv)

	// Expose server reflection for debug (grpccurl)
	reflection.Register(srv)

	// Open TCP listener
	lis, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", cfg.Addr, err)
	}

	return &Server{server: srv, listener: lis}, nil
}

func (s *Server) Run(ctx context.Context) error {
	serveErr := make(chan error, 1)

	// Start gRPC serve loop
	go func() {
		serveErr <- s.server.Serve(s.listener)
	}()

	select {
	case <-ctx.Done():
		// Stop after context canceled
		s.Stop()
		return nil
	case err := <-serveErr:
		// Return an error
		return fmt.Errorf("grpc serve: %w", err)
	}
}

func (s *Server) Stop() {
	s.server.GracefulStop()
}
