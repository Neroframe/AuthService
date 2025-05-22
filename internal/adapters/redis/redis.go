package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Neroframe/AuthService/internal/domain"
	redisv9 "github.com/redis/go-redis/v9"
)

var ErrCacheMiss = errors.New("cache miss")

type AuthCache struct {
	client *redisv9.Client
	ttl    time.Duration
}

var _ domain.CodeCache = (*AuthCache)(nil)

func NewCodeCache(client *redisv9.Client, ttl time.Duration) *AuthCache {
	return &AuthCache{client: client, ttl: ttl}
}

func (c *AuthCache) Set(ctx context.Context, code *domain.VerificationCode) error {
	data, err := json.Marshal(code)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	if err := c.client.Set(ctx, code.UserID, data, c.ttl).Err(); err != nil {
		return fmt.Errorf("redis Set: %w", err)
	}

	return nil
}

func (c *AuthCache) Get(ctx context.Context, userID string) (*domain.VerificationCode, error) {
	data, err := c.client.Get(ctx, userID).Bytes()
	if err != nil {
		if err == redisv9.Nil {
			return nil, ErrCacheMiss
		}
		return nil, fmt.Errorf("redis Get: %w", err)
	}

	var code domain.VerificationCode
	if err := json.Unmarshal(data, &code); err != nil {
		return nil, fmt.Errorf("failed to unmarshal code: %w", err)
	}

	return &code, nil
}

func (c *AuthCache) Delete(ctx context.Context, userID string) error {
	err := c.client.Del(ctx, userID).Err()
	if err != nil {
		return fmt.Errorf("redis Del: %w", err)
	}

	return nil
}
