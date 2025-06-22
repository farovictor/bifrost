CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(255) PRIMARY KEY,
    api_key TEXT NOT NULL,
    UNIQUE(api_key)
);
