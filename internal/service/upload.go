package service

import (
	"book-foto-art-back/internal/model"
	"book-foto-art-back/internal/storage"
	"context"
	"mime/multipart"
)

type UploadService struct {
	Storage *storage.Storage
	S3      S3Uploader
}

type S3Uploader interface {
	UploadFile(file multipart.File, fileHeader *multipart.FileHeader, userID, collectionID int64) (originalURL, thumbURL string, err error)
}

func NewUploadService(s *storage.Storage, s3 S3Uploader) *UploadService {
	return &UploadService{Storage: s, S3: s3}
}

func (s *UploadService) UploadFiles(ctx context.Context, userID int64, collectionID int64, files []*multipart.FileHeader) ([]model.UploadResult, error) {
	var results []model.UploadResult
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
		results = append(results, model.UploadResult{
			OriginalURL:  origURL,
			ThumbnailURL: thumbURL,
		})
	}
	return results, nil
}
