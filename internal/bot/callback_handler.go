package bot

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/tucnak/telebot.v2"

	"telegram-budget-bot/internal/model"
	"telegram-budget-bot/internal/storage"
)

type callbackHandler struct {
	b               *telebot.Bot
	storageInstance *storage.Storage
	log             *logrus.Logger
}

func newCallbackHandler(b *telebot.Bot, storageInstance *storage.Storage, log *logrus.Logger) *callbackHandler {
	return &callbackHandler{
		b:               b,
		storageInstance: storageInstance,
		log:             log,
	}
}

func (h *callbackHandler) handleCallback(c *telebot.Callback, userSessions map[int64]*model.UserSession) {
	x := strings.ReplaceAll(c.Data, "\f", "")
	prefixes := strings.Split(x, ":")
	if len(prefixes) == 0 {
		h.log.Warnf("Empty callback from user %d", c.Sender.ID)
		return
	}

	switch prefixes[0] {
	case "rename":
		h.log.Infof("Renaming category through callback by user %d", c.Sender.ID)
		h.handleRenameCallback(c, userSessions, prefixes[1])
	case "delete":
		h.log.Infof("Deleting category through callback by user %d", c.Sender.ID)
		h.handleDeleteCallback(c, prefixes[1])
	case "expense":
		h.log.Infof("Processing expenses through callback by user %d", c.Sender.ID)
		h.handleTransactionCategories(c)
	case "income":
		h.log.Infof("Processing income through callback by user %d", c.Sender.ID)
		h.handleTransactionCategories(c)
	case "transaction":
		h.log.Infof("Processing transaction through callback by user %d", c.Sender.ID)
		h.handleTransactionCallback(c)
	case "today":
		h.log.Infof("Processing today's statistics through callback by user %d", c.Sender.ID)
		h.handleTodayCallback(c)
	case "period":
		h.log.Infof("User %d entering a period through callback", c.Sender.ID)
		userSessions[c.Sender.ID] = &model.UserSession{State: model.StateAwaitingPeriod}
		h.b.Send(c.Sender, "Введите период в формате ДД.ММ.ГГГГ-ДД.ММ.ГГГГ:")
	default:
		h.log.Warnf("Unknown command through callback from user %d", c.Sender.ID)
		h.b.Send(c.Sender, "Команда не распознана. Пожалуйста, используйте одну из доступных команд.")
	}
}

func (h *callbackHandler) handleTransactionCategories(c *telebot.Callback) {
	h.log.Info("Retrieving transaction categories")
	categories, err := h.storageInstance.GetCategoriesByChatID(c.Sender.ID)
	if err != nil {
		h.log.Errorf("Error retrieving categories: %v", err)
		h.b.Send(c.Sender, "Ошибка при получении категорий.")
		return
	}

	if len(categories) == 0 {
		h.log.Info("No categories found")
		h.b.Send(c.Sender, "Категории отсутствуют.")
		return
	}

	h.log.Infof("Found %d categories", len(categories))

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
	h.b.Edit(c.Message, "Выберите категорию:", markup)
}

func (h *callbackHandler) handleRenameCallback(c *telebot.Callback, userSessions map[int64]*model.UserSession, id string) {
	h.log.Infof("Processing category rename request from user %d", c.Sender.ID)

	categoryId, err := parseCategoryId(id)
	if err != nil {
		h.b.Send(c.Sender, "Ошибка формата ID категории")
		return
	}
	userSessions[c.Sender.ID] = &model.UserSession{
		State:      model.StateAwaitingRenameCategory,
		CategoryID: int(categoryId),
	}
	h.b.Send(c.Sender, "Введите новое название категории:")
}

func (h *callbackHandler) handleDeleteCallback(c *telebot.Callback, id string) {
	h.log.Infof("Processing delete category request from user %d", c.Sender.ID)

	categoryId, err := parseCategoryId(id)
	if err != nil {
		h.b.Send(c.Sender, "Ошибка при удалении категории.")
		return
	}
	err = h.storageInstance.DeleteCategory(c.Sender.ID, categoryId)
	if err != nil {
		h.b.Send(c.Sender, "Ошибка при удалении категории: "+err.Error())
	} else {
		h.b.Send(c.Sender, "Категория удалена.")
	}
}

func (h *callbackHandler) handleTransactionCallback(c *telebot.Callback) {
	h.log.Infof("Processing transaction for user %d", c.Sender.ID)

	x := strings.ReplaceAll(c.Data, "\f", "")
	prefixes := strings.Split(strings.TrimSpace(x), ":")
	categoryId, err := strconv.ParseInt(prefixes[1], 10, 64)
	if err != nil {
		h.log.Errorf("Error parsing category ID from prefixes: %s", prefixes[1])
		h.b.Send(c.Sender, "Ошибка при обработке категории")
		return
	}

	amount, err := strconv.ParseFloat(prefixes[3], 10)
	if err != nil {
		h.log.Errorf("Error parsing amount from prefixes: %s", prefixes[3])
		h.b.Send(c.Sender, "Ошибка при обработке суммы")
		return
	}

	err = h.handleTransaction(c.Sender.ID, categoryId, amount, prefixes[2])
	if err != nil {
		h.b.Send(c.Sender, "Ошибка при создании и сохранении транзакции")
	} else {
		h.b.Send(c.Sender, fmt.Sprintf("Транзакция на сумму %s в категорию %q добавлена.", prefixes[3], prefixes[2]))
	}
}

func (h *callbackHandler) handleTodayCallback(c *telebot.Callback) {
	h.log.Infof("Processing today's stats for user %d", c.Sender.ID)

	var startDate, endDate time.Time
	now := time.Now()
	startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endDate = startDate.Add(24 * time.Hour)
	h.handleStats(c.Sender, startDate, endDate)
}

func (h *callbackHandler) handleTransaction(senderId, categoryId int64, amount float64, categoryType string) error {
	h.log.Infof("Handling transaction: SenderID=%d, CategoryID=%d, Amount=%.2f, Type=%s", senderId, categoryId, amount, categoryType)

	transaction := model.Transaction{
		UserChat:        senderId,
		CategoryID:      categoryId,
		Amount:          amount,
		TransactionType: categoryType,
		CreatedAt:       time.Now(),
	}

	if err := h.storageInstance.AddTransaction(transaction); err != nil {
		h.log.Errorf("Error adding transaction: %v", err)
		return err
	}

	h.log.Info("Transaction added successfully")
	return nil
}

func (h *callbackHandler) handleStats(sender *telebot.User, startDate, endDate time.Time) {
	h.log.Infof("Handling stats: SenderID=%d, StartDate=%s, EndDate=%s", sender.ID, startDate, endDate)
	incomeCategories, expenseCategories, err := h.storageInstance.GetTransactionsStatsByCategory(sender.ID, startDate, endDate)
	if err != nil {
		h.log.Errorf("Error retrieving statistics: %v", err)
		h.b.Send(sender, "Ошибка при получении статистики: "+err.Error())
		return
	}

	response := getStats(incomeCategories, expenseCategories)

	h.b.Send(sender, response, &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
}
