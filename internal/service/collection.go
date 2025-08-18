package service

import (
	"book-foto-art-back/internal/model"
	"book-foto-art-back/internal/shared"
	"book-foto-art-back/internal/storage/postgres"
	"book-foto-art-back/internal/storage/s3"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
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

func (s *CollectionService) GetCollectionPhotos(ctx context.Context, userID uuid.UUID, collectionID uuid.UUID, sortParam string) (
	[]model.UploadedPhoto, string, error) {
	// Выбираем параметр сортировки
	sort := shared.SortOption(sortParam)
	if _, ok := shared.ValidSorts[sort]; !ok {
		sort = shared.DefaultSort
	}
	// Получаем содержимое коллекции из БД
	photos, err := s.Postgres.GetCollectionPhotos(ctx, userID, collectionID, sort)
	if err != nil {
		log.Printf("Storage ERROR: %v\n", err)
		return []model.UploadedPhoto{}, "", err
	}
	return photos, string(sort), err
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
