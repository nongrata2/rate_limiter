package config

import (
	"log"
	"ratelimiter/internal/models"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Address          string               `yaml:"address" env:"ADDRESS" env-default:":8080"`
	LogLevel         string               `yaml:"log_level" env:"LOG_LEVEL" env-default:"DEBUG"`
	DefaultLimit     models.Limit         `yaml:"default_limit" env:"DEFAULT_LIMIT"`
	ClientRateLimits []models.ClientLimit `yaml:"client_rate_limits" env:"CLIENT_RATE_LIMITS"`
	DBHost           string               `env:"DB_HOST" env-default:"db"`
	DBUser           string               `env:"DB_USER" env-default:"postgres"`
	DBPassword       string               `env:"DB_PASSWORD" env-default:"postgres"`
	DBName           string               `env:"DB_NAME" env-default:"postgres"`
	DBPort           string               `env:"DB_PORT" env-default:"5432"`
}

func MustLoad(configPath string) Config {
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config %q: %s", configPath, err)
	}
	return cfg
}
