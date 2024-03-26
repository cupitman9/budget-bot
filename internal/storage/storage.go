package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cupitman9/budget-bot/internal/model"
)

type Storage struct {
	pool *pgxpool.Pool
}

func NewStorage(ctx context.Context, postgresDsn string) (*Storage, error) {
	poolConfig, err := pgxpool.ParseConfig(postgresDsn)
	if err != nil {
		return nil, fmt.Errorf("error parsing config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("error connecting: %w", err)
	}

	if err = pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("error pinging pool: %w", err)
	}

	return &Storage{pool: pool}, nil
}

func (s *Storage) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}

func (s *Storage) AddUser(user model.User) error {
	query := `INSERT INTO users (chat_id, username, language) VALUES ($1, $2, $3)`
	_, err := s.pool.Exec(context.Background(), query, user.ChatID, user.Username, user.Language)
	return err
}

func (s *Storage) GetUserByChatID(chatID int64) (model.User, error) {
	query := `SELECT chat_id, username, language, created_at FROM users WHERE chat_id = $1`
	u := model.User{}
	err := s.pool.QueryRow(context.Background(), query, chatID).Scan(&u.ChatID, &u.Username, &u.Language, &u.CreatedAt)
	return u, err
}

func (s *Storage) AddCategory(category model.Category) error {
	query := `INSERT INTO categories (name, chat_id) VALUES ($1, $2)`
	_, err := s.pool.Exec(context.Background(), query, category.Name, category.ChatID)
	return err
}

func (s *Storage) RenameCategory(categoryId int64, newName string) error {
	query := `UPDATE categories SET name = $1 WHERE id = $2`
	_, err := s.pool.Exec(context.Background(), query, newName, categoryId)
	return err
}

func (s *Storage) GetCategoriesByChatID(chatID int64) ([]model.Category, error) {
	query := `SELECT id, name FROM categories WHERE chat_id = $1`
	rows, err := s.pool.Query(context.Background(), query, chatID)
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

	return categories, rows.Err()
}

func (s *Storage) AddTransaction(transaction model.Transaction) error {
	query := `INSERT INTO transactions (chat_id, category_id, amount, transaction_type) VALUES ($1, $2, $3, $4)`
	_, err := s.pool.Exec(
		context.Background(),
		query,
		transaction.ChatID,
		transaction.CategoryID,
		transaction.Amount,
		transaction.TransactionType,
	)
	return err
}

func (s *Storage) GetTransactionsStatsByCategory(chatID int64, startDate, endDate time.Time) (
	map[string]float64,
	map[string]float64,
	error,
) {
	incomeCategories := make(map[string]float64)
	expenseCategories := make(map[string]float64)
	query := `SELECT c.name, t.transaction_type, SUM(t.amount)
              FROM transactions t
              JOIN categories c ON t.category_id = c.id
              WHERE t.chat_id = $1 
                AND t.created_at >= $2 
                AND t.created_at < $3
              GROUP BY c.name, t.transaction_type`

	rows, err := s.pool.Query(context.Background(), query, chatID, startDate, endDate)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			categoryName    string
			transactionType uint8
			amount          float64
		)

		if err := rows.Scan(&categoryName, &transactionType, &amount); err != nil {
			return nil, nil, err
		}

		if transactionType == model.TransactionTypeIncome {
			incomeCategories[categoryName] = amount
		} else if transactionType == model.TransactionTypeExpense {
			expenseCategories[categoryName] = amount
		}
	}

	return incomeCategories, expenseCategories, rows.Err()
}
