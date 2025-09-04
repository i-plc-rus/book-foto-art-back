package postgres

import (
	"book-foto-art-back/internal/model"
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentStorage struct {
	DB *pgxpool.Pool
}

func (s *Storage) CreatePayment(ctx context.Context, payment *model.Payment) error {
	query := `
		INSERT INTO payments (id, user_id, yookassa_payment_id, amount, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	res, err := s.DB.Exec(
		ctx, query, payment.ID, payment.UserID, payment.YooKassaPaymentID,
		payment.Amount, payment.Status, payment.CreatedAt,
	)
	if err != nil {
		return err
	}
	rowsAffected := res.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Storage) GetPayment(ctx context.Context, userID uuid.UUID, paymentID string) (*model.Payment, error) {
	query := `
		SELECT id, user_id, yookassa_payment_id, amount, status, created_at, updated_at
		FROM payments
		WHERE id = $1 AND user_id = $2
	`
	var payment model.Payment
	err := s.DB.QueryRow(ctx, query, paymentID, userID).Scan(
		&payment.ID, &payment.UserID, &payment.YooKassaPaymentID, &payment.Amount,
		&payment.Status, &payment.CreatedAt, &payment.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (s *Storage) UpdatePaymentStatus(ctx context.Context, yookassaPaymentID, status string) error {
	query := `
		UPDATE payments
		SET status = $1, updated_at = $2
		WHERE yookassa_payment_id = $3`

	res, err := s.DB.Exec(ctx, query, status, time.Now(), yookassaPaymentID)
	if err != nil {
		return err
	}
	rowsAffected := res.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}
