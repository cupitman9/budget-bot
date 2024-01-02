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
	query := `INSERT INTO users (username, chat_id, language, created_at) VALUES ($1, $2, $3, $4)`
	_, err := s.db.Exec(query, user.Username, user.ChatID, user.Language, user.CreatedAt)
	return err
}
