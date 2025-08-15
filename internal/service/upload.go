package service

import (
	"book-foto-art-back/internal/model"
	"book-foto-art-back/internal/storage/postgres"
	"book-foto-art-back/internal/storage/s3"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type UploadService struct {
	Storage *postgres.Storage
	S3      *s3.S3Storage
}

func NewUploadService(s *postgres.Storage, s3 *s3.S3Storage) *UploadService {
	return &UploadService{
		Storage: s,
		S3:      s3,
	}
}

func (s *UploadService) UploadFiles(ctx context.Context, userID uuid.UUID, collectionID uuid.UUID, files []*multipart.FileHeader) (
	[]model.UploadedPhoto, error) {

	var results []model.UploadedPhoto
	for _, fileHeader := range files {
		src, err := fileHeader.Open()
		if err != nil {
			return nil, err
		}
		defer src.Close()

		// Читаем данные и хешируем
		buf, err := io.ReadAll(src)
		if err != nil {
			return nil, err
		}
		hash := sha1.Sum(buf)
		hashName := hex.EncodeToString(hash[:])
		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		fileName := hashName + ext

		// Загружаем файл в S3
		src.Seek(0, 0)
		originalURL, thumbnailURL, err := s.S3.UploadFile(src, fileHeader, userID, collectionID)
		if err != nil {
			return nil, fmt.Errorf("failed to upload file to S3: %w", err)
		}

		upload := &model.UploadedPhoto{
			UserID:       userID,
			CollectionID: collectionID,
			FileName:     fileHeader.Filename,
			FileExt:      ext,
			HashName:     fileName,
			OriginalURL:  originalURL,
			ThumbnailURL: thumbnailURL,
			UploadedAt:   time.Now(),
		}

		res, err := s.Storage.SaveUpload(ctx, upload)
		if err != nil {
			return nil, err
		}

		results = append(results, *res)
	}

	return results, nil
}
