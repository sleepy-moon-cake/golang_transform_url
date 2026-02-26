CREATE TABLE IF NOT EXISTS urls (
    id SERIAL PRIMARY KEY,
    short_url TEXT NOT NULL,
    original_url TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_urls_short_url ON urls(short_url);