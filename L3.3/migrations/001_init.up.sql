CREATE TABLE IF NOT EXISTS comments (
    id BIGSERIAL PRIMARY KEY,
    parent_id BIGINT REFERENCES comments(id),
    author TEXT NOT NULL,
    body TEXT NOT NULL CHECK (length(trim(body)) > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    search_vector tsvector
);

CREATE INDEX IF NOT EXISTS idx_comments_parent_id
    ON comments(parent_id);

CREATE INDEX IF NOT EXISTS idx_comments_created_at
    ON comments(created_at);

CREATE INDEX IF NOT EXISTS idx_comments_search_vector
    ON comments
    USING GIN(search_vector);