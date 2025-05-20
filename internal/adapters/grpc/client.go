package grpc

import (
	"fmt"

	"github.com/Neroframe/AuthService/config"
	authpb "github.com/Neroframe/AuthService/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewAuthClient(cfg *config.Config) (
	authpb.AuthServiceClient,
	*grpc.ClientConn,
	error,
) {
	var opts []grpc.DialOption

	// no TLS
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	conn, err := grpc.NewClient(cfg.Server.Addr, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("dial auth service: %w", err)
	}

	client := authpb.NewAuthServiceClient(conn)
	
	return client, conn, nil
}
