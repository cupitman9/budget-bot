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

	b.Handle(telebot.OnText, func(m *telebot.Message) {
		if _, err := strconv.ParseFloat(m.Text, 64); err == nil {
			handleIncomeExpenseButtons(b, m, storageInstance)
		}
	})
	b.Handle(telebot.OnCallback, func(c *telebot.Callback) {
		x := strings.ReplaceAll(c.Data, "\f", "")
		prefixes := strings.Split(x, ":")
		if len(prefixes) == 0 {
			return
		}
		switch {
		case prefixes[0] == "expense":
			handleExpense(b, c, storageInstance)
		case prefixes[0] == "category":
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
			b.Send(c.Sender, fmt.Sprintf("Расход на сумму %s в категории %q добавлен.", prefixes[3], prefixes[2]))

		default:
			fmt.Println("DEFAULT")
		}
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

func handleAddCategory(b *telebot.Bot, m *telebot.Message, storageInstance *storage.Storage) {
	categoryName := strings.TrimSpace(m.Payload)
	if categoryName == "" {
		b.Send(m.Sender, "Пожалуйста, укажите название категории.")
		return
	}

	category := model.Category{
		Name:   categoryName,
		ChatID: m.Chat.ID,
	}

	if err := storageInstance.AddCategory(category); err != nil {
		b.Send(m.Sender, fmt.Sprintf("Ошибка при добавлении категории: %v", err))
		return
	}

	b.Send(m.Sender, fmt.Sprintf("Категория '%s' успешно добавлена!", categoryName))
}

func handleShowCategories(b *telebot.Bot, m *telebot.Message, storageInstance *storage.Storage) {
	categories, err := storageInstance.GetCategoriesByChatID(m.Chat.ID)
	if err != nil {
		b.Send(m.Sender, fmt.Sprintf("Ошибка при получении категорий: %v", err))
		return
	}

	var response string
	if len(categories) == 0 {
		response = "Категории отсутствуют."
	} else {
		response = "Доступные категории:\n"
		for _, c := range categories {
			response += fmt.Sprintf("- %s\n", c.Name)
		}
	}

	b.Send(m.Sender, response)
}

func handleIncomeExpenseButtons(b *telebot.Bot, m *telebot.Message, storageInstance *storage.Storage) {
	markup := &telebot.ReplyMarkup{}
	btnIncome := markup.Data("Доход", "income:"+m.Text)
	btnExpense := markup.Data("Расход", "expense:"+m.Text)
	markup.Inline(markup.Row(btnIncome, btnExpense))
	b.Send(m.Sender, "Выберите тип транзакции:", markup)
}

func handleExpense(b *telebot.Bot, c *telebot.Callback, storageInstance *storage.Storage) {
	categories, err := storageInstance.GetCategoriesByChatID(c.Sender.ID)
	if err != nil {
		b.Send(c.Sender, "Ошибка при получении категорий расходов.")
		return
	}

	markup := &telebot.ReplyMarkup{}
	var btns []telebot.Btn

	for _, category := range categories {
		btn := markup.Data(category.Name, "category:"+strconv.Itoa(int(category.ID))+":"+c.Data)
		btns = append(btns, btn)
	}
	markup.Inline(markup.Row(btns...))
	b.Edit(c.Message, "Выберите категорию расхода:", markup)
}

func handleTransaction(senderId, categoryId int64, amount float64, categoryType string, storageInstance *storage.Storage) error {
	transaction := model.Transaction{
		UserChat:        senderId,
		CategoryID:      categoryId,
		Amount:          amount,
		TransactionType: categoryType,
		CreatedAt:       time.Now(),
	}

	if err := storageInstance.AddTransaction(transaction); err != nil {
		return err
	}
	return nil
}
