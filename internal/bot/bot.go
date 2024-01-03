package bot

import (
	"fmt"
	"gopkg.in/tucnak/telebot.v2"
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
		switch {
		case strings.HasPrefix(c.Data[1:8], "expense"):
			handleExpense(b, c, storageInstance)
		case strings.HasPrefix(c.Data[1:9], "category"):
			handleTransaction(b, c, storageInstance)
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
	btnIncome := markup.Data("Доход", "income", m.Text)
	btnExpense := markup.Data("Расход", "expense", m.Text)
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
		btn := markup.Data(category.Name, "category|"+strconv.Itoa(int(category.ID)), c.Data)
		btns = append(btns, btn)
	}
	markup.Inline(markup.Row(btns...))
	b.Edit(c.Message, "Выберите категорию расхода:", markup)
}

func handleTransaction(b *telebot.Bot, c *telebot.Callback, storageInstance *storage.Storage) {
	data := strings.Split(c.Data, "|")

	if len(data) != 4 || data[0][1:] != "category" {
		b.Send(c.Sender, "Неверный формат данных")
		return
	}
	categoryID, _ := strconv.Atoi(data[1])
	amount, _ := strconv.ParseFloat(data[3], 64)

	transaction := model.Transaction{
		UserChat:        c.Sender.ID,
		CategoryID:      int64(categoryID),
		Amount:          amount,
		TransactionType: "expense",
		CreatedAt:       time.Now(),
	}

	if err := storageInstance.AddTransaction(transaction); err != nil {
		b.Send(c.Sender, "Ошибка при добавлении транзакции.")
		return
	}

	b.Send(c.Sender, fmt.Sprintf("Расход на сумму %.2f в категории '%d' добавлен.", amount, categoryID))
}
