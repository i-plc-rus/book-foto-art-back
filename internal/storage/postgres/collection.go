package postgres

import (
	"book-foto-art-back/internal/model"
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CollectionStorage struct {
	DB *pgxpool.Pool
}

func (s *Storage) CreateCollection(ctx context.Context, col model.Collection) (*model.Collection, error) {
	row := s.DB.QueryRow(ctx,
		`INSERT INTO collections (name, date, user_id)
		 VALUES ($1, $2, $3)
		 RETURNING id`,
		col.Name, col.Date, col.UserID,
	)
	var id uuid.UUID
	if err := row.Scan(&id); err != nil {
		return nil, err
	}
	col.ID = id
	return &col, nil
}

func (s *Storage) GetCollectionByID(ctx context.Context, id uuid.UUID) (*model.Collection, error) {
	row := s.DB.QueryRow(ctx,
		`SELECT id, name, date, user_id
		 FROM collections
		 WHERE id = $1`, id,
	)
	var col model.Collection
	if err := row.Scan(&col.ID, &col.Name, &col.Date, &col.UserID); err != nil {
		return nil, err
	}
	return &col, nil
}

func (s *Storage) GetCollections(ctx context.Context, userID uuid.UUID) ([]model.Collection, error) {
	rows, err := s.DB.Query(ctx, `SELECT id, name, date FROM collections WHERE user_id = $1 ORDER BY date DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var collections []model.Collection
	for rows.Next() {
		var c model.Collection
		err := rows.Scan(&c.ID, &c.Name, &c.Date)
		if err != nil {
			return nil, err
		}
		c.UserID = userID
		collections = append(collections, c)
	}
	return collections, nil
}

func (s *Storage) DeleteCollection(ctx context.Context, id uuid.UUID) error {
	_, err := s.DB.Exec(ctx, "DELETE FROM collections WHERE id = $1", id)
	return err
}
