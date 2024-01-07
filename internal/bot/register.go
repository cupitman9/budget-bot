package bot

import (
	"fmt"
	"gopkg.in/tucnak/telebot.v2"
	"log"
	"strconv"
	"strings"
	"telegram-budget-bot/internal/model"
	"telegram-budget-bot/internal/storage"
	"time"
)

func RegisterHandlers(b *telebot.Bot, storageInstance *storage.Storage, userSessions map[int64]*model.UserSession) {
	b.Handle("/start", func(m *telebot.Message) {
		handleStart(b, m, storageInstance)
	})

	b.Handle("/add_category", func(m *telebot.Message) {
		userSessions[m.Sender.ID] = &model.UserSession{
			State: model.StateAwaitingNewCategoryName,
		}
		b.Send(m.Sender, "Введите название новой категории:")
	})

	b.Handle("/show_categories", func(m *telebot.Message) {
		handleShowCategories(b, m, storageInstance)
	})

	b.Handle("/stats", func(m *telebot.Message) {
		handleStatsButtons(b, m)
	})

	b.Handle(telebot.OnText, func(m *telebot.Message) {
		handleOnText(b, m, storageInstance, userSessions)
	})

	b.Handle(telebot.OnCallback, func(c *telebot.Callback) {
		handleCallback(b, c, storageInstance, userSessions)
	})
}

func handleOnText(b *telebot.Bot, m *telebot.Message, storageInstance *storage.Storage, userSessions map[int64]*model.UserSession) {
	if _, err := strconv.ParseFloat(m.Text, 64); err == nil {
		handleIncomeExpenseButtons(b, m)
	}
	session, exists := userSessions[m.Sender.ID]
	if !exists {
		// Если сессии нет, можно обрабатывать сообщение как обычно
		// Например, можно обрабатывать общие запросы или команды
		// ...
		return
	}

	switch session.State {
	case model.StateAwaitingRenameCategory:
		handleAwaitingRenameCategory(b, m, storageInstance, session, userSessions)
	case model.StateAwaitingNewCategoryName:
		handleAwaitingNewCategoryName(b, m, storageInstance, userSessions)
	default:
		//handleDefaultText(b, m)
	}
}

func handleCallback(b *telebot.Bot, c *telebot.Callback, storageInstance *storage.Storage, userSessions map[int64]*model.UserSession) {
	x := strings.ReplaceAll(c.Data, "\f", "")
	prefixes := strings.Split(x, ":")
	if len(prefixes) == 0 {
		return
	}

	switch prefixes[0] {
	case "rename":
		handleRenameCallback(b, c, userSessions, prefixes[1])
	case "delete":

		handleDeleteCallback(b, c, storageInstance, prefixes[1])
	case "expense":
		handleTransactionCategories(b, c, storageInstance)
	case "income":

		handleTransactionCategories(b, c, storageInstance)
	case "transaction":

		handleTransactionCallback(b, c, storageInstance)
	case "today":

		handleTodayCallback(b, c, storageInstance, userSessions)
	default:
		b.Send(c.Sender, "Неизвестная команда")
	}
}

func handleAwaitingRenameCategory(b *telebot.Bot, m *telebot.Message, storageInstance *storage.Storage, session *model.UserSession, userSessions map[int64]*model.UserSession) {
	newCategoryName := m.Text
	categoryId := session.CategoryID
	err := storageInstance.RenameCategory(int64(categoryId), newCategoryName)
	if err != nil {
		b.Send(m.Sender, "Ошибка при переименовании категории: "+err.Error())
	} else {
		b.Send(m.Sender, "Категория успешно переименована в '"+newCategoryName+"'")
	}
	delete(userSessions, m.Sender.ID)
}

func handleAwaitingNewCategoryName(b *telebot.Bot, m *telebot.Message, storageInstance *storage.Storage, userSessions map[int64]*model.UserSession) {
	categoryName := m.Text
	err := storageInstance.AddCategory(model.Category{
		Name:   categoryName,
		ChatID: m.Chat.ID,
	})
	if err != nil {
		b.Send(m.Sender, "Ошибка при добавлении категории: "+err.Error())
	} else {
		b.Send(m.Sender, "Категория '"+categoryName+"' успешно добавлена.")
	}
	delete(userSessions, m.Sender.ID)
}

func handleRenameCallback(b *telebot.Bot, c *telebot.Callback, userSessions map[int64]*model.UserSession, id string) {
	categoryId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		b.Send(c.Sender, "Ошибка формата ID категории")
		return
	}
	userSessions[c.Sender.ID] = &model.UserSession{
		State:      model.StateAwaitingRenameCategory,
		CategoryID: int(categoryId),
	}
	b.Send(c.Sender, "Введите новое название категории:")
}

func handleDeleteCallback(b *telebot.Bot, c *telebot.Callback, storageInstance *storage.Storage, id string) {
	categoryId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		b.Send(c.Sender, "Ошибка при удалении категории.")
		return
	}
	err = storageInstance.DeleteCategory(c.Sender.ID, categoryId)
	if err != nil {
		b.Send(c.Sender, "Ошибка при удалении категории: "+err.Error())
	} else {
		b.Send(c.Sender, "Категория удалена.")
	}
}

func handleTransactionCallback(b *telebot.Bot, c *telebot.Callback, storageInstance *storage.Storage) {
	x := strings.ReplaceAll(c.Data, "\f", "")
	prefixes := strings.Split(strings.TrimSpace(x), ":")
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
		b.Send(c.Sender, "Ошибка при создании и сохранении транзакции")
	} else {
		b.Send(c.Sender, fmt.Sprintf("Транзакция на сумму %s в категорию %q добавлена.", prefixes[3], prefixes[2]))
	}
}

func handleTodayCallback(b *telebot.Bot, c *telebot.Callback, storageInstance *storage.Storage, userSessions map[int64]*model.UserSession) {
	var startDate, endDate time.Time
	now := time.Now()
	startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endDate = startDate.Add(24 * time.Hour)
	handleStats(b, c.Sender, storageInstance, startDate, endDate)
}
