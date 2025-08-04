package model

import (
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

type ErrorMessage struct {
	Error string `json:"error" example:"Invalid credentials"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

type RefreshResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

type RegisterRequest struct {
	UserName string `json:"username" example:"user1"`
	Email    string `json:"email" example:"user1@example.com"`
	Password string `json:"password" example:"password123"`
}

type TokenResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

type LoginRequest struct {
	Email    string `json:"email" example:"user1@example.com"`
	Password string `json:"password" example:"password123"`
}

type ProfileResponse struct {
	ID    string `json:"id" example:"06301788-e325-488f-94b5-1711e211b82a"`
	Email string `json:"email" example:"user1@example.com"`
}

type CreateCollectionRequest struct {
	Name string `json:"name" example:"My Collection"`
	Date string `json:"date" example:"2025-07-28T00:00:00Z"`
}

type CreateCollectionResponse struct {
	ID string `json:"id" example:"06301788-e325-488f-94b5-1711e211b82a"`
}

type CollectionInfoResponse struct {
	ID                string `json:"id" example:"06301788-e325-488f-94b5-1711e211b82a"`
	UserID            string `json:"user_id" example:"06301788-e325-488f-94b5-1711e211b82a"`
	Name              string `json:"name" example:"My Collection"`
	Date              string `json:"date" example:"2025-07-20T00:00:00Z"`
	CreatedAt         string `json:"created_at" example:"2025-0715:12:00Z"`
	CoverURL          string `json:"cover_url"`
	CoverThumbnailURL string `json:"cover_thumbnail_url"`
}

type CollectionsListResponse struct {
	Collections []CollectionInfoResponse `json:"collections"`
}

type BooleanResponse struct {
	Success bool `json:"success" example:"true"`
}

type UploadFilesRequest struct {
	CollectionID string                  `form:"collection_id" example:"06301788-e325-488f-94b5-1711e211b82a"`
	Files        []*multipart.FileHeader `form:"files"`
}

type UploadedFile struct {
	ID           uuid.UUID `json:"id" example:"06301788-e325-488f-94b5-1711e211b82a"`
	CollectionID uuid.UUID `json:"collection_id" example:"06301788-e325-488f-94b5-1711e211b82a"`
	UserID       uuid.UUID `json:"user_id" example:"06301788-e325-488f-94b5-1711e211b82a"`
	OriginalURL  string    `json:"original_url"`
	ThumbnailURL string    `json:"thumbnail_url"`
	FileName     string    `json:"file_name"`
	FileExt      string    `json:"file_ext"`
	HashName     string    `json:"hash_name"`
	UploadedAt   time.Time `json:"uploaded_at" example:"2025-0715:12:00Z"`
}

type UploadFilesResponse struct {
	UploadFiles []UploadedFile `json:"files"`
}

type CollectionPhotosResponse struct {
	Photos []UploadedPhoto `json:"files"`
	Sort   string          `json:"sort" example:"uploaded_new"`
}

type UpdateCollectionCoverRequest struct {
	PhotoID string `json:"photo_id" example:"06301788-e325-488f-94b5-1711e211b82a"`
}
