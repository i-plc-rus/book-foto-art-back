package model

import (
	"time"

	"github.com/google/uuid"
)

type Payment struct {
	ID                uuid.UUID `db:"id"`
	UserID            uuid.UUID `db:"user_id"`
	YooKassaPaymentID string    `db:"yookassa_payment_id"`
	Plan              string    `db:"plan"`
	Amount            float64   `db:"amount"`
	Status            string    `db:"status"`
	CreatedAt         time.Time `db:"created_at"`
	UpdatedAt         time.Time `db:"updated_at"`
}
