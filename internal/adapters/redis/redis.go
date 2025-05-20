package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Neroframe/AuthService/internal/domain"
	redisv9 "github.com/redis/go-redis/v9"
)

var ErrCacheMiss = fmt.Errorf("cache miss")

type UserCache struct {
	client *redisv9.Client
	ttl    time.Duration
}

var _ domain.UserCache = (*UserCache)(nil)

func NewUserCache(client *redisv9.Client, ttl time.Duration) *UserCache {
	return &UserCache{client: client, ttl: ttl}
}

func (c *UserCache) GetByID(ctx context.Context, id string) (*domain.User, error) {
	data, err := c.client.Get(ctx, id).Bytes()
	if err == redisv9.Nil {
		return nil, ErrCacheMiss
	} else if err != nil {
		return nil, err
	}
	var u domain.User
	if err := json.Unmarshal(data, &u); err != nil {
		return nil, err
	}
	return &u, nil
}
