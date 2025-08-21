package s3

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	exif "github.com/rwcarlsen/goexif/exif"
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
)

type S3Storage struct {
	client   *minio.Client
	bucket   string
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
	// Парсим endpoint для получения host и port
	endpoint := strings.TrimPrefix(cfg.Endpoint, "https://")

	// Создаем minio клиент
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: true,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	// Проверяем существование bucket
	exists, err := client.BucketExists(context.Background(), cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("bucket %s does not exist", cfg.Bucket)
	}

	return &S3Storage{
		client:   client,
		bucket:   cfg.Bucket,
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

	// Получаем имя файла и расширение
	fileName := fileHeader.Filename
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))

	// Формируем S3 пути
	originalPath := fmt.Sprintf("collection_%s/originals/%s", collectionID.String(), fileName)
	thumbPath := fmt.Sprintf("collection_%s/thumbnails/%s", collectionID.String(), fileName)

	// Создаем и загружаем миниатюру
	thumbBytes, err := s.createThumbnail(fileBytes, ext)
	if err != nil {
		log.Printf("failed to create thumbnail: %v", err)
	}
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
func (s *S3Storage) createThumbnail(data []byte, ext string) ([]byte, error) {
	// Декодируем изображение
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Корректируем ориентацию на основе EXIF данных
	correctedImg, err := s.correctImageOrientation(data, img)
	if err != nil {
		log.Printf("Warning: failed to correct image orientation: %v", err)
		correctedImg = img
	}

	// Создаем миниатюру
	thumb := imaging.Thumbnail(correctedImg, 300, 300, imaging.Lanczos)

	// Кодируем миниатюру в зависимости от расширения файла
	var buf bytes.Buffer
	switch ext {
	case ".jpg", ".jpeg":
		err = jpeg.Encode(&buf, thumb, &jpeg.Options{Quality: 100})
	case ".png":
		err = png.Encode(&buf, thumb)
	case ".tiff":
		err = tiff.Encode(&buf, thumb, &tiff.Options{Compression: tiff.Uncompressed})
	case ".bmp":
		err = bmp.Encode(&buf, thumb)
	default:
		err = jpeg.Encode(&buf, thumb, &jpeg.Options{Quality: 100})
	}
	if err != nil {
		return nil, fmt.Errorf("failed to encode thumbnail: %w", err)
	}
	return buf.Bytes(), nil
}

// correctImageOrientation корректирует ориентацию изображения на основе EXIF данных
func (s *S3Storage) correctImageOrientation(data []byte, img image.Image) (image.Image, error) {
	// Читаем EXIF данные
	exifData, err := exif.Decode(bytes.NewReader(data))
	if err != nil {
		// Если EXIF данные отсутствуют или не могут быть прочитаны, возвращаем оригинал
		return img, fmt.Errorf("no EXIF data: %w", err)
	}

	// Получаем тег ориентации
	orientation, err := exifData.Get(exif.Orientation)
	if err != nil {
		// Если тег ориентации отсутствует, возвращаем оригинал
		return img, fmt.Errorf("no orientation tag: %w", err)
	}

	// Получаем значение ориентации
	orientationValue, err := orientation.Int(0)
	if err != nil {
		return img, fmt.Errorf("failed to get orientation value: %w", err)
	}

	fmt.Printf("img: %v\norientationValue: %v\n", img.Bounds(), orientationValue)

	// Применяем соответствующее преобразование
	switch orientationValue {
	case 1: // Normal
		return img, nil
	case 2: // Flip horizontal
		return imaging.FlipH(img), nil
	case 3: // Rotate 180
		return imaging.Rotate180(img), nil
	case 4: // Flip vertical
		return imaging.FlipV(img), nil
	case 5: // Transpose (flip horizontal + rotate 90 CCW)
		img = imaging.FlipH(img)
		return imaging.Rotate90(img), nil
	case 6: // Rotate 90 CW
		return imaging.Rotate270(img), nil
	case 7: // Transverse (flip horizontal + rotate 270 CW)
		img = imaging.FlipH(img)
		return imaging.Rotate270(img), nil
	case 8: // Rotate 90 CCW
		return imaging.Rotate270(img), nil
	default:
		return img, nil
	}
}

// uploadBytes загружает байты в S3
func (s *S3Storage) uploadBytes(data []byte, path, contentType string) (string, error) {
	// Загружаем файл в S3
	_, err := s.client.PutObject(
		context.Background(),
		s.bucket,
		path,
		bytes.NewReader(data),
		int64(len(data)),
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	// Формируем URL для доступа к файлу
	url := fmt.Sprintf("%s/%s/%s", s.endpoint, s.bucket, path)
	return url, nil
}

// DeletePhoto удаляет файлы из S3
func (s *S3Storage) DeletePhoto(ctx context.Context, collectionID uuid.UUID, fileName string) error {

	// Удаляем оригинал
	err := s.client.RemoveObject(
		context.Background(),
		s.bucket,
		fmt.Sprintf("collection_%s/originals/%s", collectionID.String(), fileName),
		minio.RemoveObjectOptions{},
	)
	if err != nil {
		return fmt.Errorf("failed to delete original from S3: %w", err)
	}
	// Удаляем миниатюру
	err = s.client.RemoveObject(
		context.Background(),
		s.bucket,
		fmt.Sprintf("collection_%s/thumbnails/%s", collectionID.String(), fileName),
		minio.RemoveObjectOptions{},
	)
	if err != nil {
		return fmt.Errorf("failed to delete thumbnail from S3: %w", err)
	}

	return nil
}

// DeleteCollection удаляет коллекцию из S3
func (s *S3Storage) DeleteCollection(ctx context.Context, collectionID uuid.UUID) error {
	// Получаем список всех объектов с префиксом collection_id/
	objectsCh := s.client.ListObjects(ctx, s.bucket, minio.ListObjectsOptions{
		Prefix:    fmt.Sprintf("collection_%s/", collectionID.String()),
		Recursive: true,
	})

	// Удаляем файлы
	for obj := range objectsCh {
		if obj.Err != nil {
			continue
		}
		err := s.client.RemoveObject(ctx, s.bucket, obj.Key, minio.RemoveObjectOptions{})
		if err != nil {
			log.Printf("Error deleting %s: %v", obj.Key, err)
		}
	}
	return nil
}
