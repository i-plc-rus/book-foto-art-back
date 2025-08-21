package model

import "github.com/google/uuid"

type User struct {
	ID           uuid.UUID `db:"id"`
	UserName     string    `db:"username"`
	Email        string    `db:"email"`
	Password     string    `db:"password"`
	RefreshToken string    `db:"refresh_token"`
	ResetToken   string    `db:"reset_token"`
	// OAuthProvider string    `db:"oauth_provider"`
	// OAuthID       string    `db:"oauth_id"`
}
