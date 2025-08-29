package model

import (
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

// ErrorMessage представляет сообщение об ошибке
// @Description Структура для сообщений об ошибках API
type ErrorMessage struct {
	Error string `json:"error" example:"Invalid credentials"`
}

// YandexLoginResponse представляет ответ с URL для перенаправления на Яндекс OAuth
// @Description Структура ответа с URL для перенаправления на Яндекс OAuth
type YandexLoginResponse struct {
	URL string `json:"url" example:"https://oauth.yandex.ru/authorize?response_type=code&client_id=1234567890&redirect_uri=http://localhost:8080/auth/yandex/callback&state=1234567890"`
}

// YandexCallbackResponse представляет ответ с данными пользователя после успешной аутентификации
// @Description Структура ответа с данными пользователя после успешной аутентификации
type YandexCallbackResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"`
	User         struct {
		ID       string `json:"id" example:"06301788-e325-488f-94b5-1711e211b82a"`
		UserName string `json:"username" example:"user1"`
		Email    string `json:"email" example:"user1@example.com"`
	} `json:"user"`
}

// RefreshRequest содержит refresh токен для обновления access токена
// @Description Структура запроса для обновления токена доступа
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c" validate:"required"`
}

// RefreshResponse представляет ответ с обновленным access токеном
// @Description Структура ответа при успешном обновлении токена
type RefreshResponse struct {
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"`
}

// RegisterRequest содержит данные для регистрации нового пользователя
// @Description Структура запроса для регистрации пользователя в системе
type RegisterRequest struct {
	UserName string `json:"username" example:"user1"`
	Email    string `json:"email" example:"user1@example.com"`
	Password string `json:"password" example:"password123"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" example:"user1@example.com"`
}

type ResetPasswordRequest struct {
	NewPassword string `json:"new_password" example:"password123"`
}

// TokenResponse представляет ответ с токенами аутентификации
// @Description Структура ответа с access и refresh токенами
type TokenResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"`
}

// LoginRequest содержит данные для аутентификации пользователя
// @Description Структура запроса для входа в систему
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
	UserName          string `json:"username"`
	IsPublished       bool   `json:"is_published"`
	CountPhotos       uint   `json:"count_photos"`
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
	Photos []UploadedFile `json:"files"`
	Sort   string         `json:"sort" example:"uploaded_new"`
}

type UpdateCollectionCoverRequest struct {
	PhotoID string `json:"photo_id" example:"06301788-e325-488f-94b5-1711e211b82a"`
}

type PublishCollectionResponse struct {
	Link string `json:"link" example:"https://book-foto-art.ru/s/e325488f-94b5-1711e211b82a"`
}

type ShortLinkInfoResponse struct {
	ID                uuid.UUID `json:"id" example:"06301788-e325-488f-94b5-1711e211b82a"`
	CollectionID      uuid.UUID `json:"collection_id" example:"06301788-e325-488f-94b5-1711e211b82a"`
	URL               string    `json:"url" example:"https://book-foto-art.ru/s/e325488f-94b5-1711e211b82a"`
	Token             string    `json:"token" example:"e325488f-94b5-1711e211b82a"`
	CreatedAt         time.Time `json:"created_at" example:"2025-0715:12:00Z"`
	ClickCount        uint      `json:"click_count" example:"100"`
	Name              string    `json:"name" example:"My Collection"`
	UserName          string    `json:"username" example:"user1"`
	CoverURL          string    `json:"cover_url"`
	CoverThumbnailURL string    `json:"cover_thumbnail_url"`
}
