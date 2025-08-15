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
	Postgres *postgres.Storage
	S3       *s3.S3Storage
}

func NewUploadService(pg *postgres.Storage, s3 *s3.S3Storage) *UploadService {
	return &UploadService{
		Postgres: pg,
		S3:       s3,
	}
}

func (s *UploadService) UploadFiles(ctx context.Context, userID uuid.UUID, collectionID uuid.UUID, files []*multipart.FileHeader) (
	[]model.UploadedPhoto, error) {

	var results []model.UploadedPhoto
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			return nil, err
		}
		defer file.Close()

		// Получаем имя файла и расширение
		fileName := fileHeader.Filename
		ext := strings.ToLower(filepath.Ext(fileName))

		// Читаем данные и хешируем
		buf, err := io.ReadAll(file)
		if err != nil {
			return nil, err
		}
		hash := sha1.Sum(buf)
		hashName := hex.EncodeToString(hash[:]) + ext

		// Загружаем файл в S3
		file.Seek(0, 0)
		originalURL, thumbnailURL, err := s.S3.UploadFile(file, fileHeader, userID, collectionID)
		if err != nil {
			return nil, fmt.Errorf("failed to upload file to S3: %w", err)
		}

		upload := &model.UploadedPhoto{
			UserID:       userID,
			CollectionID: collectionID,
			FileName:     fileName,
			FileExt:      ext,
			HashName:     hashName,
			OriginalURL:  originalURL,
			ThumbnailURL: thumbnailURL,
			UploadedAt:   time.Now(),
		}

		res, err := s.Postgres.SaveUpload(ctx, upload)
		if err != nil {
			return nil, err
		}

		results = append(results, *res)
	}

	return results, nil
}
