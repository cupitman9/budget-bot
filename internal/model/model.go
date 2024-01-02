package model

import "time"

type User struct {
	Username  string
	ChatID    int64
	Language  string
	CreatedAt time.Time
}
