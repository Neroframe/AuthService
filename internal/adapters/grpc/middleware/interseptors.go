package middleware

import (
	"context"
	"strings"

	"github.com/Neroframe/AuthService/internal/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const UserCtxKey string = "userClaims"

type AuthInterceptor struct {
	jwtService  domain.JWTService
	skipMethods map[string]struct{}
}

// skipMethods is a slice of RPC method names to bypass (e.g. Login, Register)
func NewAuthInterceptor(jwtSvc domain.JWTService, skipMethods []string) *AuthInterceptor {
	sm := make(map[string]struct{}, len(skipMethods))
	for _, m := range skipMethods {
		sm[m] = struct{}{}
	}
	return &AuthInterceptor{
		jwtService:  jwtSvc,
		skipMethods: sm,
	}
}

func (i *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if _, ok := i.skipMethods[info.FullMethod]; ok {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		authHeaders := md["authorization"]
		if len(authHeaders) == 0 {
			return nil, status.Error(codes.Unauthenticated, "authorization header not supplied")
		}

		token := strings.TrimSpace(strings.TrimPrefix(authHeaders[0], "Bearer"))
		claims, err := i.jwtService.Validate(ctx, token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		// Inject claims into context for gprc handlers
		ctx = context.WithValue(ctx, UserCtxKey, &domain.TokenPayload{
			UserID:    claims.UserID,
			Email:     claims.Email,
			Role:      claims.Role,
			ExpiresAt: claims.ExpiresAt,
		})

		return handler(ctx, req)
	}
}
