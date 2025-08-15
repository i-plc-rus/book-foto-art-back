package s3

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/disintegration/imaging"
	"github.com/google/uuid"
)

type S3Storage struct {
	client   *s3.Client
	bucket   string
	region   string
	endpoint string
}

type S3Config struct {
	Region          string
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
	Endpoint        string
}

func NewS3Storage(cfg S3Config) (*S3Storage, error) {
	// Создаем AWS конфигурацию с кастомным endpoint
	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(aws.CredentialsProviderFunc(
			func(ctx context.Context) (aws.Credentials, error) {
				return aws.Credentials{
					AccessKeyID:     cfg.AccessKeyID,
					SecretAccessKey: cfg.SecretAccessKey,
				}, nil
			},
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Создаем S3 клиент с кастомным endpoint
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		// o.BaseEndpoint = aws.String(cfg.Endpoint)
		o.BaseEndpoint = aws.String(cfg.Endpoint)
		o.UsePathStyle = true // Используем path-style для совместимости с S3-совместимыми сервисами
		o.Region = cfg.Region
	})

	return &S3Storage{
		client:   client,
		bucket:   cfg.Bucket,
		region:   cfg.Region,
		endpoint: cfg.Endpoint,
	}, nil
}

// UploadFile загружает файл в S3 и возвращает URL
func (s *S3Storage) UploadFile(file multipart.File, fileHeader *multipart.FileHeader, userID, collectionID uuid.UUID) (
	originalURL, thumbnailURL string, err error) {

	// Читаем содержимое файла
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return "", "", fmt.Errorf("failed to read file: %w", err)
	}

	// Генерируем уникальное имя файла
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	fileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	// Формируем S3 пути
	originalPath := fmt.Sprintf("collection_%s/originals/%s", collectionID.String(), fileName)
	thumbPath := fmt.Sprintf("collection_%s/thumbnails/%s", collectionID.String(), fileName)

	// Создаем и загружаем миниатюру
	thumbBytes, err := s.createThumbnail(fileBytes)
	if err == nil {
		thumbnailURL, err = s.uploadBytes(thumbBytes, thumbPath, fileHeader.Header.Get("Content-Type"))
		if err != nil {
			log.Printf("failed to upload thumbnail: %v", err)
		}
	}

	// Загружаем оригинал
	originalURL, err = s.uploadBytes(fileBytes, originalPath, fileHeader.Header.Get("Content-Type"))
	if err != nil {
		return "", "", fmt.Errorf("failed to upload original: %w", err)
	}

	return originalURL, thumbnailURL, nil
}

// createThumbnail создает миниатюру изображения
func (s *S3Storage) createThumbnail(data []byte) ([]byte, error) {
	// Декодируем изображение
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Создаем миниатюру
	thumb := imaging.Thumbnail(img, 300, 300, imaging.Lanczos)

	// Кодируем миниатюру в JPEG
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, thumb, &jpeg.Options{Quality: 85})
	if err != nil {
		return nil, fmt.Errorf("failed to encode thumbnail: %w", err)
	}

	return buf.Bytes(), nil
}

// uploadBytes загружает байты в S3
func (s *S3Storage) uploadBytes(data []byte, key, contentType string) (string, error) {
	ctx := context.Background()

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
		// ContentLength: aws.Int64(info),
		// ACL:           "public-read",
		// CacheControl:  aws.String("max-age=31536000"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	// Формируем URL для доступа к файлу
	url := fmt.Sprintf("%s/%s/%s", s.endpoint, s.bucket, key)
	return url, nil
}

// DeleteFile удаляет файл из S3
func (s *S3Storage) DeleteFile(key string) error {
	_, err := s.client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete from S3: %w", err)
	}
	return nil
}
