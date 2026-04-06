CREATE TABLE IF NOT EXISTS service_accounts (
    id               VARCHAR(255) PRIMARY KEY,
    name             VARCHAR(255) NOT NULL,
    api_key          VARCHAR(255) NOT NULL UNIQUE,
    allowed_services TEXT         NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_service_accounts_api_key ON service_accounts (api_key);
