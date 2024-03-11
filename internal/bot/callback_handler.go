package bot

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/telebot.v3"

	"github.com/cupitman9/budget-bot/internal/model"
	"github.com/cupitman9/budget-bot/internal/storage"
)

type callbackHandler struct {
	b               *telebot.Bot
	storageInstance *storage.Storage
	log             *log.Logger
}

func newCallbackHandler(b *telebot.Bot, storageInstance *storage.Storage, log *log.Logger) *callbackHandler {
	return &callbackHandler{b: b, storageInstance: storageInstance, log: log}
}

func (h *callbackHandler) handleCallback(c *telebot.Callback) error {
	x := strings.ReplaceAll(c.Data, "\f", "") // telegram or this lib puts \f to data
	prefixes := strings.Split(x, ":")
	if len(prefixes) == 0 {
		return errors.New("no prefixes after splitting")
	}

	switch prefixes[0] {
	case "rename":
		err := h.handleRenameCallback(c, prefixes[1])
		if err != nil {
			return fmt.Errorf("error handling rename callback: %w", err)
		}
	case "delete":
		err := h.handleDeleteCallback(c, prefixes[1])
		if err != nil {
			return fmt.Errorf("error handling delete callback: %w", err)
		}
	case "expense":
		err := h.handleTransactionCategories(c)
		if err != nil {
			return fmt.Errorf("error handling expense callback: %w", err)
		}
	case "income":
		err := h.handleTransactionCategories(c)
		if err != nil {
			return fmt.Errorf("error handling income callback: %w", err)
		}
	case "transaction":
		err := h.handleTransactionCallback(c)
		if err != nil {
			return fmt.Errorf("error handling transaction callback: %w", err)
		}
	case "today":
		err := h.handleTodayCallback(c)
		if err != nil {
			return fmt.Errorf("error handling today callback: %w", err)
		}
	case "period":
		userSessions[c.Sender.ID] = &model.UserSession{State: model.StateAwaitingPeriod}
		_, err := h.b.Send(c.Sender, "Введите период в формате ДД.ММ.ГГГГ-ДД.ММ.ГГГГ:")
		if err != nil {
			return fmt.Errorf("error sending message to choose period: %w", err)
		}
	default:
		_, err := h.b.Send(c.Sender, "Команда не распознана. Пожалуйста, используйте одну из доступных команд.")
		if err != nil {
			return fmt.Errorf("error sending message for undefined callback action: %w", err)
		}
	}
	return nil
}

func (h *callbackHandler) handleTransactionCategories(c *telebot.Callback) error {
	categories, err := h.storageInstance.GetCategoriesByChatID(c.Sender.ID)
	if err != nil {
		_, sendErr := h.b.Send(c.Sender, "Ошибка при получении категорий.")
		if sendErr != nil {
			return fmt.Errorf("%v: %w", err, sendErr)
		}
		return err
	}

	if len(categories) == 0 {
		_, err := h.b.Send(c.Sender, "Категории отсутствуют.")
		if err != nil {
			return err
		}
		return nil
	}

	h.log.Infof("found %d categories", len(categories))

	markup := &telebot.ReplyMarkup{}
	var allRows []telebot.Row
	var row telebot.Row

	for i, category := range categories {
		btn := markup.Data(category.Name, "transaction:"+strconv.Itoa(int(category.ID))+":"+c.Data)
		row = append(row, btn)

		if (i+1)%3 == 0 || i == len(categories)-1 {
			allRows = append(allRows, row)
			row = telebot.Row{}
		}
	}
	markup.Inline(allRows...)
	_, err = h.b.Edit(c.Message, "Выберите категорию:", markup)
	if err != nil {
		return err
	}
	return nil
}

func (h *callbackHandler) handleRenameCallback(c *telebot.Callback, id string) error {
	categoryId, err := parseCategoryId(id)
	if err != nil {
		_, sendErr := h.b.Send(c.Sender, "Ошибка формата ID категории")
		if sendErr != nil {
			return fmt.Errorf("%v: %w", err, sendErr)
		}
		return err
	}
	userSessions[c.Sender.ID] = &model.UserSession{
		State:      model.StateAwaitingRenameCategory,
		CategoryID: int(categoryId),
	}
	_, err = h.b.Send(c.Sender, "Введите новое название категории:")
	if err != nil {
		return err
	}
	return nil
}

func (h *callbackHandler) handleDeleteCallback(c *telebot.Callback, id string) error {
	categoryId, err := parseCategoryId(id)
	if err != nil {
		_, sendErr := h.b.Send(c.Sender, "Ошибка при удалении категории.")
		if sendErr != nil {
			return fmt.Errorf("%v: %w", err, sendErr)
		}
		return err
	}
	err = h.storageInstance.DeleteCategory(c.Sender.ID, categoryId)
	if err != nil {
		_, sendErr := h.b.Send(c.Sender, "Ошибка при удалении категории: "+err.Error())
		if sendErr != nil {
			return fmt.Errorf("%v: %w", err, sendErr)
		}
	}
	_, err = h.b.Send(c.Sender, "Категория удалена.")
	if err != nil {
		return err
	}
	return nil
}

func (h *callbackHandler) handleTransactionCallback(c *telebot.Callback) error {
	x := strings.ReplaceAll(c.Data, "\f", "")
	prefixes := strings.Split(strings.TrimSpace(x), ":")
	categoryId, err := strconv.ParseInt(prefixes[1], 10, 64)
	if err != nil {
		_, sendErr := h.b.Send(c.Sender, "Ошибка при обработке категории")
		if sendErr != nil {
			return fmt.Errorf("%v: %w", err, sendErr)
		}
		return nil
	}

	amount, err := strconv.ParseFloat(prefixes[3], 10)
	if err != nil {
		_, sendErr := h.b.Send(c.Sender, "Ошибка при обработке суммы")
		if sendErr != nil {
			return fmt.Errorf("%v: %w", err, sendErr)
		}
		return nil
	}

	err = h.handleTransaction(c.Sender.ID, categoryId, amount, prefixes[2])
	if err != nil {
		_, sendErr := h.b.Send(c.Sender, "Ошибка при создании и сохранении транзакции")
		if sendErr != nil {
			return fmt.Errorf("%v: %w", err, sendErr)
		}
	}
	_, err = h.b.Send(c.Sender, fmt.Sprintf(
		"Транзакция на сумму %s в категорию %q добавлена.",
		prefixes[3],
		prefixes[2],
	))
	if err != nil {
		return err
	}
	return nil
}

func (h *callbackHandler) handleTodayCallback(c *telebot.Callback) error {
	var startDate, endDate time.Time
	now := time.Now()
	startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endDate = startDate.Add(24 * time.Hour)
	err := h.handleStats(c.Sender, startDate, endDate)
	if err != nil {
		return err
	}
	return nil
}

func (h *callbackHandler) handleTransaction(senderId, categoryId int64, amount float64, categoryType string) error {
	transaction := model.Transaction{
		UserChat:        senderId,
		CategoryID:      categoryId,
		Amount:          amount,
		TransactionType: categoryType,
		CreatedAt:       time.Now(),
	}

	if err := h.storageInstance.AddTransaction(transaction); err != nil {
		return err
	}

	return nil
}

func (h *callbackHandler) handleStats(sender *telebot.User, startDate, endDate time.Time) error {
	incomeCategories, expenseCategories, err := h.storageInstance.GetTransactionsStatsByCategory(sender.ID, startDate, endDate)
	if err != nil {
		_, sendErr := h.b.Send(sender, "Ошибка при получении статистики: "+err.Error())
		if sendErr != nil {
			return fmt.Errorf("%v: %w", err, sendErr)
		}
		return err
	}

	response := getStats(incomeCategories, expenseCategories)

	_, err = h.b.Send(sender, response, &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
	if err != nil {
		return err
	}
	return nil
}
