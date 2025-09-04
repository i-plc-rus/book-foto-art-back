
-- Удаляем поля подписки из users
ALTER TABLE users DROP COLUMN IF EXISTS subscription_active;
ALTER TABLE users DROP COLUMN IF EXISTS subscription_expires_at;

-- Удаляем таблицу платежей
DROP TABLE IF EXISTS payments;
