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

// TODO: Проверить поля базы данных, запросы
func (s *Storage) AddUser(user model.User) error {
	query := `INSERT INTO users (username, chat_id, language, created_at) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING`
	_, err := s.pool.Exec(context.Background(), query, user.Username, user.ChatID, user.Language, user.CreatedAt)
	return err
}

func (s *Storage) AddCategory(category model.Category) error {
	query := `INSERT INTO categories (name, chat_id) VALUES ($1, $2)`
	_, err := s.pool.Exec(context.Background(), query, category.Name, category.ChatID)
	return err
}

func (s *Storage) RenameCategory(categoryId int64, newName string) error {
	query := `UPDATE categories SET name = $1 WHERE id = $2`
	_, err := s.pool.Exec(context.Background(), query, newName, categoryId)
	if err != nil {
		return fmt.Errorf("не удалось переименовать категорию: %v", err)
	}
	return nil
}

func (s *Storage) DeleteCategory(chatID, categoryId int64) error {
	tx, err := s.pool.Begin(context.Background())
	if err != nil {
		return err
	}

	_, err = tx.Exec(context.Background(), `UPDATE transactions SET is_deleted = TRUE WHERE category_id = $1 AND user_chat = $2`, categoryId, chatID)
	if err != nil {
		tx.Rollback(context.Background())
		return err
	}

	_, err = tx.Exec(context.Background(), `UPDATE categories SET is_deleted = TRUE WHERE id = $1 AND chat_id = $2`, categoryId, chatID)
	if err != nil {
		tx.Rollback(context.Background())
		return err
	}

	return tx.Commit(context.Background())
}

func (s *Storage) GetCategoriesByChatID(chatID int64) ([]model.Category, error) {
	query := `SELECT id, name FROM categories WHERE chat_id = $1 AND is_deleted = FALSE`
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

	return categories, nil
}

func (s *Storage) AddTransaction(transaction model.Transaction) error {
	query := `INSERT INTO transactions (user_chat, category_id, amount, transaction_type, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := s.pool.Exec(context.Background(), query, transaction.UserChat, transaction.CategoryID, transaction.Amount, transaction.TransactionType, transaction.CreatedAt)
	return err
}

func (s *Storage) GetTransactionsStatsByCategory(chatID int64, startDate, endDate time.Time) (map[string]float64, map[string]float64, error) {
	incomeCategories := make(map[string]float64)
	expenseCategories := make(map[string]float64)
	query := `SELECT c.name, t.transaction_type, SUM(t.amount)
              FROM transactions t
              JOIN categories c ON t.category_id = c.id
              WHERE t.user_chat = $1 
                AND t.created_at >= $2 
                AND t.created_at < $3
                AND c.is_deleted = FALSE
              GROUP BY c.name, t.transaction_type`

	rows, err := s.pool.Query(context.Background(), query, chatID, startDate, endDate)
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
