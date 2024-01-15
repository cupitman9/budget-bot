package main

import (
	"fmt"
	"log"
	"time"

	"gopkg.in/tucnak/telebot.v2"

	"telegram-budget-bot/internal/bot"
	"telegram-budget-bot/internal/config"
	"telegram-budget-bot/internal/logger"
	"telegram-budget-bot/internal/model"
	"telegram-budget-bot/internal/storage"
)

var userSessions = make(map[int64]*model.UserSession)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("error loading configuration: %v", err)
	}
	appLogger := logger.New(cfg.LogLevel)

	dbInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	storageInstance, err := storage.NewStorage(dbInfo)
	if err != nil {
		appLogger.Fatalf("unable to connect to database: %v", err)
	}

	botSettings := telebot.Settings{
		Token:  cfg.BotToken,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}
	botAPI, err := telebot.NewBot(botSettings)
	if err != nil {
		appLogger.Fatalf("error creating bot instance: %v", err)
	}

	bot.RegisterHandlers(botAPI, storageInstance, appLogger, userSessions)
	appLogger.Info("bot start")
	botAPI.Start()
}
