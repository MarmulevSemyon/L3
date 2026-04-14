CREATE TABLE images (
    id TEXT PRIMARY KEY,
    original_name TEXT NOT NULL,
    content_type TEXT NOT NULL,
    status TEXT NOT NULL,
    operation_resize BOOLEAN NOT NULL DEFAULT FALSE,
    operation_thumb BOOLEAN NOT NULL DEFAULT FALSE,
    operation_watermark BOOLEAN NOT NULL DEFAULT FALSE,
    original_path TEXT NOT NULL,
    processed_path TEXT NOT NULL DEFAULT '',
    thumb_path TEXT NOT NULL DEFAULT '',
    error_text TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);