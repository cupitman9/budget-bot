package config

import (
	"os"
)

type Config struct {
	BotToken    string
	BotDebug    bool
	PostgresDSN string
	LogLevel    string
}

func LoadConfig() (*Config, error) {
	return &Config{
		BotToken:    os.Getenv("TELEGRAM_BOT_TOKEN"),
		BotDebug:    os.Getenv("BOT_DEBUG") == "true",
		PostgresDSN: os.Getenv("POSTGRES_DSN"),
		LogLevel:    os.Getenv("LOG_LEVEL"),
	}, nil
}
