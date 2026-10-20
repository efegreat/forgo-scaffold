-- +migrate Up
-- Outbox of forgo-adapter's outbox.NewPostgres. The application inserts a row in the transaction of
-- the change that produced the message, and the producer server claims, publishes and deletes it.
CREATE TABLE outbox_events (
    id            TEXT        PRIMARY KEY, -- Message.ID; UUIDv7 keeps claim order chronological
    name          TEXT        NOT NULL,
    payload       JSONB       NOT NULL,
    headers       JSONB,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    message_key   TEXT        NOT NULL DEFAULT '',
    destination   TEXT        NOT NULL DEFAULT '',
    processing_id TEXT,
    processing_at TIMESTAMPTZ,
    processed_at  TIMESTAMPTZ
);

-- Claim reads unprocessed rows in id order and stops at its limit. The index holds only those rows,
-- so it stays small however many processed rows wait for the cleanup. (processed_at, id) would not
-- do: PostgreSQL does not treat IS NULL as an equality, so it cannot read that index in id order,
-- and walks the primary key past every processed row instead.
CREATE INDEX outbox_events_pending ON outbox_events (id) WHERE processed_at IS NULL;

-- Complete releases the rows of a claim by processing_id alone. Without this index, every batch
-- scans the whole backlog, which is largest while the broker is down. Only claimed rows have one.
-- The predicates of these partial indexes hold in generic plans too, whatever the parameters.
CREATE INDEX outbox_events_claim ON outbox_events (processing_id) WHERE processing_id IS NOT NULL;

-- The cleanup deletes processed rows in (processed_at, id) order and stops at its limit.
CREATE INDEX outbox_events_processed ON outbox_events (processed_at, id) WHERE processed_at IS NOT NULL;

-- +migrate Down
DROP TABLE outbox_events;
