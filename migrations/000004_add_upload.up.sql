CREATE TABLE uploaded_photos (
    id UUID PRIMARY KEY,
    collection_id UUID NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    original_url TEXT NOT NULL,
    thumbnail_url TEXT NOT NULL,
    file_name TEXT NOT NULL,
    file_ext TEXT NOT NULL,
    hash_name TEXT NOT NULL,
    uploaded_at TIMESTAMP NOT NULL DEFAULT now()
);
