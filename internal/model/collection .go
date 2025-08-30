package model

import (
	"time"

	"github.com/google/uuid"
)

type Collection struct {
	ID                uuid.UUID `json:"id"`
	UserID            uuid.UUID `json:"user_id"`
	Name              string    `json:"name"`
	Date              time.Time `json:"date"`
	CreatedAt         time.Time `json:"created_at"`
	CoverURL          string    `json:"cover_url"`
	CoverThumbnailURL string    `json:"cover_thumbnail_url"`
	UserName          string    `json:"username"`
	IsPublished       bool      `json:"is_published"`
	CountPhotos       uint      `json:"count_photos"`
}

type ShortLink struct {
	ID           uuid.UUID `json:"id"`
	CollectionID uuid.UUID `json:"collection_id"`
	URL          string    `json:"url"`
	Token        string    `json:"token"`
	CreatedAt    time.Time `json:"created_at"`
	ClickCount   uint      `json:"click_count"`
}
