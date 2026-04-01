CREATE TABLE IF NOT EXISTS api_keys (
    id              TEXT PRIMARY KEY,
    key_hash        TEXT UNIQUE NOT NULL,
    name            TEXT NOT NULL,
    tier            TEXT NOT NULL DEFAULT 'free',
    active          BOOLEAN NOT NULL DEFAULT true,
    rate_limit      INTEGER NOT NULL DEFAULT 10,
    token_limit     INTEGER NOT NULL DEFAULT 10000,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    last_used_at    TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS usage_records (
    id                  TEXT PRIMARY KEY,
    api_key_id          TEXT REFERENCES api_keys(id),
    provider            TEXT NOT NULL,
    model               TEXT NOT NULL,
    prompt_tokens       INTEGER,
    completion_tokens   INTEGER,
    total_tokens        INTEGER,
    latency_ms          BIGINT,
    cost_usd            DOUBLE PRECISION,
    status_code         INTEGER,
    created_at          TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX IF NOT EXISTS idx_usage_api_key_id ON usage_records(api_key_id);
CREATE INDEX IF NOT EXISTS idx_usage_created_at ON usage_records(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_usage_provider ON usage_records(provider);
CREATE INDEX IF NOT EXISTS idx_usage_model ON usage_records(model);
