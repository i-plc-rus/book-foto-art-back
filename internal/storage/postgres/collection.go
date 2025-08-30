package postgres

import (
	"book-foto-art-back/internal/model"
	"book-foto-art-back/internal/shared"
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CollectionStorage struct {
	DB *pgxpool.Pool
}

func (s *Storage) CreateCollection(ctx context.Context, col model.Collection) (*model.Collection, error) {
	row := s.DB.QueryRow(ctx,
		`INSERT INTO collections (name, date, user_id, cover_url, cover_thumbnail_url)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id`,
		col.Name, col.Date, col.UserID, col.CoverURL, col.CoverThumbnailURL,
	)
	var id uuid.UUID
	if err := row.Scan(&id); err != nil {
		return nil, err
	}
	col.ID = id
	return &col, nil
}

func (s *Storage) GetCollectionInfo(ctx context.Context, userID uuid.UUID, collectionID uuid.UUID) (*model.Collection, error) {
	row := s.DB.QueryRow(ctx, `
		SELECT c.id, c.user_id, c.name, c.date, c.created_at, c.cover_url,
			   c.cover_thumbnail_url, u.username, c.is_published,
		(SELECT COUNT(*) FROM uploaded_photos WHERE collection_id = c.id) AS count_photos
		FROM collections c
		JOIN users u ON c.user_id = u.id
		WHERE c.user_id = $1 AND c.id = $2
		GROUP BY c.id, c.user_id, c.name, c.date, c.created_at, c.cover_url,
                 c.cover_thumbnail_url, u.username, c.is_published
		`, userID, collectionID,
	)
	var col model.Collection
	if err := row.Scan(
		&col.ID, &col.UserID, &col.Name, &col.Date, &col.CreatedAt, &col.CoverURL,
		&col.CoverThumbnailURL, &col.UserName, &col.IsPublished, &col.CountPhotos); err != nil {
		return nil, err
	}
	return &col, nil
}

func (s *Storage) GetCollections(ctx context.Context, userID uuid.UUID, searchParam string) ([]model.Collection, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT c.id, c.user_id, c.name, c.date, c.created_at, c.cover_url,
		c.cover_thumbnail_url, u.username, c.is_published,
		(SELECT COUNT(*) FROM uploaded_photos WHERE collection_id = c.id) AS count_photos
		FROM collections c
		JOIN users u ON c.user_id = u.id
        WHERE c.user_id = $1 AND ($2 = '' OR c.name ILIKE $3)
        ORDER BY
            CASE
                WHEN $2 = '' THEN 1
                WHEN c.name ILIKE $4 THEN 1  -- точное совпадение в начале
                WHEN c.name ILIKE $5 THEN 2  -- совпадение в начале
                ELSE 3                       -- совпадение в любом месте
            END,
            c.date DESC
		`, userID, searchParam, "%"+searchParam+"%", searchParam+"%", "%"+searchParam+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var collections []model.Collection
	for rows.Next() {
		var c model.Collection
		err := rows.Scan(
			&c.ID, &c.UserID, &c.Name, &c.Date, &c.CreatedAt, &c.CoverURL,
			&c.CoverThumbnailURL, &c.UserName, &c.IsPublished, &c.CountPhotos)
		if err != nil {
			return nil, err
		}
		c.UserID = userID
		collections = append(collections, c)
	}
	return collections, nil
}

func (s *Storage) DeleteCollection(ctx context.Context, userID uuid.UUID, collectionID uuid.UUID) error {
	res, err := s.DB.Exec(ctx, "DELETE FROM collections WHERE user_id = $1 AND id = $2", userID, collectionID)
	if err != nil {
		return err
	}
	rowsAffected := res.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Storage) UpdateCollectionCover(
	ctx context.Context, userID uuid.UUID, collectionID uuid.UUID, uploadedPhotoID uuid.UUID) error {

	// Сначала проверяем, что uploaded_photo принадлежит коллекции и пользователю
	var photo model.UploadedPhoto
	err := s.DB.QueryRow(ctx,
		`SELECT original_url, thumbnail_url
		 FROM uploaded_photos
		 WHERE user_id = $1 AND collection_id = $2 AND id = $3`,
		userID, collectionID, uploadedPhotoID,
	).Scan(&photo.OriginalURL, &photo.ThumbnailURL)
	if err != nil {
		return err
	}

	// Обновляем cover_url и cover_thumbnail_url коллекции
	res, err := s.DB.Exec(ctx,
		`UPDATE collections
		 SET cover_url = $1, cover_thumbnail_url = $2
		 WHERE user_id = $3 AND id = $4`,
		photo.OriginalURL, photo.ThumbnailURL, userID, collectionID)
	if err != nil {
		return err
	}

	rowsAffected := res.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Storage) GetCollectionPhotos(ctx context.Context, userID uuid.UUID, collectionID uuid.UUID, sort shared.SortOption) (
	[]model.UploadedPhoto, error) {

	// Определяем SQL для сортировки
	var orderBy string
	switch sort {
	case shared.SortUploadedNew:
		orderBy = " ORDER BY uploaded_at DESC"
	case shared.SortUploadedOld:
		orderBy = " ORDER BY uploaded_at ASC"
	case shared.SortNameAZ:
		orderBy = " ORDER BY file_name ASC"
	case shared.SortNameZA:
		orderBy = " ORDER BY file_name DESC"
	case shared.SortRandom:
		orderBy = " ORDER BY RANDOM()"
	default:
		orderBy = " ORDER BY uploaded_at DESC"
	}

	rows, err := s.DB.Query(ctx,
		`SELECT id, collection_id, user_id, original_url, thumbnail_url, file_name, file_ext, hash_name, uploaded_at
		 FROM uploaded_photos
		 WHERE user_id = $1 AND collection_id = $2`+orderBy, userID, collectionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.UploadedPhoto
	for rows.Next() {
		var f model.UploadedPhoto
		if err := rows.Scan(&f.ID, &f.CollectionID, &f.UserID, &f.OriginalURL, &f.ThumbnailURL,
			&f.FileName, &f.FileExt, &f.HashName, &f.UploadedAt); err != nil {
			return nil, err
		}
		result = append(result, f)
	}
	return result, nil
}

func (s *Storage) PublishCollection(ctx context.Context, userID uuid.UUID, collectionID uuid.UUID, token string, newLink string) (string, error) {

	// Проверяем наличие короткой ссылки на коллекцию
	row := s.DB.QueryRow(ctx, `
		SELECT url
		FROM short_links
		WHERE collection_id = $1
	`, collectionID)
	var link string
	_ = row.Scan(&link)
	if link != "" {
		return link, nil
	}

	// Если короткая ссылка не существует, то обновляем коллекцию и создаем короткую ссылку
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		UPDATE collections
		SET is_published = true
		WHERE user_id = $1 AND id = $2
	`, userID, collectionID)
	if err != nil {
		return "", err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO short_links (collection_id, token, url)
		VALUES ($1, $2, $3)
	`, collectionID, token, newLink)
	if err != nil {
		return "", err
	}
	err = tx.Commit(ctx)
	if err != nil {
		return "", err
	}
	return newLink, nil
}

func (s *Storage) UpdateShortLink(ctx context.Context, token string) error {
	res, err := s.DB.Exec(ctx, `
		UPDATE short_links
		SET click_count = click_count + 1
		WHERE token = $1
	`, token)
	if err != nil {
		return err
	}
	rowsAffected := res.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Storage) GetPublicCollectionInfo(ctx context.Context, token string) (*model.Collection, error) {
	row := s.DB.QueryRow(ctx, `
		SELECT c.id, c.name, c.date, c.cover_url, c.cover_thumbnail_url, u.username,
		(SELECT COUNT(*) FROM uploaded_photos WHERE collection_id = c.id) AS count_photos
		FROM collections c
		JOIN users u ON c.user_id = u.id
		WHERE c.is_published = true AND c.id = (SELECT collection_id FROM short_links WHERE token = $1)
	`, token,
	)
	var col model.Collection
	if err := row.Scan(
		&col.ID, &col.Name, &col.Date, &col.CoverURL, &col.CoverThumbnailURL, &col.UserName, &col.CountPhotos); err != nil {
		return nil, err
	}
	return &col, nil
}

func (s *Storage) GetPublicCollectionPhotos(ctx context.Context, token string, sort shared.SortOption) (
	[]model.UploadedPhoto, error) {

	// Определяем SQL для сортировки
	var orderBy string
	switch sort {
	case shared.SortUploadedNew:
		orderBy = " ORDER BY uploaded_at DESC"
	case shared.SortUploadedOld:
		orderBy = " ORDER BY uploaded_at ASC"
	case shared.SortNameAZ:
		orderBy = " ORDER BY file_name ASC"
	case shared.SortNameZA:
		orderBy = " ORDER BY file_name DESC"
	case shared.SortRandom:
		orderBy = " ORDER BY RANDOM()"
	default:
		orderBy = " ORDER BY uploaded_at DESC"
	}

	rows, err := s.DB.Query(ctx,
		`SELECT id, collection_id, user_id, original_url, thumbnail_url, file_name, file_ext, hash_name, uploaded_at
		 FROM uploaded_photos
		 WHERE collection_id = (SELECT collection_id FROM short_links WHERE token = $1)`+orderBy, token,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.UploadedPhoto
	for rows.Next() {
		var f model.UploadedPhoto
		if err := rows.Scan(&f.ID, &f.CollectionID, &f.UserID, &f.OriginalURL, &f.ThumbnailURL,
			&f.FileName, &f.FileExt, &f.HashName, &f.UploadedAt); err != nil {
			return nil, err
		}
		result = append(result, f)
	}
	return result, nil
}

func (s *Storage) UnpublishCollection(ctx context.Context, userID uuid.UUID, collectionID uuid.UUID) error {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	res, err := tx.Exec(ctx, `
		UPDATE collections
		SET is_published = false
		WHERE user_id = $1 AND id = $2
	`, userID, collectionID)
	if err != nil {
		return err
	}
	rowsAffected := res.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	res, err = tx.Exec(ctx, `
		DELETE FROM short_links
		WHERE collection_id = $1
	`, collectionID)
	if err != nil {
		return err
	}
	rowsAffected = res.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	err = tx.Commit(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) GetShortLinkInfo(ctx context.Context, token string) (model.ShortLink, error) {
	row := s.DB.QueryRow(ctx, `
    SELECT id, collection_id, url, token, created_at, click_count
    FROM short_links
    WHERE token = $1
	`, token)
	var shortLink model.ShortLink
	if err := row.Scan(
		&shortLink.ID, &shortLink.CollectionID, &shortLink.URL, &shortLink.Token, &shortLink.CreatedAt, &shortLink.ClickCount,
	); err != nil {
		return model.ShortLink{}, err
	}
	return shortLink, nil
}

func (s *Storage) GetCollectionPhoto(ctx context.Context, userID uuid.UUID, photoID uuid.UUID) (*model.UploadedPhoto, error) {
	row := s.DB.QueryRow(ctx, `
		SELECT id, collection_id, user_id, original_url, thumbnail_url, file_name, file_ext, hash_name, uploaded_at
		FROM uploaded_photos
		WHERE user_id = $1 AND id = $2`, userID, photoID)

	var photo model.UploadedPhoto
	if err := row.Scan(&photo.ID, &photo.CollectionID, &photo.UserID, &photo.OriginalURL, &photo.ThumbnailURL,
		&photo.FileName, &photo.FileExt, &photo.HashName, &photo.UploadedAt); err != nil {
		return nil, err
	}
	return &photo, nil
}

func (s *Storage) DeletePhoto(ctx context.Context, userID uuid.UUID, photoID uuid.UUID) error {
	res, err := s.DB.Exec(ctx, `
		DELETE FROM uploaded_photos
		WHERE user_id = $1 AND id = $2
	`, userID, photoID)
	if err != nil {
		return err
	}
	rowsAffected := res.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}
