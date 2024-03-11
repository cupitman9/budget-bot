package config

import (
	"fmt"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	BotToken    string `env:"TELEGRAM_BOT_TOKEN,required"`
	PostgresDSN string `env:"POSTGRES_DSN,required"`
	LogLevel    string `env:"LOG_LEVEL" envDefault:"info"`
}

func LoadConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("error parsing config: %w", err)
	}
	return cfg, nil
}
