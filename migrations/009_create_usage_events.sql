CREATE TABLE IF NOT EXISTS usage_events (
    id                BIGSERIAL PRIMARY KEY,
    key_id            VARCHAR(255) NOT NULL,
    timestamp         TIMESTAMPTZ  NOT NULL,
    status_code       INTEGER      NOT NULL,
    service           VARCHAR(255) NOT NULL,
    latency_ms        BIGINT       NOT NULL,
    prompt_tokens     INTEGER      NOT NULL DEFAULT 0,
    completion_tokens INTEGER      NOT NULL DEFAULT 0,
    total_tokens      INTEGER      NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_usage_events_key_id   ON usage_events (key_id);
CREATE INDEX IF NOT EXISTS idx_usage_events_timestamp ON usage_events (timestamp);
