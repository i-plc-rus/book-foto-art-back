package postgres

import (
	"book-foto-art-back/internal/model"
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserStorage struct {
	DB *pgxpool.Pool
}

func (s *Storage) CreateUser(ctx context.Context, user model.User) error {
	_, err := s.DB.Exec(ctx,
		`INSERT INTO users (username, email, password, refresh_token)
		 VALUES ($1, $2, $3, $4)`,
		user.UserName, user.Email, user.Password, user.RefreshToken)
	return err
}

func (s *Storage) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	row := s.DB.QueryRow(ctx,
		`SELECT id, username, email, password, reset_token FROM users
		 WHERE email=$1`,
		email)

	var u model.User
	err := row.Scan(&u.ID, &u.UserName, &u.Email, &u.Password, &u.ResetToken)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *Storage) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	row := s.DB.QueryRow(ctx,
		`SELECT id, username, email, password, reset_token, subscription_active, subscription_expires_at FROM users
		 WHERE id=$1`,
		id)

	var u model.User
	err := row.Scan(&u.ID, &u.UserName, &u.Email, &u.Password, &u.ResetToken, &u.SubscriptionActive, &u.SubscriptionExpiresAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *Storage) UpdateResetToken(ctx context.Context, id uuid.UUID, resetToken string) error {
	_, err := s.DB.Exec(ctx,
		`UPDATE users
		 SET reset_token=$1
		 WHERE id=$2`,
		resetToken, id)
	return err
}

func (s *Storage) ResetPassword(ctx context.Context, id uuid.UUID, password string) error {
	_, err := s.DB.Exec(ctx,
		`UPDATE users
		 SET password=$1
		 WHERE id=$2`,
		password, id)
	return err
}

func (s *Storage) GetUserByRefresh(ctx context.Context, refreshToken string) (*model.User, error) {
	row := s.DB.QueryRow(ctx,
		`SELECT id, username, email, password, refresh_token
		 FROM users
		 WHERE refresh_token=$1`,
		refreshToken)

	var u model.User
	err := row.Scan(&u.ID, &u.UserName, &u.Email, &u.Password, &u.RefreshToken)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *Storage) UpdateRefreshToken(ctx context.Context, id uuid.UUID, refreshToken string) error {
	_, err := s.DB.Exec(ctx,
		`UPDATE users
		 SET refresh_token=$1
		 WHERE id=$2`,
		refreshToken, id)
	return err
}

func (s *Storage) UpdateUserSubscription(ctx context.Context, yookassaPaymentID string, active bool) error {
	var (
		userID  uuid.UUID
		plan    string
		addDays time.Duration
	)

	query := `
		SELECT user_id, plan
		FROM payments
		WHERE yookassa_payment_id = $1
	`
	err := s.DB.QueryRow(ctx, query, yookassaPaymentID).Scan(&userID, &plan)
	if err != nil {
		return err
	}

	switch plan {
	case "month":
		addDays = time.Hour * 24 * 30
	case "year":
		addDays = time.Hour * 24 * 365
	}

	_, err = s.DB.Exec(ctx,
		`UPDATE users
		 SET subscription_active = $1, subscription_expires_at = subscription_expires_at + $2
		 WHERE id = $3`,
		active, addDays, userID)
	return err
}

// func (s *Storage) GetUserSubscriptionStatus(ctx context.Context, userID uuid.UUID) (*model.SubscriptionStatus, error) {
// 	query := `
// 		SELECT subscription_active, subscription_expires_at
// 		FROM users
// 		WHERE id = $1`

// 	var active bool
// 	var expiresAt *time.Time
// 	err := s.DB.QueryRow(ctx, query, userID).Scan(&active, &expiresAt)
// 	if err != nil {
// 		return nil, err
// 	}

// 	status := &model.SubscriptionStatus{
// 		Active:    active,
// 		ExpiresAt: time.Time{},
// 		DaysLeft:  0,
// 		IsExpired: true,
// 	}

// 	if expiresAt != nil {
// 		status.ExpiresAt = *expiresAt
// 		status.IsExpired = time.Now().After(*expiresAt)
// 		if !status.IsExpired {
// 			status.DaysLeft = int(expiresAt.Sub(time.Now()).Hours() / 24)
// 		}
// 	}

// 	status.Active = active && !status.IsExpired

// 	return status, nil
// }
