package config

import (
	"os"
	"strconv"
)

type Config struct {
	BotToken   string
	BotDebug   bool
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
}

func LoadConfig() (*Config, error) {
	port, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		return nil, err
	}

	return &Config{
		BotToken:   os.Getenv("TELEGRAM_BOT_TOKEN"),
		BotDebug:   os.Getenv("BOT_DEBUG") == "true",
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     port,
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
	}, nil
}
