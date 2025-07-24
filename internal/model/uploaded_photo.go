package model

import "time"

type UploadedPhoto struct {
	ID           int64  `json:"id"`
	CollectionID int64  `json:"collection_id"`
	UserID       int64  `json:"user_id"`
	Session      string `json:"session"`
	OriginalURL  string `json:"original_url"`
	ThumbnailURL string `json:"thumbnail_url"`
	FileExt      string `json:"file_ext"`
	HashName     string `json:"hash_name"`
	UploadedAt   time.Time
}
