CREATE TABLE uploaded_photos (
    id BIGSERIAL PRIMARY KEY,
    collection_id BIGINT NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL,
    original_url TEXT NOT NULL,
    thumbnail_url TEXT NOT NULL,
    file_name TEXT NOT NULL,
    file_ext TEXT NOT NULL,
    hash_name TEXT NOT NULL,
    uploaded_at TIMESTAMP NOT NULL DEFAULT now()
);
