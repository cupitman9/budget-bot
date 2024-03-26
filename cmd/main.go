package main

import (
	"context"
	"log"
	"time"

	"gopkg.in/telebot.v3"

	"github.com/cupitman9/budget-bot/internal/bot"
	"github.com/cupitman9/budget-bot/internal/config"
	"github.com/cupitman9/budget-bot/internal/logger"
	"github.com/cupitman9/budget-bot/internal/storage"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("error loading configuration: %v", err)
	}

	appLogger := logger.New(cfg.LogLevel)

	ctx := context.Background()
	appStorage, err := storage.NewStorage(ctx, cfg.PostgresDSN)
	if err != nil {
		appLogger.WithError(err).Fatal("error creating new storage")
	}
	defer appStorage.Close()

	botSettings := telebot.Settings{
		Token:  cfg.BotToken,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}
	botAPI, err := telebot.NewBot(botSettings)
	if err != nil {
		appLogger.WithError(err).Error("error creating bot instance")
		return
	}

	bot.RegisterHandlers(botAPI, appStorage, appLogger)
	appLogger.Info("bot starting")
	botAPI.Start()
}
