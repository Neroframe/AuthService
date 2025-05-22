package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	gomailpkg "github.com/Neroframe/AuthService/pkg/gomail"

	"github.com/Neroframe/AuthService/config"
	"github.com/Neroframe/AuthService/internal/adapters/bcrypt"
	"github.com/Neroframe/AuthService/internal/adapters/gomail"
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

	emailSender *gomailpkg.Sender
}

func New(ctx context.Context, cfg *config.Config, log *logger.Logger) (*App, error) {
	if cfg.Server.Addr == "" {
		return nil, errors.New("Server address empty")
	}

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

	// Ini email sender
	gomailSender := gomailpkg.New(gomailpkg.Config(cfg.Gomail))
	emailSender := gomail.NewGomailService(gomailSender)

	// Usecase
	userUC := usecase.NewUserUsecase(repo, hasher, publisher, redisCache, log, jwtSvc, emailSender)

	// gRPC client and clientConn (remove)
	authClient, authConn, err := grpcadapter.NewAuthClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("grpc AuthClient init: %w", err)
	}

	// Init gRPC interceptor
	authInt := middleware.NewAuthInterceptor(
		// set public routes
		[]string{
			"/auth.AuthService/Login",
			"/auth.AuthService/Register",
			"/auth.AuthService/ValidateToken",
			"/auth.AuthService/RestPassword",
			// "/auth.AuthService/GetUserByID" - admin only route
		},
		authClient,
		jwtSvc,
		log,
	)

	// Create gRPC handler that implements server logic
	authHandler := grpcadapter.NewHandler(userUC, log)
	// Create and configure gRPC server
	srv, err := grpcpkg.New(
		grpcpkg.Config(cfg.Server),
		// Attach register AuthService func
		func(s *grpc.Server) {
			authpb.RegisterAuthServiceServer(s, authHandler)
		},
		// Attach unary interceptors for logging and authentication
		[]grpc.UnaryServerInterceptor{
			authInt.UnaryLoggingInterceptor(),
			authInt.UnaryAuthentificate(),
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
		cfg: cfg,
		log: log,

		mongo: mongoClient,
		nats:  natsClient,
		redis: redisClient,
		grpc:  srv,

		authClient: authClient,
		authConn:   authConn,

		emailSender: gomailSender,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	// Share one ctx (error group)
	g, ctx := errgroup.WithContext(ctx)

	// Start the gRPC server
	g.Go(func() error {
		a.log.Info("starting gRPC", "addr", a.cfg.Server.Addr)
		return a.grpc.Run(ctx)
	})

	// Start Mongo health check
	g.Go(func() error {
		return healthLoop(ctx, a.mongo.HealthCheck, a.cfg.Mongo.SocketTimeout)
	})

	// Start NATS health check
	g.Go(func() error {
		return healthLoop(ctx, a.nats.HealthCheck, 3*time.Second)
	})

	// Start Redis health check
	g.Go(func() error {
		return healthLoop(ctx, a.redis.HealthCheck, 3*time.Second)
	})

	// Wait for all goroutines to finish or return the first encountered error
	return g.Wait()
}

func (a *App) Shutdown(ctx context.Context) error {
	var shutdownErr error

	a.log.Info("Gracefully stoping gRPC server")
	a.grpc.Stop()

	a.log.Info("Closing gRPC client connection to AuthService")
	if err := a.authConn.Close(); err != nil {
		a.log.Error("Failed to close authConn", "err", err)
		shutdownErr = errors.Join(shutdownErr, err)
	}

	a.log.Info("Disconnecting from NATS server")
	a.nats.Disconnect()

	a.log.Info("Closing Redis conn")
	a.redis.Close()

	a.log.Info("Disconnecting from Mongo")
	if err := a.mongo.Disconnect(ctx); err != nil {
		a.log.Error("Failed to disconnect from Mongo", "err", err)
		shutdownErr = errors.Join(shutdownErr, err)
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
