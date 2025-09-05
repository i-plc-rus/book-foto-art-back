package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"

	"book-foto-art-back/internal/model"
	"book-foto-art-back/internal/shared"
	"book-foto-art-back/internal/storage/postgres"
	"book-foto-art-back/internal/storage/s3"
)

type CollectionService struct {
	Postgres *postgres.Storage
	S3       *s3.S3Storage
}

func NewCollectionService(pg *postgres.Storage, s3 *s3.S3Storage) *CollectionService {
	return &CollectionService{
		Postgres: pg,
		S3:       s3,
	}
}

func (s *CollectionService) CreateCollection(ctx context.Context, userID uuid.UUID, name string, date time.Time) (
	*model.Collection, error) {
	defaultCoverURL := fmt.Sprintf("%s/%s/default_collection_cover/default_cover.jpg", os.Getenv("AWS_ENDPOINT"), os.Getenv("AWS_BUCKET"))
	defaultCoverThumbnailURL := fmt.Sprintf("%s/%s/default_collection_cover/default_cover_thumb.jpg", os.Getenv("AWS_ENDPOINT"), os.Getenv("AWS_BUCKET"))

	collection := model.Collection{
		UserID:            userID,
		Name:              name,
		Date:              date,
		CoverURL:          defaultCoverURL,
		CoverThumbnailURL: defaultCoverThumbnailURL,
	}
	return s.Postgres.CreateCollection(ctx, collection)
}

func (s *CollectionService) GetCollectionInfo(ctx context.Context, userID uuid.UUID, collectionID uuid.UUID) (
	*model.Collection, error) {
	return s.Postgres.GetCollectionInfo(ctx, userID, collectionID)
}

func (s *CollectionService) GetCollections(ctx context.Context, userID uuid.UUID, searchParam string) (
	[]model.Collection, error) {
	return s.Postgres.GetCollections(ctx, userID, searchParam)
}

func (s *CollectionService) DeleteCollection(ctx context.Context, userID uuid.UUID, collectionID uuid.UUID) error {
	err := s.S3.DeleteCollection(ctx, collectionID)
	if err != nil {
		return err
	}
	err = s.Postgres.DeleteCollection(ctx, userID, collectionID)
	if err != nil {
		return err
	}
	return nil
}

func (s *CollectionService) GetCollectionPhotos(
	ctx context.Context, userID uuid.UUID, collectionID uuid.UUID, sortParam string, favoritesOnly bool) (
	[]model.UploadedPhoto, string, error) {
	// Выбираем параметр сортировки
	sort := shared.SortOption(sortParam)
	if _, ok := shared.ValidSorts[sort]; !ok {
		sort = shared.DefaultSort
	}
	// Получаем содержимое коллекции из БД
	photos, err := s.Postgres.GetCollectionPhotos(ctx, userID, collectionID, sort, favoritesOnly)
	if err != nil {
		log.Printf("Storage ERROR: %v\n", err)
		return []model.UploadedPhoto{}, "", err
	}
	return photos, string(sort), err
}

func (s *CollectionService) PublishCollection(ctx context.Context, userID uuid.UUID, collectionID uuid.UUID) (string, error) {
	// Генерируем уникальный токен
	bytes := make([]byte, 6)
	rand.Read(bytes)
	token := base64.URLEncoding.EncodeToString(bytes)
	token = strings.ReplaceAll(token, "+", "-")
	token = strings.ReplaceAll(token, "/", "_")

	// Генерируем короткую ссылку на коллекцию
	link := fmt.Sprintf("%s/s/%s", os.Getenv("FRONTEND_URL"), token)

	// Публикуем коллекцию
	link, err := s.Postgres.PublishCollection(ctx, userID, collectionID, token, link)
	if err != nil {
		return "", err
	}
	return link, nil
}

func (s *CollectionService) GetPublicCollectionLink(ctx context.Context, token string) (string, error) {
	err := s.Postgres.UpdateShortLink(ctx, token)
	if err != nil {
		return "", err
	}
	link := fmt.Sprintf("/public/collection/%s/photos", token)
	return link, nil
}

func (s *CollectionService) GetPublicCollection(
	ctx context.Context, token string, sortParam string, favoritesOnly bool) (
	*model.Collection, []model.UploadedPhoto, string, error) {
	// Получаем информацию о публичной коллекции
	collection, err := s.Postgres.GetPublicCollectionInfo(ctx, token)
	if err != nil {
		log.Printf("Storage ERROR: %v\n", err)
		return nil, []model.UploadedPhoto{}, "", err
	}

	// Выбираем параметр сортировки
	sort := shared.SortOption(sortParam)
	if _, ok := shared.ValidSorts[sort]; !ok {
		sort = shared.DefaultSort
	}
	// Получаем фотографии коллекции из БД
	photos, err := s.Postgres.GetPublicCollectionPhotos(ctx, token, sort, favoritesOnly)
	if err != nil {
		log.Printf("Storage ERROR: %v\n", err)
		return nil, []model.UploadedPhoto{}, "", err
	}
	return collection, photos, string(sort), err
}

func (s *CollectionService) UnpublishCollection(ctx context.Context, userID uuid.UUID, collectionID uuid.UUID) error {
	err := s.Postgres.UnpublishCollection(ctx, userID, collectionID)
	if err != nil {
		return err
	}
	return nil
}

func (s *CollectionService) GetShortLinkInfo(ctx context.Context, token string) (model.ShortLink, error) {
	shortLink, err := s.Postgres.GetShortLinkInfo(ctx, token)
	if err != nil {
		return model.ShortLink{}, err
	}
	return shortLink, nil
}

func (s *CollectionService) DeletePhoto(ctx context.Context, userID uuid.UUID, photoID uuid.UUID) error {
	photo, err := s.Postgres.GetCollectionPhoto(ctx, userID, photoID)
	if err != nil {
		return err
	}
	err = s.S3.DeletePhoto(ctx, photo.CollectionID, photo.FileName)
	if err != nil {
		return err
	}
	err = s.Postgres.DeletePhoto(ctx, userID, photoID)
	if err != nil {
		return err
	}
	return nil
}

func (s *CollectionService) UpdateCollectionCover(
	ctx context.Context, userID uuid.UUID, collectionID uuid.UUID, uploadedPhotoID uuid.UUID) error {
	return s.Postgres.UpdateCollectionCover(ctx, userID, collectionID, uploadedPhotoID)
}

func (s *CollectionService) MarkPhoto(ctx context.Context, photoID uuid.UUID, action string) error {
	var isFavorite bool
	switch action {
	case "favorite":
		isFavorite = true
	case "unfavorite":
		isFavorite = false
	default:
		return errors.New("invalid action")
	}
	err := s.Postgres.MarkPhoto(ctx, photoID, isFavorite)
	if err != nil {
		return err
	}
	return nil
}
