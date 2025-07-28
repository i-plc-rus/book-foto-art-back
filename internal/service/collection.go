package service

import (
	"book-foto-art-back/internal/model"
	"book-foto-art-back/internal/storage/postgres"
	"context"
	"time"

	"github.com/google/uuid"
)

type CollectionService struct {
	Storage *postgres.Storage
}

func NewCollectionService(s *postgres.Storage) *CollectionService {
	return &CollectionService{Storage: s}
}

func (s *CollectionService) CreateCollection(ctx context.Context, userID uuid.UUID, name string, date time.Time) (*model.Collection, error) {
	collection := model.Collection{
		UserID: userID,
		Name:   name,
		Date:   date,
	}
	return s.Storage.CreateCollection(ctx, collection)
}

func (s *CollectionService) GetCollectionByID(ctx context.Context, collectionID uuid.UUID) (*model.Collection, error) {
	return s.Storage.GetCollectionByID(ctx, collectionID)
}

func (s *CollectionService) GetCollections(ctx context.Context, userID uuid.UUID) ([]model.Collection, error) {
	return s.Storage.GetCollections(ctx, userID)
}

func (s *CollectionService) DeleteCollection(ctx context.Context, userID uuid.UUID, collectionID uuid.UUID) error {
	return s.Storage.DeleteCollection(ctx, collectionID)
}
