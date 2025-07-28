package postgres

import (
	"book-foto-art-back/internal/model"
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserStorage struct {
	DB *pgxpool.Pool
}

func (s *Storage) CreateUser(ctx context.Context, user model.User) error {
	_, err := s.DB.Exec(ctx, "INSERT INTO users (username, email, password) VALUES ($1, $2, $3)",
		user.UserName, user.Email, user.Password)
	return err
}

func (s *Storage) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	row := s.DB.QueryRow(ctx, "SELECT id, username, email, password FROM users WHERE email=$1", email)

	var u model.User
	err := row.Scan(&u.ID, &u.UserName, &u.Email, &u.Password)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *Storage) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	row := s.DB.QueryRow(ctx, "SELECT id, username, email, password FROM users WHERE id=$1", id)

	var u model.User
	err := row.Scan(&u.ID, &u.Email, &u.Password)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *Storage) UpdateRefreshToken(ctx context.Context, id uuid.UUID, token string) error {
	_, err := s.DB.Exec(ctx, "UPDATE users SET refresh_token=$1 WHERE id=$2", token, id)
	return err
}

func (s *Storage) GetUserByRefresh(ctx context.Context, refreshToken string) (*model.User, error) {
	row := s.DB.QueryRow(ctx, "SELECT id, username, email, password, refresh_token FROM users WHERE refresh_token=$1", refreshToken)

	var u model.User
	err := row.Scan(&u.ID, &u.UserName, &u.Email, &u.Password, &u.RefreshToken)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
