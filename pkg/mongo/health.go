package mongo

import (
	"context"
	"time"
)

func (c *Client) HealthCheck(parentCtx context.Context, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(parentCtx, timeout)
	defer cancel()

	return c.Client.Ping(ctx, nil)
}