package model

import (
	"time"

	"github.com/google/uuid"
)

type Collection struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Name      string    `json:"name"`
	Date      time.Time `json:"date"`
	CreatedAt time.Time `json:"created_at"`
}
