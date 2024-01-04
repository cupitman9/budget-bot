package storage

import (
	"database/sql"
	_ "github.com/lib/pq"
	"telegram-budget-bot/internal/model"
	"time"
)

type Storage struct {
	db *sql.DB
}

func NewStorage(dataSourceName string) (*Storage, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &Storage{db: db}, nil
}

func (s *Storage) AddUser(user model.User) error {
	query := `INSERT INTO users (username, chat_id, language, created_at) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING`
	_, err := s.db.Exec(query, user.Username, user.ChatID, user.Language, user.CreatedAt)
	return err
}

func (s *Storage) AddCategory(category model.Category) error {
	query := `INSERT INTO categories (name, chat_id) VALUES ($1, $2)`
	_, err := s.db.Exec(query, category.Name, category.ChatID)
	return err
}

func (s *Storage) GetCategoriesByChatID(chatID int64) ([]model.Category, error) {
	query := `SELECT id, name FROM categories WHERE chat_id = $1`
	rows, err := s.db.Query(query, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []model.Category
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}

	return categories, nil
}

func (s *Storage) AddTransaction(transaction model.Transaction) error {
	query := `INSERT INTO transactions (user_chat, category_id, amount, description, transaction_type, created_at) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := s.db.Exec(query, transaction.UserChat, transaction.CategoryID, transaction.Amount, transaction.Description, transaction.TransactionType, transaction.CreatedAt)
	return err
}

func (s *Storage) GetTransactionsStatsByCategory(chatID int64, startDate, endDate time.Time) (map[string]float64, map[string]float64, error) {
	incomeCategories := make(map[string]float64)
	expenseCategories := make(map[string]float64)

	query := `SELECT c.name, t.transaction_type, SUM(t.amount)
              FROM transactions t
              JOIN categories c ON t.category_id = c.id
              WHERE t.user_chat = $1 AND t.created_at >= $2 AND t.created_at < $3
              GROUP BY c.name, t.transaction_type`

	rows, err := s.db.Query(query, chatID, startDate, endDate)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var categoryName, transactionType string
		var amount float64

		if err := rows.Scan(&categoryName, &transactionType, &amount); err != nil {
			return nil, nil, err
		}

		if transactionType == "income" {
			incomeCategories[categoryName] = amount
		} else if transactionType == "expense" {
			expenseCategories[categoryName] = amount
		}
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return incomeCategories, expenseCategories, nil
}
