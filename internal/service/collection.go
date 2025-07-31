package service

import (
	"book-foto-art-back/internal/model"
	"book-foto-art-back/internal/shared"
	"book-foto-art-back/internal/storage/postgres"
	"context"
	"log"
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
	collection := model.Collection{
		UserID: userID,
		Name:   name,
		Date:   date,
	}
	return s.Storage.CreateCollection(ctx, collection)
}

func (s *CollectionService) GetCollectionInfo(ctx context.Context, userID uuid.UUID, collectionID uuid.UUID) (
	*model.Collection, error) {
	return s.Storage.GetCollectionInfo(ctx, userID, collectionID)
}

func (s *CollectionService) GetCollections(ctx context.Context, userID uuid.UUID) (
	[]model.Collection, error) {
	return s.Storage.GetCollections(ctx, userID)
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
