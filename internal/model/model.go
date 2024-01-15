package model

import "time"

type User struct {
	Username  string
	ChatID    int64
	Language  string
	CreatedAt time.Time
}

type Category struct {
	ID        int64
	Name      string
	ChatID    int64
	CreatedAt time.Time
}

type Transaction struct {
	UserChat        int64
	CategoryID      int64
	Amount          float64
	TransactionType string
	CreatedAt       time.Time
}

type UserState int

const (
	StateNone UserState = iota
	StateAwaitingNewCategoryName
	StateAwaitingRenameCategory
	StateAwaitingTransactionAmount UserState = iota
	StateAwaitingPeriod            UserState = iota
)

type UserSession struct {
	State             UserState
	CategoryID        int
	TransactionAmount float64
	StartDate         time.Time
	EndDate           time.Time
}
