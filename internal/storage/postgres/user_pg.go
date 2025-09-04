package postgres

import (
	"book-foto-art-back/internal/model"
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserStorage struct {
	DB *pgxpool.Pool
}

func (s *Storage) CreateUser(ctx context.Context, user model.User) error {
	_, err := s.DB.Exec(ctx,
		`INSERT INTO users (username, email, password, refresh_token, subscription_active, subscription_expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		user.UserName, user.Email, user.Password, user.RefreshToken, user.SubscriptionActive, user.SubscriptionExpiresAt)
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

func (s *Storage) ExtendUserSubscription(ctx context.Context, yookassaPaymentID string, active bool) error {
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

	res, err := s.DB.Exec(ctx,
		`UPDATE users
		 SET subscription_active = $1,
		     subscription_expires_at = COALESCE(
		        GREATEST(subscription_expires_at, NOW()) + $2,
				NOW() + $2
			 )
		 WHERE id = $3`,
		active, addDays, userID)
	resAffected := res.RowsAffected()
	if resAffected == 0 {
		return sql.ErrNoRows
	}
	return err
}

func (s *Storage) GetUserSubscriptionInfo(ctx context.Context, userID uuid.UUID) (bool, *time.Time, uint, error) {
	var (
		isActive  bool
		expiresAt *time.Time
		daysLeft  uint
	)
	query := `
		SELECT subscription_active, subscription_expires_at
		FROM users
		WHERE id = $1
	`
	err := s.DB.QueryRow(ctx, query, userID).Scan(&isActive, &expiresAt)
	if err != nil {
		return false, nil, 0, err
	}
	if expiresAt != nil {
		isActive = time.Now().Before(*expiresAt)
		if isActive {
			daysLeft = uint(time.Until(*expiresAt).Hours() / 24)
		} else {
			_ = s.UpdateUserSubscriptionStatus(ctx, userID, false)
		}
	}
	return isActive, expiresAt, daysLeft, nil
}

func (s *Storage) UpdateUserSubscriptionStatus(ctx context.Context, userID uuid.UUID, active bool) error {
	res, err := s.DB.Exec(ctx,
		`UPDATE users
		 SET subscription_active = $1
		 WHERE id = $2`,
		active, userID)
	if err != nil {
		return err
	}
	rowsAffected := res.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return err
}
