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

func handleStart(b *telebot.Bot, m *telebot.Message, storageInstance *storage.Storage) {
	user := model.User{
		Username:  m.Sender.Username,
		ChatID:    m.Chat.ID,
		Language:  m.Sender.LanguageCode,
		CreatedAt: time.Now(),
	}

	if err := storageInstance.AddUser(user); err != nil {
		b.Send(m.Sender, "Ошибка при добавлении пользователя:", err)
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

	if len(categories) == 0 {
		b.Send(m.Sender, "Категории отсутствуют.")
		return
	}

	markup := &telebot.ReplyMarkup{}
	var rows []telebot.Row
	for _, category := range categories {
		btn := markup.Data(category.Name, "cat:"+strconv.Itoa(int(category.ID)))
		rows = append(rows, markup.Row(btn))
	}
	markup.Inline(rows...)
	b.Send(m.Sender, "Выберите категорию:", markup)
}

func handleIncomeExpenseButtons(b *telebot.Bot, m *telebot.Message) {
	markup := &telebot.ReplyMarkup{}
	btnIncome := markup.Data("Доход", "income:"+m.Text)
	btnExpense := markup.Data("Расход", "expense:"+m.Text)
	markup.Inline(markup.Row(btnIncome, btnExpense))
	b.Send(m.Sender, "Выберите тип транзакции:", markup)
}

func handleTransactionCategories(b *telebot.Bot, c *telebot.Callback, storageInstance *storage.Storage) {
	categories, err := storageInstance.GetCategoriesByChatID(c.Sender.ID)
	if err != nil {
		b.Send(c.Sender, "Ошибка при получении категорий.")
		return
	}

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
	b.Edit(c.Message, "Выберите категорию:", markup)
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

func handleStatsButtons(b *telebot.Bot, m *telebot.Message) {
	markup := &telebot.ReplyMarkup{}
	btnToday := markup.Data("Сегодня", "today")
	markup.Inline(markup.Row(btnToday))
	b.Send(m.Sender, "Выберите тип транзакции:", markup)
}

func handleStats(b *telebot.Bot, sender *telebot.User, storageInstance *storage.Storage, startDate, endDate time.Time) {
	incomeCategories, expenseCategories, err := storageInstance.GetTransactionsStatsByCategory(sender.ID, startDate, endDate)
	if err != nil {
		b.Send(sender, "Ошибка при получении статистики: "+err.Error())
		return
	}

	totalIncome := sumMapValues(incomeCategories)
	totalExpense := sumMapValues(expenseCategories)
	netIncome := totalIncome - totalExpense

	var response strings.Builder
	response.WriteString("📊 *Статистика за период*\n\n")

	response.WriteString(fmt.Sprintf("💰 *Доход*: %.1f\n", totalIncome))
	for category, amount := range incomeCategories {
		response.WriteString(fmt.Sprintf("  - %s: %.1f\n", category, amount))
	}

	response.WriteString(fmt.Sprintf("\n💸 *Расход*: %.1f\n", totalExpense))
	for category, amount := range expenseCategories {
		response.WriteString(fmt.Sprintf("  - %s: %.1f\n", category, amount))
	}

	response.WriteString(fmt.Sprintf("\n💹 *Итого*: %.1f", netIncome))

	b.Send(sender, response.String(), &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
}
