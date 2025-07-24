package postgres

import (
	"book-foto-art-back/internal/model"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UploadStorage struct {
	DB *pgxpool.Pool
}

func (s *Storage) SaveUpload(ctx context.Context, upload model.UploadedPhoto) (*model.UploadedPhoto, error) {
	row := s.DB.QueryRow(ctx,
		`INSERT INTO uploads (collection_id, original_url, thumbnail_url, file_name, file_ext)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id`,
		upload.CollectionID, upload.OriginalURL, upload.ThumbnailURL, upload.FileName, upload.FileExt,
	)
	var id int64
	if err := row.Scan(&id); err != nil {
		return nil, err
	}
	upload.ID = id
	return &upload, nil
}

func (s *Storage) GetUploadsByCollection(ctx context.Context, collectionID int64) ([]model.UploadedPhoto, error) {
	rows, err := s.DB.Query(ctx,
		`SELECT id, collection_id, original_url, thumbnail_url, file_name, file_ext
		 FROM uploads
		 WHERE collection_id = $1`, collectionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.UploadedPhoto
	for rows.Next() {
		var f model.UploadedPhoto
		if err := rows.Scan(&f.ID, &f.CollectionID, &f.OriginalURL, &f.ThumbnailURL, &f.FileName, &f.FileExt); err != nil {
			return nil, err
		}
		result = append(result, f)
	}
	return result, nil
}
