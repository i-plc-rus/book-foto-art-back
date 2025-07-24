package service

import (
	"book-foto-art-back/internal/model"
	"book-foto-art-back/internal/storage/postgres"
	"context"
	"mime/multipart"
)

type UploadService struct {
	Storage *postgres.Storage
	S3      S3Uploader
}

type S3Uploader interface {
	UploadFile(file multipart.File, fileHeader *multipart.FileHeader, userID, collectionID int64) (originalURL, thumbURL string, err error)
}

func NewUploadService(s *postgres.Storage) *UploadService {
	return &UploadService{Storage: s}
}

func (s *UploadService) UploadFiles(ctx context.Context, userID int64, collectionID int64, files []*multipart.FileHeader) ([]model.UploadedPhoto, error) {
	var results []model.UploadedPhoto
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			return nil, err
		}
		defer file.Close()

		origURL, thumbURL, err := s.S3.UploadFile(file, fileHeader, userID, collectionID)
		if err != nil {
			return nil, err
		}
		results = append(results, model.UploadedPhoto{
			OriginalURL:  origURL,
			ThumbnailURL: thumbURL,
		})
	}
	return results, nil
}
