package nats

import (
	"context"

	"github.com/Neroframe/AuthService/internal/domain"
	"github.com/Neroframe/AuthService/pkg/nats"
)

type AuthPublisher struct {
	conn *nats.Client
}

func NewAuthPublisher(c *nats.Client) *AuthPublisher {
	return &AuthPublisher{conn: c}
}

func (p *AuthPublisher) PublishUserRegistered(ctx context.Context, evt domain.UserRegisteredEvent) error {
	return p.conn.PublishJSON("user.registered", evt)
}
