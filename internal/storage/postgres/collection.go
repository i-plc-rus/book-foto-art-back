package postgres

import (
	"book-foto-art-back/internal/model"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CollectionStorage struct {
	DB *pgxpool.Pool
}

func (s *Storage) CreateCollection(ctx context.Context, col model.Collection) (*model.Collection, error) {
	row := s.DB.QueryRow(ctx,
		`INSERT INTO collections (name, date, user_id, session_id)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id`,
		col.Name, col.Date, col.UserID, col.SessionID,
	)
	var id int64
	if err := row.Scan(&id); err != nil {
		return nil, err
	}
	col.ID = id
	return &col, nil
}

func (s *Storage) GetCollectionByID(ctx context.Context, id int64) (*model.Collection, error) {
	row := s.DB.QueryRow(ctx,
		`SELECT id, name, date, user_id, session_id
		 FROM collections
		 WHERE id = $1`, id,
	)
	var col model.Collection
	if err := row.Scan(&col.ID, &col.Name, &col.Date, &col.UserID, &col.SessionID); err != nil {
		return nil, err
	}
	return &col, nil
}

func (s *Storage) DeleteCollection(ctx context.Context, id int64) error {
	_, err := s.DB.Exec(ctx, "DELETE FROM collections WHERE id = $1", id)
	return err
}
