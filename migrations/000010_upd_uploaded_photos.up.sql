
-- Добавляем поле is_favorite в таблицу uploaded_photos
ALTER TABLE uploaded_photos ADD COLUMN is_favorite BOOLEAN DEFAULT false;
