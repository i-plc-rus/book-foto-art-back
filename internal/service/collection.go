package service

import (
	"book-foto-art-back/internal/model"
	"book-foto-art-back/internal/shared"
	"book-foto-art-back/internal/storage/postgres"
	"context"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
)

type CollectionService struct {
	Storage *postgres.Storage
}

func NewCollectionService(s *postgres.Storage) *CollectionService {
	return &CollectionService{Storage: s}
}

func (s *CollectionService) CreateCollection(ctx context.Context, userID uuid.UUID, name string, date time.Time) (
	*model.Collection, error) {
	defaultCoverURL := os.Getenv("DEFAULT_COLLECTION_COVER_URL")
	defaultCoverThumbnailURL := os.Getenv("DEFAULT_COLLECTION_COVER_THUMB_URL")

	collection := model.Collection{
		UserID:            userID,
		Name:              name,
		Date:              date,
		CoverURL:          defaultCoverURL,
		CoverThumbnailURL: defaultCoverThumbnailURL,
	}
	return s.Storage.CreateCollection(ctx, collection)
}

func (s *CollectionService) GetCollectionInfo(ctx context.Context, userID uuid.UUID, collectionID uuid.UUID) (
	*model.Collection, error) {
	return s.Storage.GetCollectionInfo(ctx, userID, collectionID)
}

func (s *CollectionService) GetCollections(ctx context.Context, userID uuid.UUID, searchParam string) (
	[]model.Collection, error) {
	return s.Storage.GetCollections(ctx, userID, searchParam)
}

func (s *CollectionService) DeleteCollection(ctx context.Context, userID uuid.UUID, collectionID uuid.UUID) error {
	return s.Storage.DeleteCollection(ctx, userID, collectionID)
}

func (s *CollectionService) GetCollectionPhotos(ctx context.Context, userID uuid.UUID, collectionID uuid.UUID, sortParam string) (
	[]model.UploadedPhoto, string, error) {
	// Выбираем параметр сортировки
	sort := shared.SortOption(sortParam)
	if _, ok := shared.ValidSorts[sort]; !ok {
		sort = shared.DefaultSort
	}
	// Получаем содержимое коллекции из БД
	photos, err := s.Storage.GetCollectionPhotos(ctx, userID, collectionID, sort)
	if err != nil {
		log.Printf("Storage ERROR: %v\n", err)
		return []model.UploadedPhoto{}, "", err
	}
	return photos, string(sort), err
}

func (s *CollectionService) DeletePhoto(ctx context.Context, userID uuid.UUID, photoID uuid.UUID) error {
	return s.Storage.DeletePhoto(ctx, userID, photoID)
}

func (s *CollectionService) UpdateCollectionCover(
	ctx context.Context, userID uuid.UUID, collectionID uuid.UUID, uploadedPhotoID uuid.UUID) error {
	return s.Storage.UpdateCollectionCover(ctx, userID, collectionID, uploadedPhotoID)
}
