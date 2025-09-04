
-- Добавляем поля подписки в таблицу users
ALTER TABLE users ADD COLUMN subscription_active BOOLEAN DEFAULT false;
ALTER TABLE users ADD COLUMN subscription_expires_at TIMESTAMP;

-- Создаем таблицу платежей
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    yookassa_payment_id TEXT UNIQUE NOT NULL,
    amount DECIMAL(10, 2) NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending', -- 'pending', 'waiting_for_capture', 'succeeded', 'canceled'
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);
