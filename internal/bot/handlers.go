package bot

import (
	"fmt"
	"gopkg.in/tucnak/telebot.v2"
	"strconv"
	"strings"
	"telegram-budget-bot/internal/logger"
	"telegram-budget-bot/internal/model"
	"telegram-budget-bot/internal/storage"
	"time"
)

var log = logger.GetLogger()

func handleStart(b *telebot.Bot, m *telebot.Message, storageInstance *storage.Storage) {
	user := model.User{
		Username:  m.Sender.Username,
		ChatID:    m.Chat.ID,
		Language:  m.Sender.LanguageCode,
		CreatedAt: time.Now(),
	}

	if err := storageInstance.AddUser(user); err != nil {
		log.Errorf("Error adding user: %v", err)
		b.Send(m.Sender, "Ошибка при добавлении пользователя:", err)
		return
	}

	log.Infof("User %s added successfully", user.Username)

	welcomeText := "Привет! Нажмите /help для подробной информации"
	b.Send(m.Sender, welcomeText)
}

func handleShowCategories(b *telebot.Bot, m *telebot.Message, storageInstance *storage.Storage) {
	log.Info("Retrieving categories")

	categories, err := storageInstance.GetCategoriesByChatID(m.Chat.ID)
	if err != nil {
		log.Errorf("Error retrieving categories: %v", err)
		b.Send(m.Sender, fmt.Sprintf("Ошибка при получении категорий: %v", err))
		return
	}

	if len(categories) == 0 {
		log.Info("No categories found")
		b.Send(m.Sender, "Категории отсутствуют.")
		return
	}

	log.Infof("Found %d categories", len(categories))

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
	b.Send(m.Sender, "Категории:", markup)
}

func handleIncomeExpenseButtons(b *telebot.Bot, m *telebot.Message) {
	log.Info("Handling income and expense buttons")

	markup := &telebot.ReplyMarkup{}
	btnIncome := markup.Data("Доход", "income:"+m.Text)
	btnExpense := markup.Data("Расход", "expense:"+m.Text)
	markup.Inline(markup.Row(btnIncome, btnExpense))
	b.Send(m.Sender, "Выберите тип транзакции:", markup)
}

func handleTransactionCategories(b *telebot.Bot, c *telebot.Callback, storageInstance *storage.Storage) {
	log.Info("Retrieving transaction categories")

	categories, err := storageInstance.GetCategoriesByChatID(c.Sender.ID)
	if err != nil {
		log.Errorf("Error retrieving categories: %v", err)
		b.Send(c.Sender, "Ошибка при получении категорий.")
		return
	}

	if len(categories) == 0 {
		log.Info("No categories found")
		b.Send(c.Sender, "Категории отсутствуют.")
		return
	}

	log.Infof("Found %d categories", len(categories))

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
	log.Infof("Handling transaction: SenderID=%d, CategoryID=%d, Amount=%.2f, Type=%s", senderId, categoryId, amount, categoryType)

	transaction := model.Transaction{
		UserChat:        senderId,
		CategoryID:      categoryId,
		Amount:          amount,
		TransactionType: categoryType,
		CreatedAt:       time.Now(),
	}

	if err := storageInstance.AddTransaction(transaction); err != nil {
		log.Errorf("Error adding transaction: %v", err)
		return err
	}

	log.Info("Transaction added successfully")
	return nil
}

func handleStatsButtons(b *telebot.Bot, m *telebot.Message) {
	log.Info("Handling stats buttons")

	markup := &telebot.ReplyMarkup{}
	btnToday := markup.Data("Сегодня", "today")
	btnPeriod := markup.Data("Период", "period")
	markup.Inline(markup.Row(btnToday, btnPeriod))
	b.Send(m.Sender, "Выберите период:", markup)
}

func handleStats(b *telebot.Bot, sender *telebot.User, storageInstance *storage.Storage, startDate, endDate time.Time) {
	log.Infof("Handling stats: SenderID=%d, StartDate=%s, EndDate=%s", sender.ID, startDate, endDate)
	incomeCategories, expenseCategories, err := storageInstance.GetTransactionsStatsByCategory(sender.ID, startDate, endDate)
	if err != nil {
		log.Errorf("Error retrieving statistics: %v", err)
		b.Send(sender, "Ошибка при получении статистики: "+err.Error())
		return
	}

	totalIncome := sumMapValues(incomeCategories)
	totalExpense := sumMapValues(expenseCategories)
	netIncome := totalIncome - totalExpense
	log.Infof("Statistics retrieved: TotalIncome=%.1f, TotalExpense=%.1f, NetIncome=%.1f", totalIncome, totalExpense, netIncome)

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
