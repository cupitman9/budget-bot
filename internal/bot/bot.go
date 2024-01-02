package bot

import (
	"fmt"
	"gopkg.in/tucnak/telebot.v2"
	"telegram-budget-bot/internal/model"
	"telegram-budget-bot/internal/storage"
	"time"
)

func RegisterHandlers(b *telebot.Bot, storageInstance *storage.Storage) {
	b.Handle("/start", func(m *telebot.Message) {
		handleStart(b, m, storageInstance)
	})
}

func handleStart(b *telebot.Bot, m *telebot.Message, storageInstance *storage.Storage) {
	user := model.User{
		Username:  m.Sender.Username,
		ChatID:    m.Chat.ID,
		Language:  m.Sender.LanguageCode,
		CreatedAt: time.Now(),
	}

	if err := storageInstance.AddUser(user); err != nil {
		fmt.Println("Ошибка при добавлении пользователя:", err)
		return
	}

	welcomeText := "Привет! Нажмите /help для подробной информации"
	b.Send(m.Sender, welcomeText)
}
