package config

import (
	"time"

	"github.com/caarlos0/env/v10"
)

type (
	Config struct {
		Version string `env:"APP_VERSION" envDefault:"1.0.0"`
		Server  Server `envPrefix:"GRPC_"`
		Mongo   Mongo
		Nats    Nats
		Redis   Redis
		JWT     JWT
		Gomail  Gomail
		Log     Log
	}

	// ------------ Server (gRPC) ------------
	Server struct {
		Addr string `env:"ADDRESS,notEmpty"`

		CertFile string
		KeyFile  string

		KeepaliveEnforcement struct {
			MinTime             time.Duration `env:"MIN_TIME" envDefault:"5s"`
			PermitWithoutStream bool          `env:"PERMIT_WITHOUT_STREAM" envDefault:"true"`
		}

		KeepaliveParams struct {
			MaxConnectionAge      time.Duration `env:"MAX_CONNECTION_AGE" envDefault:"30s"`
			MaxConnectionAgeGrace time.Duration `env:"MAX_CONNECTION_AGE_GRACE" envDefault:"10s"`
			MaxRecvMsgSizeMiB     int           `env:"MAX_MESSAGE_SIZE_MIB" envDefault:"12"`
		}
	}

	// ------------ Mongo ------------
	Mongo struct {
		Database       string        `env:"MONGO_DATABASE,notEmpty"`
		URI            string        `env:"MONGO_URI,notEmpty"`
		Username       string        `env:"MONGO_USER"`
		Password       string        `env:"MONGO_PASS"`
		ConnectTimeout time.Duration `env:"MONGO_CONN_TIMEOUT" envDefault:"3s"`
		SocketTimeout  time.Duration `env:"MONGO_SOCKET_TIMEOUT" envDefault:"3s"`
		MaxPoolSize    uint64        `env:"MONGO_MAX_POOL" envDefault:"100"`
		MinPoolSize    uint64        `env:"MONGO_MIN_POOL"`
		ReplicaSet     string        `env:"MONGO_REPLICA_SET"`
	}

	// ------------ NATS ------------
	Nats struct {
		Hosts         []string      `env:"NATS_HOSTS,notEmpty" envSeparator:","`
		Name          string        `env:"NATS_NAME" envDefault:"AuthService-NATS-Client"`
		MaxReconnects int           `env:"NATS_MAX_RECONNECTS"`
		ReconnectWait time.Duration `env:"NATS_RECONNECT_WAIT"`
		NatsSubjects  NatsSubjects
	}

	NatsSubjects struct {
		UserRegistered string
		UserLoggedIn   string
	}

	// ------------ Redis ------------
	Redis struct {
		Addr         string        `env:"REDIS_ADDR" envDefault:"localhost:6379"`
		Password     string        `env:"REDIS_PASSWORD"`
		DB           int           `env:"REDIS_DB" envDefault:"0"`
		DialTimeout  time.Duration `env:"REDIS_DIAL_TIMEOUT" envDefault:"5s"`
		ReadTimeout  time.Duration `env:"REDIS_READ_TIMEOUT" envDefault:"3s"`
		WriteTimeout time.Duration `env:"REDIS_WRITE_TIMEOUT" envDefault:"3s"`
		TLSEnable    bool          `env:"REDIS_TLS_ENABLE" envDefault:"false"`
	}

	// ------------ JWT ------------
	JWT struct {
		Secret     string        `env:"JWT_SECRET"`     // HMAC signing key
		Expiration time.Duration `env:"JWT_EXPIRATION"` // token ttl
	}

	// ------------ Gomail ---------
	Gomail struct {
		From         string `env:"GOMAIL_FROM"`
		Host         string `env:"GOMAIL_HOST"`
		Port         int    `env:"GOMAIL_PORT"`
		SMTPUsername string `env:"GOMAIL_SMTP_USERNAME"`
		SMTPPassword string `env:"GOMAIL_SMTP_PASSWORD"`
	}

	// ------------ Log ------------
	Log struct {
		Level        string `env:"LOG_LEVEL" envDefault:"info"`  // "debug", "info", "warn", "error"
		Format       string `env:"LOG_FORMAT" envDefault:"text"` // "text" or "json"
		SourceFolder string `env:"LOG_SOURCE_FOLDER"`            // project folder name
	}
)

func New() (*Config, error) {
	var cfg Config
	err := env.Parse(&cfg)

	return &cfg, err
}
