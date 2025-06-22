CREATE TABLE IF NOT EXISTS services (
    id VARCHAR(255) PRIMARY KEY,
    endpoint TEXT NOT NULL,
    root_key_id VARCHAR(255) NOT NULL REFERENCES root_keys(id)
);
