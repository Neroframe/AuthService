package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Config struct {
	Database       string
	URI            string
	Username       string
	Password       string
	ConnectTimeout time.Duration
	SocketTimeout  time.Duration
	MaxPoolSize    uint64
	MinPoolSize    uint64
	ReplicaSet     string
}

type Client struct {
	Client *mongo.Client
	DB     *mongo.Database
}

func NewClient(parentCtx context.Context, cfg Config) (*Client, error) {
	dialCtx, dialCancel := context.WithTimeout(parentCtx, cfg.ConnectTimeout)
	defer dialCancel()

	// Build client options
	opts := options.Client()

	opts.ApplyURI(cfg.URI)

	if cfg.Username != "" {
		opts.SetAuth(options.Credential{
			Username: cfg.Username,
			Password: cfg.Password,
		})
	}

	// connection pool and socket settings
	opts = opts.
		SetConnectTimeout(cfg.ConnectTimeout).
		SetSocketTimeout(cfg.SocketTimeout).
		SetMaxPoolSize(cfg.MaxPoolSize).
		SetMinPoolSize(cfg.MinPoolSize)

	if cfg.ReplicaSet != "" {
		opts.SetReplicaSet(cfg.ReplicaSet)
	}

	// Connect to MongoDB
	client, err := mongo.Connect(dialCtx, opts)
	if err != nil {
		return nil, fmt.Errorf("mongo connect: %w", err)
	}

	// verify
	pingCtx, pingCancel := context.WithTimeout(parentCtx, cfg.SocketTimeout)
	defer pingCancel()
	if err := client.Ping(pingCtx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("mongo ping: %w", err)
	}

	db := client.Database(cfg.Database)

	return &Client{DB: db, Client: client}, nil
}

func (c *Client) Disconnect(ctx context.Context) error {
	return c.Client.Disconnect(ctx)
}
