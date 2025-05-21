package middleware

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Neroframe/AuthService/internal/domain"
	"github.com/Neroframe/AuthService/pkg/logger"
	authpb "github.com/Neroframe/AuthService/proto"
	"github.com/google/uuid"
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
	log         *logger.Logger
}

func NewAuthInterceptor(skipMethods []string, client authpb.AuthServiceClient, jwt domain.JWTService, log *logger.Logger) *AuthInterceptor {
	// skipMethods is a slice of RPC method names to bypass (e.g. Login, Register)
	sm := make(map[string]struct{}, len(skipMethods))
	for _, m := range skipMethods {
		sm[m] = struct{}{}
	}

	return &AuthInterceptor{
		skipMethods: sm,
		authClient:  client,
		jwtSvc:      jwt,
		log:         log,
	}
}

// Validate the token and pass tokenPayload into ctx
func (i *AuthInterceptor) UnaryAuthentificate() grpc.UnaryServerInterceptor {
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

		// Inject token payload into ctx
		ctx = context.WithValue(ctx, UserCtxKey, claims)
		fmt.Println("AUTH SUCCESS")
		return handler(ctx, req)
	}
}

// Logging incoming req
func (i *AuthInterceptor) UnaryLoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		// Extract or generate request ID
		md, _ := metadata.FromIncomingContext(ctx)
		var reqID string
		if vals := md.Get("x-request-id"); len(vals) > 0 {
			reqID = vals[0]
		} else {
			// Append reqID to ctx
			reqID = uuid.New().String()
			ctx = metadata.AppendToOutgoingContext(ctx, "x-request-id", reqID)
		}

		// Start timer for req
		start := time.Now()

		// Log incoming req
		i.log.Info("incoming gRPC request", "method", info.FullMethod, "request_id", reqID)

		// Call handler
		resp, err = handler(ctx, req)

		// Log duration, and returning value or err
		duration := time.Since(start)
		if err != nil {
			i.log.Error("gRPC request failed",
				"method", info.FullMethod,
				"request_id", reqID,
				"duration", duration,
				"error", err,
			)
		} else {
			i.log.Error("gRPC request failed",
				"method", info.FullMethod,
				"request_id", reqID,
				"duration", duration,
			)

		}

		return resp, err
	}
}

// TODO: rate limiting
