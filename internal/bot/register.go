package bot

import (
	"fmt"
	"gopkg.in/tucnak/telebot.v2"
	"log"
	"strconv"
	"strings"
	"telegram-budget-bot/internal/storage"
	"time"
)

func RegisterHandlers(b *telebot.Bot, storageInstance *storage.Storage) {
	b.Handle("/start", func(m *telebot.Message) {
		handleStart(b, m, storageInstance)
	})

	b.Handle("/add_category", func(m *telebot.Message) {
		handleAddCategory(b, m, storageInstance)
	})

	b.Handle("/show_categories", func(m *telebot.Message) {
		handleShowCategories(b, m, storageInstance)
	})

	b.Handle("/stats", func(m *telebot.Message) {
		handleStatsButtons(b, m)
	})

	b.Handle(telebot.OnText, func(m *telebot.Message) {
		if _, err := strconv.ParseFloat(m.Text, 64); err == nil {
			handleIncomeExpenseButtons(b, m)
		}
	})

	b.Handle(telebot.OnCallback, func(c *telebot.Callback) {
		handleCallback(b, c, storageInstance)
	})
}

func handleCallback(b *telebot.Bot, c *telebot.Callback, storageInstance *storage.Storage) {
	x := strings.ReplaceAll(c.Data, "\f", "")
	prefixes := strings.Split(x, ":")
	if len(prefixes) == 0 {
		return
	}
	switch prefixes[0] {

	case "expense":
		handleTransactionCategories(b, c, storageInstance)

	case "income":
		handleTransactionCategories(b, c, storageInstance)

	case "transaction":
		categoryId, err := strconv.ParseInt(prefixes[1], 10, 64)
		if err != nil {
			log.Printf("error parse categoryId from prefixes: %s", prefixes[1])
			b.Send(c.Sender, "Ошибка при обработке категории")
			return
		}

		amount, err := strconv.ParseFloat(prefixes[3], 10)
		if err != nil {
			log.Printf("error parse amount from prefixes: %s", prefixes[3])
			b.Send(c.Sender, "Ошибка при обработке суммы")
			return
		}

		err = handleTransaction(c.Sender.ID, categoryId, amount, prefixes[2], storageInstance)
		if err != nil {
			b.Send(c.Sender, "Ошибка при создание и сохранение транзакции")
			return
		}
		b.Send(c.Sender, fmt.Sprintf("Транзакция на сумму %s в категорию %q добавлена.", prefixes[3], prefixes[2]))

	case "today":
		var startDate, endDate time.Time
		now := time.Now()
		startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		endDate = startDate.Add(24 * time.Hour)
		handleStats(b, c.Sender, storageInstance, startDate, endDate)

	default:
		b.Send(c.Sender, "Неизвестная команда")
	}
}
