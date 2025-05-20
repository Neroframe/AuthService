package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/Neroframe/AuthService/internal/domain"
	authpb "github.com/Neroframe/AuthService/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const UserCtxKey string = "userClaims"

type AuthInterceptor struct {
	skipMethods map[string]struct{}
	authClient  authpb.AuthServiceClient
	jwtSvc      domain.JWTService
}

// skipMethods is a slice of RPC method names to bypass (e.g. Login, Register)
func NewAuthInterceptor(skipMethods []string, client authpb.AuthServiceClient, jwt domain.JWTService) *AuthInterceptor {
	sm := make(map[string]struct{}, len(skipMethods))
	for _, m := range skipMethods {
		sm[m] = struct{}{}
	}
	return &AuthInterceptor{
		skipMethods: sm,
		authClient:  client,
		jwtSvc:      jwt,
	}
}

func (i *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		// Skip public methods
		if _, ok := i.skipMethods[info.FullMethod]; ok {
			return handler(ctx, req)
		}

		// Extract token
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		authHeaders := md["authorization"]
		if len(authHeaders) == 0 {
			return nil, status.Error(codes.Unauthenticated, "authorization header not supplied")
		}

		token := strings.TrimSpace(strings.TrimPrefix(authHeaders[0], "Bearer "))

		// Validate token
		claims, err := i.jwtSvc.Validate(ctx, token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		// Inject the proto response for user info
		ctx = context.WithValue(ctx, UserCtxKey, claims)
		fmt.Println("AUTH SUCCESS")
		return handler(ctx, req)
	}
}

// Audit logging, rate limiting
