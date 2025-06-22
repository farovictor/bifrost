CREATE TABLE IF NOT EXISTS virtual_keys (
    id VARCHAR(255) PRIMARY KEY,
    scope VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    target VARCHAR(255) NOT NULL,
    rate_limit INTEGER NOT NULL
);
