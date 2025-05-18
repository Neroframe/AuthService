package redis

import (
	"time"

	"github.com/redis/go-redis"
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
	Client *redis.Client
}

func New()
