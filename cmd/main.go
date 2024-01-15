package main

import (
	"context"
	"log"
	"time"

	"gopkg.in/tucnak/telebot.v2"

	"github.com/jackc/pgx/v4/pgxpool"

	"budget-bot/internal/bot"
	"budget-bot/internal/config"
	"budget-bot/internal/logger"
	"budget-bot/internal/model"
	"budget-bot/internal/storage"
)

var userSessions = make(map[int64]*model.UserSession)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("error loading configuration: %v", err)
	}

	appLogger := logger.New(cfg.LogLevel)

	if cfg.PostgresDSN == "" {
		appLogger.Fatal("POSTGRES_DSN is not set")
	}

	poolConfig, err := pgxpool.ParseConfig(cfg.PostgresDSN)
	if err != nil {
		appLogger.Fatalf("error parsing database configuration: %v", err)
	}
	pool, err := pgxpool.ConnectConfig(context.Background(), poolConfig)
	if err != nil {
		appLogger.Fatalf("unable to connect to database: %v", err)
	}
	defer pool.Close()

	storageInstance, err := storage.NewStorage(pool)
	if err != nil {
		appLogger.Fatalf("unable to initialize storage: %v", err)
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
