package postgres

import (
	"book-foto-art-back/internal/model"
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UploadStorage struct {
	DB *pgxpool.Pool
}

func (s *Storage) SaveUpload(ctx context.Context, upload *model.UploadedPhoto) (*model.UploadedPhoto, error) {
	row := s.DB.QueryRow(ctx,
		`INSERT INTO uploaded_photos
    	 (collection_id, user_id, original_url, thumbnail_url, file_name, file_ext, hash_name)
     	 VALUES ($1, $2, $3, $4, $5, $6, $7)
     	 RETURNING id`,
		upload.CollectionID, upload.UserID, upload.OriginalURL, upload.ThumbnailURL,
		upload.FileName, upload.FileExt, upload.HashName,
	)
	var id uuid.UUID
	if err := row.Scan(&id); err != nil {
		return nil, err
	}
	upload.ID = id
	return upload, nil
}
