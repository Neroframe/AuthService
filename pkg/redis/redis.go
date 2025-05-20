package redis

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	redisv9 "github.com/redis/go-redis/v9"
)

type Config struct {
	Addr         string
	Password     string
	DB           int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	TLSEnable    bool
}

type Client struct {
	Client *redisv9.Client
}

func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	// TODO: Telemetry

	opts := &redisv9.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	if cfg.TLSEnable {
		opts.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	client := redisv9.NewClient(opts)

	err := client.Ping(ctx).Err()
	if err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	return &Client{Client: client}, nil
}

func (c *Client) Close() error {
	err := c.Client.Close()
	if err != nil {
		return fmt.Errorf("close: %w", err)
	}
	return nil
}

func (c *Client) Ping(ctx context.Context) error {
	err := c.Client.Ping(ctx).Err()
	if err != nil {
		return fmt.Errorf("ping: %w", err)
	}
	return nil
}
