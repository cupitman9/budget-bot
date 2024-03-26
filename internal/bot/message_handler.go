package bot

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/telebot.v3"

	"github.com/cupitman9/budget-bot/internal/model"
	"github.com/cupitman9/budget-bot/internal/storage"
)

type messageHandler struct {
	b               *telebot.Bot
	storageInstance *storage.Storage
	log             *logrus.Logger
}

func newMessageHandler(b *telebot.Bot, storageInstance *storage.Storage, log *logrus.Logger) *messageHandler {
	return &messageHandler{b: b, storageInstance: storageInstance, log: log}
}

func (h *messageHandler) handleOnText(m *telebot.Message) error {
	if _, err := strconv.ParseFloat(m.Text, 64); err == nil {
		expErr := h.handleIncomeExpenseButtons(m)
		if expErr != nil {
			return fmt.Errorf("%v: %w", err, expErr)
		}
		return err
	}

	session, ok := userSessions[m.Sender.ID]
	if ok {
		switch session.State {
		case model.StateAwaitingRenameCategory:
			err := h.handleAwaitingRenameCategory(m, session)
			if err != nil {
				return err
			}
			return nil
		case model.StateAwaitingNewCategoryName:
			err := h.handleAwaitingNewCategoryName(m)
			if err != nil {
				return err
			}
			return nil
		case model.StateAwaitingPeriod:
			err := h.handlePeriodInput(m)
			if err != nil {
				return err
			}
			return nil
		default:
			if _, err := h.b.Send(m.Sender, "Извините, я не понимаю эту команду."); err != nil {
				return err
			}
			return nil
		}
	}

	_, err := h.b.Send(m.Sender, "Извините, я не понимаю эту команду. Введите /help для списка команд.")
	if err != nil {
		return err
	}
	return nil
}

func (h *messageHandler) handleStart(m *telebot.Message) error {
	u, err := h.storageInstance.GetUserByChatID(m.Chat.ID)
	if err != nil {
		_, err := h.b.Send(m.Sender, "Ошибка при проверке существования пользователя:", err)
		if err != nil {
			return err
		}
	}

	if !u.IsEmpty() {
		h.log.WithField("chat_id", m.Chat.ID).Warn("found the same user")
		return nil
	}

	user := model.User{
		Username:  m.Sender.Username,
		ChatID:    m.Chat.ID,
		Language:  m.Sender.LanguageCode,
		CreatedAt: time.Now(),
	}

	if err := h.storageInstance.AddUser(user); err != nil {
		_, err := h.b.Send(m.Sender, "Ошибка при добавлении пользователя:", err)
		if err != nil {
			return err
		}
	}

	defaultCategory := model.Category{
		Name:   "Общее",
		ChatID: m.Chat.ID,
	}
	if err := h.storageInstance.AddCategory(defaultCategory); err != nil {
		_, err := h.b.Send(m.Sender, "Ошибка при добавлении общей категории:", err)
		if err != nil {
			return err
		}
	}

	welcomeText := "Привет! Нажмите /help для подробной информации"
	_, err = h.b.Send(m.Sender, welcomeText)
	if err != nil {
		return err
	}
	return nil
}

func (h *messageHandler) handleHelp(m *telebot.Message) error {
	helpMessage := "Команды бота:\n" +
		"/start - начать работу с ботом\n" +
		"/add_category - добавить новую категорию\n" +
		"/show_categories - показать все категории\n" +
		"/stats - показать статистику\n" +
		"/help - показать эту справку\n" +
		"...\n" +
		"Для добавления транзакции просто введите сумму."

	_, err := h.b.Send(m.Sender, helpMessage)
	if err != nil {
		return err
	}
	return nil
}

func (h *messageHandler) handleShowCategories(m *telebot.Message) error {
	categories, err := h.storageInstance.GetCategoriesByChatID(m.Chat.ID)
	if err != nil {
		_, sendErr := h.b.Send(m.Sender, fmt.Sprintf("Ошибка при получении категорий: %v", err))
		if sendErr != nil {
			return fmt.Errorf("%v: %w", err, sendErr)
		}
		return err
	}

	if len(categories) == 0 {
		h.log.Info("no categories found")
		if _, err := h.b.Send(m.Sender, "Категории отсутствуют."); err != nil {
			return err
		}
	}

	markup := &telebot.ReplyMarkup{}
	var rows []telebot.Row
	for _, category := range categories {
		btnCategory := markup.Text(category.Name)
		btnRename := markup.Data("Переименовать", "rename:"+strconv.Itoa(int(category.ID)))
		rows = append(rows, markup.Row(btnCategory), markup.Row(btnRename))
	}

	markup.Inline(rows...)
	_, err = h.b.Send(m.Sender, "Категории:", markup)
	if err != nil {
		return err
	}

	return nil
}

func (h *messageHandler) handleIncomeExpenseButtons(m *telebot.Message) error {
	markup := &telebot.ReplyMarkup{}
	btnIncome := markup.Data("Доход", strconv.Itoa(int(model.TransactionTypeIncome))+":"+m.Text)
	btnExpense := markup.Data("Расход", strconv.Itoa(int(model.TransactionTypeExpense))+":"+m.Text)
	markup.Inline(markup.Row(btnIncome, btnExpense))
	_, err := h.b.Send(m.Sender, "Выберите тип транзакции:", markup)
	if err != nil {
		return err
	}
	return nil
}

func (h *messageHandler) handleStatsButtons(m *telebot.Message) error {
	markup := &telebot.ReplyMarkup{}
	btnToday := markup.Data("Сегодня", "today")
	btnPeriod := markup.Data("Период", "period")
	markup.Inline(markup.Row(btnToday, btnPeriod))
	_, err := h.b.Send(m.Sender, "Выберите период:", markup)
	if err != nil {
		return err
	}
	return nil
}

func (h *messageHandler) handleAwaitingRenameCategory(m *telebot.Message, session *model.UserSession) error {
	err := h.storageInstance.RenameCategory(int64(session.CategoryID), m.Text)
	if err != nil {
		_, sendErr := h.b.Send(m.Sender, "Ошибка при переименовании категории: "+err.Error())
		if sendErr != nil {
			return fmt.Errorf("%v: %w", err, sendErr)
		}
	}

	_, err = h.b.Send(m.Sender, "Категория успешно переименована в '"+m.Text+"'")
	if err != nil {
		return err
	}

	delete(userSessions, m.Sender.ID)
	return nil
}

func (h *messageHandler) handleAwaitingNewCategoryName(m *telebot.Message) error {
	err := h.storageInstance.AddCategory(model.Category{
		Name:   m.Text,
		ChatID: m.Chat.ID,
	})
	if err != nil {
		_, sendErr := h.b.Send(m.Sender, "Ошибка при добавлении категории: "+err.Error())
		if sendErr != nil {
			return fmt.Errorf("%v: %w", err, sendErr)
		}
	}

	_, err = h.b.Send(m.Sender, "Категория '"+m.Text+"' успешно добавлена.")
	if err != nil {
		return err
	}

	delete(userSessions, m.Sender.ID)
	return nil
}

func (h *messageHandler) handlePeriodInput(m *telebot.Message) error {
	periodParts := strings.Split(m.Text, "-")
	if len(periodParts) != 2 {
		_, err := h.b.Send(m.Sender, "Неправильный формат периода. Используйте формат ДД.ММ.ГГГГ-ДД.ММ.ГГГГ.")
		if err != nil {
			return err
		}
		return nil
	}

	startDate, errStart := time.Parse("02.01.2006", periodParts[0])
	endDate, errEnd := time.Parse("02.01.2006", periodParts[1])
	if errStart != nil || errEnd != nil {
		_, sendErr := h.b.Send(m.Sender, "Ошибка в датах. Используйте формат ДД.ММ.ГГГГ.")
		if sendErr != nil {
			return fmt.Errorf("%v, %v: %w", errStart, errEnd, sendErr)
		}
		return fmt.Errorf("%v, %v", errStart, errEnd)
	}

	err := h.handleStats(m.Sender, startDate, endDate)
	if err != nil {
		return err
	}
	delete(userSessions, m.Sender.ID)
	return nil
}

func (h *messageHandler) handleStats(sender *telebot.User, startDate, endDate time.Time) error {
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
