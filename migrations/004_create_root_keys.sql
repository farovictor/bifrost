CREATE TABLE IF NOT EXISTS root_keys (
    id               VARCHAR(255) PRIMARY KEY,
    encrypted_api_key BYTEA        NOT NULL DEFAULT '',
    key_hint         VARCHAR(8)   NOT NULL DEFAULT ''
);
