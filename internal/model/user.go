package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                    uuid.UUID  `db:"id"`
	UserName              string     `db:"username"`
	Email                 string     `db:"email"`
	Password              string     `db:"password"`
	RefreshToken          string     `db:"refresh_token"`
	ResetToken            string     `db:"reset_token"`
	SubscriptionActive    bool       `db:"subscription_active"`
	SubscriptionExpiresAt *time.Time `db:"subscription_expires_at"`
}
