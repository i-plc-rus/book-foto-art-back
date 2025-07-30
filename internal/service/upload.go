package service

import (
	"book-foto-art-back/internal/model"
	"book-foto-art-back/internal/storage/postgres"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"image"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
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

func (s *UploadService) UploadFiles(ctx context.Context, userID uuid.UUID, collectionID uuid.UUID, files []*multipart.FileHeader) (
	[]model.UploadedPhoto, error) {
	var results []model.UploadedPhoto
	//basePath := fmt.Sprintf("./data/user_%d/collection_%d", userID, collectionID)
	//os.MkdirAll(filepath.Join(basePath, "thumbs"), os.ModePerm)
	basePath := fmt.Sprintf("/uploads/collection_%s", collectionID.String())
	thumbsPath := filepath.Join(basePath, "thumbs")

	// Создаём директории
	if err := os.MkdirAll(thumbsPath, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create directories: %w", err)
	}

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

		// Сохраняем оригинал
		origPath := filepath.Join(basePath, fileName)
		if err := os.WriteFile(origPath, buf, 0644); err != nil {
			return nil, err
		}

		// Создаем миниатюру (если возможно)
		var thumbPath string
		img, _, err := image.Decode(strings.NewReader(string(buf)))
		if err == nil {
			thumb := imaging.Thumbnail(img, 300, 300, imaging.Lanczos)
			thumbPath = filepath.Join(basePath, "thumbs", fileName)
			_ = imaging.Save(thumb, thumbPath)
		}

		upload := &model.UploadedPhoto{
			UserID:       userID,
			CollectionID: collectionID,
			FileName:     fileHeader.Filename,
			FileExt:      ext,
			HashName:     fileName,
			OriginalURL:  origPath,
			ThumbnailURL: thumbPath,
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
