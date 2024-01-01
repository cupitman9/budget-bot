package main

import (
	"gopkg.in/tucnak/telebot.v2"
	"log"
	"telegram-budget-bot/internal/bot"
	"telegram-budget-bot/internal/config"
	"telegram-budget-bot/internal/logger"
	"time"
)

func main() {
	logger.SetupLogger("application.log")

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Ошибка при загрузке конфигурации: %v", err)
	}

	b, err := telebot.NewBot(telebot.Settings{
		Token:  cfg.BotToken,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		log.Fatal(err)
	}

	bot.SetupHandlers(b)

	b.Start()
}
