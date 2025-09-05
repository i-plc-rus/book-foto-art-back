
-- Удаляем поля is_favorite из uploaded_photos
ALTER TABLE uploaded_photos DROP COLUMN IF EXISTS is_favorite;
