-- +migrate Up
-- Outbox of forgo-adapter's outbox.NewMySQL. The application inserts a row in the transaction of the
-- change that produced the message, and the producer server claims, publishes and deletes it.
-- DATETIME columns hold UTC: the storage stamps them with UTC_TIMESTAMP(6).
--
-- utf8mb4_0900_bin compares IDs and names byte for byte, as Go does. The default collation is case
-- and accent insensitive.
CREATE TABLE outbox_events (
    id            VARCHAR(255) NOT NULL PRIMARY KEY, -- Message.ID; UUIDv7 keeps claim order chronological
    name          VARCHAR(255) NOT NULL,
    payload       JSON         NOT NULL,
    headers       JSON,
    created_at    DATETIME(6)  NOT NULL DEFAULT (UTC_TIMESTAMP(6)),
    message_key   VARCHAR(255) NOT NULL DEFAULT '',
    destination   VARCHAR(255) NOT NULL DEFAULT '',
    processing_id VARCHAR(255),
    processing_at DATETIME(6),
    processed_at  DATETIME(6),

    -- Claim reads unprocessed rows (processed_at IS NULL, which MySQL treats as an equality) in id
    -- order, and the cleanup reads processed rows in (processed_at, id) order. Both read the index in
    -- that order and stop at their limit. Without this index, the claim walks the primary key past
    -- every processed row, and the cleanup sorts every expired row, which under REPEATABLE READ it
    -- also locks.
    INDEX outbox_events_processed (processed_at, id),

    -- Complete releases the rows of a claim by processing_id alone. Without this index, every batch
    -- scans the whole table.
    INDEX outbox_events_claim (processing_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_0900_bin;

-- +migrate Down
DROP TABLE outbox_events;
