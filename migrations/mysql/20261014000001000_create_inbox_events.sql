-- +migrate Up
-- Inbox of forgo-adapter's inbox.NewMySQL: one row per message and subscriber, which makes the
-- consumer server's handling idempotent. Every operation but the cleanup reads it by primary key.
-- DATETIME columns hold UTC: the storage stamps them with UTC_TIMESTAMP(6).
--
-- utf8mb4_0900_bin compares IDs byte for byte. Under the default, case insensitive collation, two
-- messages whose IDs differ only in case share a row, and the second is dropped as a duplicate.
CREATE TABLE inbox_events (
    message_id      VARCHAR(255) NOT NULL,
    subscriber_name VARCHAR(255) NOT NULL,
    status          TINYINT      NOT NULL, -- worker.InboxStatus: 0 waiting, 1 completed, 2 failed, 3 rejected
    attempts        INT          NOT NULL,
    created_at      DATETIME(6)  NOT NULL,
    updated_at      DATETIME(6)  NOT NULL,
    PRIMARY KEY (message_id, subscriber_name),

    -- The cleanup deletes completed and rejected rows in (status, updated_at) order and stops at its
    -- limit, so it never reads, nor under REPEATABLE READ locks, the failed rows, which are kept for
    -- inspection. Every status change also sets updated_at, so status adds no index writes. A retry
    -- of failed rows forces it, at READ COMMITTED, as forgo's REFERENCE.md shows: without the hint,
    -- MySQL scans the whole table once failed rows are a large share of it, and under REPEATABLE
    -- READ, consumers wait for the retry to commit.
    INDEX inbox_events_cleanup (status, updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_0900_bin;

-- +migrate Down
DROP TABLE inbox_events;
