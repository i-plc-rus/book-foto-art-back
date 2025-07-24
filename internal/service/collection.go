package service

import (
	"book-foto-art-back/internal/model"
	"book-foto-art-back/internal/storage"
	"context"
)

type CollectionService struct {
	Storage *storage.Storage
}

func NewCollectionService(s *storage.Storage) *CollectionService {
	return &CollectionService{Storage: s}
}

func (s *CollectionService) CreateCollection(ctx context.Context, userID int64, name string, date string) error {
	collection := model.Collection{
		UserID: userID,
		Name:   name,
		Date:   date,
	}
	return s.Storage.CreateCollection(ctx, collection)
}

func (s *CollectionService) GetCollections(ctx context.Context, userID int64) ([]model.Collection, error) {
	return s.Storage.GetCollectionsByUser(ctx, userID)
}

func (s *CollectionService) DeleteCollection(ctx context.Context, userID int64, collectionID int64) error {
	return s.Storage.DeleteCollection(ctx, userID, collectionID)
}
