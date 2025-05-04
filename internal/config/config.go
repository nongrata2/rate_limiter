package config

import (
	"log"
	"ratelimiter/internal/models"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Address          string `yaml:"address" env:"ADDRESS" env-default:"localhost:8080"`
	LogLevel         string `yaml:"log_level" env:"LOG_LEVEL" env-default:"DEBUG"`
	DefaultLimit     models.Limit
	ClientRateLimits []models.ClientLimit
}

func MustLoad(configPath string) Config {
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config %q: %s", configPath, err)
	}
	return cfg
}
