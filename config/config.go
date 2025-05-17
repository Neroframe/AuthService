package config

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

	}

	Nats struct {

	}

	Redis struct {
		
	}

	Log struct {
		Level  string // info warn error
		Format string // json or text
	}
)
