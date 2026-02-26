CREATE TABLE IF NOT EXISTS urls (
    id SERIAL PRIMARY KEY,
    short_url VARCHAR(8) NOT NULL,
    original_url VARCHAR(250) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_urls_short_url ON urls(short_url);