package config

import (
	"time"

	"github.com/caarlos0/env/v10"
)

type (
	Config struct {
		Version string
		Server  Server
		Mongo   Mongo
		Nats    Nats
		Redis   Redis
		JWT     JWT
		Log     Log
	}

	Server struct {
		gRPCServer gRPCServer
		// HTTPServer?
	}
	gRPCServer struct {
	}

	Nats struct {
	}

	Redis struct {
		Addr         string        `env:"REDIS_ADDR" envDefault:"localhost:6379"`
		Password     string        `env:"REDIS_PASSWORD"`
		DB           int           `env:"REDIS_DB" envDefault:"0"`
		DialTimeout  time.Duration `env:"REDIS_DIAL_TIMEOUT" envDefault:"5s"`
		ReadTimeout  time.Duration `env:"REDIS_READ_TIMEOUT" envDefault:"3s"`
		WriteTimeout time.Duration `env:"REDIS_WRITE_TIMEOUT" envDefault:"3s"`
		TLSEnable    bool          `env:"REDIS_TLS_ENABLE" envDefault:"false"`
	}

	JWT struct {
		Secret     string        `env:"JWT_SECRET"`     // HMAC signing key
		Expiration time.Duration `env:"JWT_EXPIRATION"` // token ttl
	}

	Log struct {
		Level  string `env:"LOG_LEVEL" envDefault:"info"`  // "debug", "info", "warn", "error"
		Format string `env:"LOG_FORMAT" envDefault:"text"` // "text" or "json"
	}
)

func NewCfg() (*Config, error) {
	var cfg Config
	err := env.Parse(&cfg)

	return &cfg, err
}
