CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  username TEXT UNIQUE NOT NULL DEFAULT '',
  email TEXT UNIQUE NOT NULL,
  password TEXT NOT NULL,
  refresh_token TEXT
);

CREATE TABLE collections (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  date TIMESTAMP NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE uploaded_photos (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  collection_id UUID NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  original_url TEXT NOT NULL,
  thumbnail_url TEXT NOT NULL,
  file_name TEXT NOT NULL,
  file_ext TEXT NOT NULL,
  hash_name TEXT NOT NULL,
  uploaded_at TIMESTAMP NOT NULL DEFAULT now()
);
