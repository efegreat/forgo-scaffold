-- +migrate Up
-- Inbox of forgo-adapter's inbox.NewPostgres: one row per message and subscriber, which makes the
-- consumer server's handling idempotent. Every operation but the cleanup reads it by primary key.
CREATE TABLE inbox_events (
    message_id      TEXT        NOT NULL,
    subscriber_name TEXT        NOT NULL,
    status          SMALLINT    NOT NULL, -- worker.InboxStatus: 0 waiting, 1 completed, 2 failed, 3 rejected
    attempts        INT         NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (message_id, subscriber_name)
);

-- The cleanup deletes completed and rejected rows in (status, updated_at) order and stops at its
-- limit, so it never reads the failed rows, which are kept for inspection. Queries for failed rows,
-- such as a retry of them, use it too while failed rows are a small share of the table. Every status
-- change also sets updated_at, so status adds no index writes.
CREATE INDEX inbox_events_cleanup ON inbox_events (status, updated_at);

-- +migrate Down
DROP TABLE inbox_events;
