package model

import "time"

type Collection struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Name      string    `json:"name"`
	Date      time.Time `json:"date"`
	CreatedAt time.Time `json:"created_at"`
}
