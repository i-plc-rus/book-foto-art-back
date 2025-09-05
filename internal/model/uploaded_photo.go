package model

import (
	"time"

	"github.com/google/uuid"
)

type UploadedPhoto struct {
	ID           uuid.UUID `json:"id"`
	CollectionID uuid.UUID `json:"collection_id"`
	UserID       uuid.UUID `json:"user_id"`
	OriginalURL  string    `json:"original_url"`
	ThumbnailURL string    `json:"thumbnail_url"`
	FileName     string    `json:"file_name"`
	FileExt      string    `json:"file_ext"`
	HashName     string    `json:"hash_name"`
	UploadedAt   time.Time `json:"uploaded_at"`
	IsFavorite   bool      `json:"is_favorite"`
}
