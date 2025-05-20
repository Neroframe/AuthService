package app

import (
	"context"
	"fmt"
	"time"

	"github.com/Neroframe/AuthService/config"
	"github.com/Neroframe/AuthService/internal/adapters/bcrypt"
	grpcadapter "github.com/Neroframe/AuthService/internal/adapters/grpc"
	"github.com/Neroframe/AuthService/internal/adapters/grpc/middleware"
	mongoadapter "github.com/Neroframe/AuthService/internal/adapters/mongo"
	natsadapter "github.com/Neroframe/AuthService/internal/adapters/nats"
	redisadapter "github.com/Neroframe/AuthService/internal/adapters/redis"
	"github.com/Neroframe/AuthService/internal/adapters/token"
	"github.com/Neroframe/AuthService/internal/usecase"
	grpcpkg "github.com/Neroframe/AuthService/pkg/grpc"
	"github.com/Neroframe/AuthService/pkg/logger"
	mongopkg "github.com/Neroframe/AuthService/pkg/mongo"
	natspkg "github.com/Neroframe/AuthService/pkg/nats"
	redispkg "github.com/Neroframe/AuthService/pkg/redis"
	authpb "github.com/Neroframe/AuthService/proto"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type App struct {
	cfg *config.Config
	log *logger.Logger

	mongo *mongopkg.Client
	nats  *natspkg.Client
	redis *redispkg.Client

	grpc       *grpcpkg.Server
	authClient authpb.AuthServiceClient
	authConn   *grpc.ClientConn
}

func New(ctx context.Context, cfg *config.Config, log *logger.Logger) (*App, error) {
	log.Info("initializing infra clients")

	// MongoDB
	mongoClient, err := mongopkg.NewClient(ctx, mongopkg.Config(cfg.Mongo))
	if err != nil {
		return nil, fmt.Errorf("mongo connect: %w", err)
	}

	// NATS
	natsClient, err := natspkg.NewClient(natspkg.Config{
		Hosts:         cfg.Nats.Hosts,
		Name:          cfg.Nats.Name,
		MaxReconnects: cfg.Nats.MaxReconnects,
		ReconnectWait: cfg.Nats.ReconnectWait,
	})
	if err != nil {
		mongoClient.Disconnect(ctx)
		return nil, fmt.Errorf("nats connect: %w", err)
	}

	// Redis
	redisClient, err := redispkg.NewClient(ctx, redispkg.Config(cfg.Redis))
	if err != nil {
		mongoClient.Disconnect(ctx)
		natsClient.Disconnect()
		return nil, fmt.Errorf("redis connect: %w", err)
	}

	// Adapters
	repo, err := mongoadapter.NewUserRepository(ctx, mongoClient.DB)
	if err != nil {
		return nil, fmt.Errorf("mongo repo init: %w", err)
	}
	publisher := natsadapter.NewAuthPublisher(natsClient)
	redisCache := redisadapter.NewUserCache(redisClient.Client, cfg.Redis.DialTimeout)

	// jwt and bcrypt helper services
	jwtSvc := token.NewJWTService(cfg.JWT.Secret, cfg.JWT.Expiration)
	hasher := bcrypt.NewHasher()

	// Usecase
	userUC := usecase.NewUserUsecase(repo, hasher, publisher, redisCache, log, jwtSvc)

	// gRPC client and clientConn (remove)
	authClient, authConn, err := grpcadapter.NewAuthClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("grpc AuthClient init: %w", err)
	}

	// gRPC interceptor that uses client
	// it validates token and injects user info
	authInt := middleware.NewAuthInterceptor(
		// public routes
		[]string{
			"/auth.AuthService/Login",
			"/auth.AuthService/Register",
			"/auth.AuthService/ValidateToken",
			// "/auth.AuthService/GetUserByID" - admin only route
		},
		authClient,
		jwtSvc,
	)

	// gRPC server setup
	authHandler := grpcadapter.NewHandler(userUC, log) // handler implements server logic
	srv, err := grpcpkg.New(
		grpcpkg.Config(cfg.Server),
		func(s *grpc.Server) {
			authpb.RegisterAuthServiceServer(s, authHandler) // register AuthService in a gRPC server
		},
		[]grpc.UnaryServerInterceptor{
			authInt.Unary(),
		},
	)
	if err != nil {
		mongoClient.Disconnect(ctx)
		natsClient.Disconnect()
		redisClient.Close()
		authConn.Close()
		return nil, fmt.Errorf("grpc server init: %w", err)
	}

	return &App{
		cfg:        cfg,
		log:        log,
		mongo:      mongoClient,
		nats:       natsClient,
		redis:      redisClient,
		grpc:       srv,
		authClient: authClient,
		authConn:   authConn,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	// Share one ctx (error group)
	g, ctx := errgroup.WithContext(ctx)

	// gRPC
	g.Go(func() error {
		a.log.Info("starting gRPC", "addr", a.cfg.Server.Addr)
		return a.grpc.Run(ctx)
	})

	// Mongo health
	g.Go(func() error {
		return healthLoop(ctx, a.mongo.HealthCheck, a.cfg.Mongo.SocketTimeout)
	})

	// NATS health
	g.Go(func() error {
		return healthLoop(ctx, a.nats.HealthCheck, 3*time.Second)
	})

	// Redis health
	g.Go(func() error {
		return healthLoop(ctx, a.redis.HealthCheck, 3*time.Second)
	})

	// returns the first error
	return g.Wait()
}

func (a *App) Shutdown(ctx context.Context) error {
	var shutdownErr error

	a.log.Info("shutting down gRPC")
	a.grpc.Stop()

	a.log.Info("closing AuthService gRPC connection")
	if err := a.authConn.Close(); err != nil {
		a.log.Error("failed to close authConn", "err", err)
		shutdownErr = err
	}

	a.log.Info("disconnecting NATS")
	a.nats.Disconnect()

	a.log.Info("closing Redis")
	a.redis.Close()

	a.log.Info("disconnecting Mongo")
	if err := a.mongo.Disconnect(ctx); err != nil {
		a.log.Error("failed to disconnect Mongo", "err", err)
		shutdownErr = err
	}

	return shutdownErr
}

func healthLoop(ctx context.Context, hc func(context.Context, time.Duration) error, timeout time.Duration) error {
	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()

	var fails int
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := hc(ctx, timeout); err != nil {
				fails++
				if fails > 3 {
					return fmt.Errorf("unhealthy: %w", err)
				}
			} else {
				fails = 0
			}
		}
	}
}
