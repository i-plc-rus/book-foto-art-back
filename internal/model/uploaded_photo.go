package model

import "time"

type UploadedPhoto struct {
	ID           int64  `json:"id"`
	CollectionID int64  `json:"collection_id"`
	UserID       int64  `json:"user_id"`
	OriginalURL  string `json:"original_url"`
	ThumbnailURL string `json:"thumbnail_url"`
	FileName     string `json:"file_name"`
	FileExt      string `json:"file_ext"`
	HashName     string `json:"hash_name"`
	UploadedAt   time.Time
}
