package configs

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Server struct {
		Port int `yaml:"port" env:"SERVER_PORT" env-default:"50051"`
	} `yaml:"server"`
	Limiter struct {
		// Type of limiter: redis, sliding_window, token_bucket, memory
		Type          string        `yaml:"type" env:"LIMITER_TYPE" env-default:"sliding_window"`
		DefaultLimit  int           `yaml:"default_limit" env:"LIMITER_DEFAULT_LIMIT" env-default:"100"`
		DefaultWindow time.Duration `yaml:"default_window" env:"LIMITER_DEFAULT_WINDOW" env-default:"60s"`
	} `yaml:"limiter"`
	Redis struct {
		Address  string `yaml:"address" env:"REDIS_ADDRESS" env-default:"localhost:6379"`
		Password string `yaml:"password" env:"REDIS_PASSWORD" env-default:""`
	} `yaml:"redis"`
}

func MustLoad(configPath string) *Config {
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("failed to read config: %v", err)
	}
	return &cfg
}
