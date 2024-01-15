package bot

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/tucnak/telebot.v2"

	"budget-bot/internal/model"
	"budget-bot/internal/storage"
)

type messageHandler struct {
	b               *telebot.Bot
	storageInstance *storage.Storage
	log             *logrus.Logger
}

func newMessageHandler(b *telebot.Bot, storageInstance *storage.Storage, log *logrus.Logger) *messageHandler {
	return &messageHandler{b: b, storageInstance: storageInstance, log: log}
}

func (h *messageHandler) handleOnText(m *telebot.Message, userSessions map[int64]*model.UserSession) {
	if _, err := strconv.ParseFloat(m.Text, 64); err == nil {
		h.log.Infof("processing number from user %d: %s", m.Sender.ID, m.Text)
		h.handleIncomeExpenseButtons(m)
		return
	}
	session, exists := userSessions[m.Sender.ID]
	if exists {
		switch session.State {
		case model.StateAwaitingRenameCategory:
			h.log.Infof("renaming category by user %d", m.Sender.ID)
			h.handleAwaitingRenameCategory(m, session, userSessions)
			return
		case model.StateAwaitingNewCategoryName:
			h.log.Infof("adding a new category by user %d", m.Sender.ID)
			h.handleAwaitingNewCategoryName(m, userSessions)
			return
		case model.StateAwaitingPeriod:
			h.log.Infof("user %d entering a period", m.Sender.ID)
			h.handlePeriodInput(m, userSessions)
			return
		default:
			h.log.Warnf("unknown command from user %d: %s", m.Sender.ID, m.Text)
			h.b.Send(m.Sender, "Извините, я не понимаю эту команду.")
			return
		}
	}

	h.log.Warnf("unknown command from user %d: %s", m.Sender.ID, m.Text)
	h.b.Send(m.Sender, "Извините, я не понимаю эту команду. Введите /help для списка команд.")
}

func (h *messageHandler) handleStart(m *telebot.Message) {
	user := model.User{
		Username:  m.Sender.Username,
		ChatID:    m.Chat.ID,
		Language:  m.Sender.LanguageCode,
		CreatedAt: time.Now(),
	}

	if err := h.storageInstance.AddUser(user); err != nil {
		h.log.Errorf("error adding user: %v", err)
		h.b.Send(m.Sender, "Ошибка при добавлении пользователя:", err)
		return
	}

	h.log.Infof("user %s added successfully", user.Username)

	welcomeText := "Привет! Нажмите /help для подробной информации"
	h.b.Send(m.Sender, welcomeText)
}

func (h *messageHandler) handleHelp(m *telebot.Message) {
	h.log.Infof("processing /help for user %d", m.Sender.ID)

	helpMessage := "Команды бота:\n" +
		"/start - начать работу с ботом\n" +
		"/add_category - добавить новую категорию\n" +
		"/show_categories - показать все категории\n" +
		"/stats - показать статистику\n" +
		"/help - показать эту справку\n" +
		"...\n" +
		"Для добавления транзакции просто введите сумму."

	h.b.Send(m.Sender, helpMessage)
}

func (h *messageHandler) handleShowCategories(m *telebot.Message) {
	h.log.Info("retrieving categories")

	categories, err := h.storageInstance.GetCategoriesByChatID(m.Chat.ID)
	if err != nil {
		h.log.Errorf("error retrieving categories: %v", err)
		h.b.Send(m.Sender, fmt.Sprintf("Ошибка при получении категорий: %v", err))
		return
	}

	if len(categories) == 0 {
		h.log.Info("no categories found")
		h.b.Send(m.Sender, "Категории отсутствуют.")
		return
	}

	h.log.Infof("found %d categories", len(categories))

	markup := &telebot.ReplyMarkup{}
	var rows []telebot.Row
	for _, category := range categories {
		btnCategory := markup.Text(category.Name)
		btnRename := markup.Data("Переименовать", "rename:"+strconv.Itoa(int(category.ID)))
		btnDelete := markup.Data("Удалить", "delete:"+strconv.Itoa(int(category.ID)))
		rows = append(rows, markup.Row(btnCategory))
		rows = append(rows, markup.Row(btnRename, btnDelete))
	}

	markup.Inline(rows...)
	h.b.Send(m.Sender, "Категории:", markup)
}

func (h *messageHandler) handleIncomeExpenseButtons(m *telebot.Message) {
	h.log.Info("handling income and expense buttons")

	markup := &telebot.ReplyMarkup{}
	btnIncome := markup.Data("Доход", "income:"+m.Text)
	btnExpense := markup.Data("Расход", "expense:"+m.Text)
	markup.Inline(markup.Row(btnIncome, btnExpense))
	h.b.Send(m.Sender, "Выберите тип транзакции:", markup)
}

func (h *messageHandler) handleStatsButtons(m *telebot.Message) {
	h.log.Info("handling stats buttons")

	markup := &telebot.ReplyMarkup{}
	btnToday := markup.Data("Сегодня", "today")
	btnPeriod := markup.Data("Период", "period")
	markup.Inline(markup.Row(btnToday, btnPeriod))
	h.b.Send(m.Sender, "Выберите период:", markup)
}

func (h *messageHandler) handleAwaitingRenameCategory(m *telebot.Message, session *model.UserSession, userSessions map[int64]*model.UserSession) {
	h.log.Infof("renaming category for user %d", m.Sender.ID)

	newCategoryName := m.Text
	categoryId := session.CategoryID
	err := h.storageInstance.RenameCategory(int64(categoryId), newCategoryName)
	if err != nil {
		h.b.Send(m.Sender, "Ошибка при переименовании категории: "+err.Error())
	} else {
		h.b.Send(m.Sender, "Категория успешно переименована в '"+newCategoryName+"'")
	}
	delete(userSessions, m.Sender.ID)
}

func (h *messageHandler) handleAwaitingNewCategoryName(m *telebot.Message, userSessions map[int64]*model.UserSession) {
	h.log.Infof("adding new category for user %d", m.Sender.ID)

	categoryName := m.Text
	err := h.storageInstance.AddCategory(model.Category{
		Name:   categoryName,
		ChatID: m.Chat.ID,
	})
	if err != nil {
		h.b.Send(m.Sender, "Ошибка при добавлении категории: "+err.Error())
	} else {
		h.b.Send(m.Sender, "Категория '"+categoryName+"' успешно добавлена.")
	}
	delete(userSessions, m.Sender.ID)
}

func (h *messageHandler) handlePeriodInput(m *telebot.Message, userSessions map[int64]*model.UserSession) {
	h.log.Infof("processing period input from user %d", m.Sender.ID)

	periodParts := strings.Split(m.Text, "-")
	if len(periodParts) != 2 {
		h.b.Send(m.Sender, "Неправильный формат периода. Используйте формат ДД.ММ.ГГГГ-ДД.ММ.ГГГГ.")
		return
	}

	startDate, errStart := time.Parse("02.01.2006", periodParts[0])
	endDate, errEnd := time.Parse("02.01.2006", periodParts[1])
	if errStart != nil || errEnd != nil {
		h.b.Send(m.Sender, "Ошибка в датах. Используйте формат ДД.ММ.ГГГГ.")
		return
	}

	h.handleStats(m.Sender, startDate, endDate)
	delete(userSessions, m.Sender.ID)
}

func (h *messageHandler) handleStats(sender *telebot.User, startDate, endDate time.Time) {
	h.log.Infof("handling stats: SenderID=%d, StartDate=%s, EndDate=%s", sender.ID, startDate, endDate)
	incomeCategories, expenseCategories, err := h.storageInstance.GetTransactionsStatsByCategory(sender.ID, startDate, endDate)
	if err != nil {
		h.log.Errorf("error retrieving statistics: %v", err)
		h.b.Send(sender, "Ошибка при получении статистики: "+err.Error())
		return
	}

	response := getStats(incomeCategories, expenseCategories)

	h.b.Send(sender, response, &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
}
