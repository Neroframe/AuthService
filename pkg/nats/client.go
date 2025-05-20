package nats

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
)

type Config struct {
	Hosts         []string
	Name          string
	MaxReconnects int
	ReconnectWait time.Duration
}

type Client struct {
	conn *nats.Conn
}

func NewClient(cfg Config) (*Client, error) {
	opts := []nats.Option{
		nats.Name(cfg.Name),
		nats.MaxReconnects(cfg.MaxReconnects),
		nats.ReconnectWait(cfg.ReconnectWait),
	}

	url := strings.Join(cfg.Hosts, ",")
	nc, err := nats.Connect(url, opts...)
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}
	return &Client{conn: nc}, nil
}

// drain and close
func (c *Client) Disconnect() {
	if err := c.conn.Drain(); err != nil {
		log.Printf("nats drain error: %v", err)
	}
	c.conn.Close()
}

func (c *Client) Publish(subject string, data []byte) error {
	return c.conn.Publish(subject, data)
}

func (c *Client) PublishJSON(subject string, v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}
	return c.conn.Publish(subject, b)
}

func (c *Client) Subscribe(subject string, handler nats.MsgHandler) (*nats.Subscription, error) {
	sub, err := c.conn.Subscribe(subject, handler)
	if err != nil {
		return nil, fmt.Errorf("subscribe %s: %w", subject, err)
	}
	return sub, nil
}
