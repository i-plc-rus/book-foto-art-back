
-- Добавляем поля для OAuth аутентификации
ALTER TABLE users
ADD COLUMN IF NOT EXISTS oauth_provider TEXT,
ADD COLUMN IF NOT EXISTS oauth_id TEXT;

-- Делаем пароль необязательным для OAuth пользователей
ALTER TABLE users ALTER COLUMN password DROP NOT NULL;
