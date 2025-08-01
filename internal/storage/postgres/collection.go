package postgres

import (
	"book-foto-art-back/internal/model"
	"book-foto-art-back/internal/shared"
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CollectionStorage struct {
	DB *pgxpool.Pool
}

func (s *Storage) CreateCollection(ctx context.Context, col model.Collection) (*model.Collection, error) {
	row := s.DB.QueryRow(ctx,
		`INSERT INTO collections (name, date, user_id, cover_url, cover_thumbnail_url)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id`,
		col.Name, col.Date, col.UserID, col.CoverURL, col.CoverThumbnailURL,
	)
	var id uuid.UUID
	if err := row.Scan(&id); err != nil {
		return nil, err
	}
	col.ID = id
	return &col, nil
}

func (s *Storage) GetCollectionInfo(ctx context.Context, userID uuid.UUID, collectionID uuid.UUID) (*model.Collection, error) {
	row := s.DB.QueryRow(ctx,
		`SELECT id, user_id, name, date, created_at, cover_url, cover_thumbnail_url
		 FROM collections
		 WHERE user_id = $1 AND id = $2`, userID, collectionID,
	)
	var col model.Collection
	if err := row.Scan(&col.ID, &col.UserID, &col.Name, &col.Date, &col.CreatedAt, &col.CoverURL, &col.CoverThumbnailURL); err != nil {
		return nil, err
	}
	return &col, nil
}

func (s *Storage) GetCollections(ctx context.Context, userID uuid.UUID) ([]model.Collection, error) {
	rows, err := s.DB.Query(ctx,
		`SELECT id, name, date, created_at, cover_url, cover_thumbnail_url
		  FROM collections
		  WHERE user_id = $1
		  ORDER BY date DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var collections []model.Collection
	for rows.Next() {
		var c model.Collection
		err := rows.Scan(&c.ID, &c.Name, &c.Date, &c.CreatedAt, &c.CoverURL, &c.CoverThumbnailURL)
		if err != nil {
			return nil, err
		}
		c.UserID = userID
		collections = append(collections, c)
	}
	return collections, nil
}

func (s *Storage) DeleteCollection(ctx context.Context, userID uuid.UUID, collectionID uuid.UUID) error {
	res, err := s.DB.Exec(ctx, "DELETE FROM collections WHERE user_id = $1 AND id = $2", userID, collectionID)
	if err != nil {
		return err
	}
	rowsAffected := res.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// func (s *Storage) UpdateCollectionCover(ctx context.Context, userID uuid.UUID, collectionID uuid.UUID, coverURL string, coverThumbnailURL string) error {
// 	res, err := s.DB.Exec(ctx,
// 		"UPDATE collections SET default_cover_url = $1, default_cover_thumbnail_url = $2 WHERE user_id = $3 AND id = $4",
// 		coverURL, coverThumbnailURL, userID, collectionID)
// 	if err != nil {
// 		return err
// 	}
// 	rowsAffected := res.RowsAffected()
// 	if rowsAffected == 0 {
// 		return sql.ErrNoRows
// 	}
// 	return nil
// }

func (s *Storage) GetCollectionPhotos(ctx context.Context, userID uuid.UUID, collectionID uuid.UUID, sort shared.SortOption) (
	[]model.UploadedPhoto, error) {

	// Определяем SQL для сортировки
	var orderBy string
	switch sort {
	case shared.SortUploadedNew:
		orderBy = " ORDER BY uploaded_at DESC"
	case shared.SortUploadedOld:
		orderBy = " ORDER BY uploaded_at ASC"
	case shared.SortNameAZ:
		orderBy = " ORDER BY file_name ASC"
	case shared.SortNameZA:
		orderBy = " ORDER BY file_name DESC"
	case shared.SortRandom:
		orderBy = " ORDER BY RANDOM()"
	default:
		orderBy = " ORDER BY uploaded_at DESC"
	}

	rows, err := s.DB.Query(ctx,
		`SELECT id, collection_id, user_id, original_url, thumbnail_url, file_name, file_ext, hash_name, uploaded_at
		 FROM uploaded_photos
		 WHERE user_id = $1 AND collection_id = $2`+orderBy, userID, collectionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.UploadedPhoto
	for rows.Next() {
		var f model.UploadedPhoto
		if err := rows.Scan(&f.ID, &f.CollectionID, &f.UserID, &f.OriginalURL, &f.ThumbnailURL,
			&f.FileName, &f.FileExt, &f.HashName, &f.UploadedAt); err != nil {
			return nil, err
		}
		result = append(result, f)
	}
	return result, nil
}
