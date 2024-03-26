package model

import "time"

const (
	StateAwaitingNewCategoryName UserState = iota + 1
	StateAwaitingRenameCategory
	StateAwaitingPeriod
)

const (
	TransactionTypeIncome  uint8 = 1
	TransactionTypeExpense uint8 = 2
)

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
	ChatID          int64
	CategoryID      int64
	Amount          float64
	TransactionType uint8
	CreatedAt       time.Time
}

type UserState int

type UserSession struct {
	State             UserState
	CategoryID        int
	TransactionAmount float64
	StartDate         time.Time
	EndDate           time.Time
}

func (u *User) IsEmpty() bool {
	return u.ChatID == 0 && u.CreatedAt.IsZero()
}
