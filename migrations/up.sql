CREATE TABLE IF NOT EXISTS urls (
    id SERIAL PRIMARY KEY,
    url VARCHAR(255) NOT NULL,
    short_url VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_urls_short_url ON urls (short_url);