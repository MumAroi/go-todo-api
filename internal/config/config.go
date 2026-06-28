package config

import (
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	DatabaseURL   string        `env:"DATABASE_URL"`
	AppPort       string        `env:"APP_PORT"`
	JWTSecret     string        `env:"JWT_SECRET"`
	JWTExpiration time.Duration `env:"JWT_EXPIRATION"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
