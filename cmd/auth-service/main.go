package main

import (
	"context"
	stdlog "log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Neroframe/AuthService/config"
	"github.com/Neroframe/AuthService/internal/app"
	"github.com/Neroframe/AuthService/pkg/logger"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		println("No .env file found, falling back to real env")
	}

	cfg, err := config.New()
	if err != nil {
		stdlog.Fatalf("config load error: %v", err)
	}

	// logger
	log := logger.New(logger.Config(cfg.Log))
	log.Info("config loaded", "version", cfg.Version)

	// Build the app
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app, err := app.New(ctx, cfg, log)
	if err != nil {
		log.Fatal("app init failed", "err", err)
	}

	// Run the app with error channel
	appErr := make(chan error)

	go func() {
		appErr <- app.Run(ctx)
	}()

	select {
	case <-ctx.Done():
		log.Info("shutdown signal received")
	case err := <-appErr:
		if err != nil {
			log.Error("app run error", "err", err)
		}
	}

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.Shutdown(shutdownCtx); err != nil {
		log.Error("error during shutdown", "err", err)
	}

	log.Info("Authentification service stopped")
}
