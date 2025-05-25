package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type MongoDB struct {
	MongoDSN             string `env:"MONGO_DSN"`
	MongoDatabase        string `env:"MONGO_DATABASE"`
	MaxConnectionRetries uint   `env:"MONGO_MAX_RETRIES"`
}

type Handler struct {
	Host         string        `env:"HTTP_HOST" env-default:"0.0.0.0"`
	Port         string        `env:"HTTP_PORT" env-default:"8080"`
	ReadTimeout  time.Duration `env:"HTTP_READ_TIMEOUT" env-default:"10s"`
	WriteTimeout time.Duration `env:"HTTP_WRITE_TIMEOUT" env-default:"10s"`
	IdleTimeout  time.Duration `env:"HTTP_IDLE_TIMEOUT" env-default:"10s"`
	JWTSecret    string        `env:"JWT_SECRET"`
	CORSOrigins  string        `env:"CORS_ORIGINS"`
}

type Logger struct {
	Level string `env:"LOGGER_LEVEL" env-default:"INFO"`
}

type Config struct {
	MongoDB
	Logger
	Handler
	Env string `env:"ENV" env-default:"dev"`
}

const filepath = "./.env"

func New() (*Config, error) {
	var c Config

	err := cleanenv.ReadConfig(filepath, &c)
	if errors.Is(err, os.ErrNotExist) {
		err = cleanenv.ReadEnv(&c)
	}
	if err != nil {
		return nil, fmt.Errorf("unable to read config: %w", err)
	}

	return &c, nil
}
