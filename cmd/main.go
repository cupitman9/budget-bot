package main

import (
	"fmt"
	"gopkg.in/tucnak/telebot.v2"
	"telegram-budget-bot/internal/bot"
	"telegram-budget-bot/internal/config"
	"telegram-budget-bot/internal/logger"
	"telegram-budget-bot/internal/model"
	"telegram-budget-bot/internal/storage"
	"time"
)

var userSessions = make(map[int64]*model.UserSession)

func main() {
	log := logger.GetLogger()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Ошибка при загрузке конфигурации: %v", err)
	}

	dbInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	storageInstance, err := storage.NewStorage(dbInfo)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}

	botSettings := telebot.Settings{
		Token:  cfg.BotToken,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}
	botAPI, err := telebot.NewBot(botSettings)
	if err != nil {
		log.Fatalf("Error creating bot instance: %v", err)
	}

	bot.RegisterHandlers(botAPI, storageInstance, userSessions)
	log.Info("Bot start")
	botAPI.Start()
}
