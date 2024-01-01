package config

import (
	"log"
	"os"
)

type Config struct {
	BotToken string
}

func LoadConfig() (*Config, error) {
	token, exists := os.LookupEnv("TELEGRAM_BOT_TOKEN")
	if !exists {
		log.Fatal("Ошибка: переменная окружения TELEGRAM_BOT_TOKEN не задана")
	}

	return &Config{
		BotToken: token,
	}, nil
}
