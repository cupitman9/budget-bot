package storage

import (
	"database/sql"
	_ "github.com/lib/pq"
	"telegram-budget-bot/internal/model"
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
