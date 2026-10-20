-- +migrate Up
-- Store of forgo-adapter's liveconfig.NewPostgres. Reads and versioned saves go by name only, so the
-- primary key is the only index.
CREATE TABLE live_configs (
    name       TEXT        PRIMARY KEY,
    data       JSONB       NOT NULL,
    version    BIGINT      NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

-- +migrate Down
DROP TABLE live_configs;
