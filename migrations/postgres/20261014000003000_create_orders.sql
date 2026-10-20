-- +migrate Up
-- The orders of internal/order. Placing one also writes its order.placed message to outbox_events, in
-- the same transaction.
CREATE TABLE orders (
    id         TEXT        PRIMARY KEY, -- UUIDv7
    sku        TEXT        NOT NULL,
    quantity   INT         NOT NULL CHECK (quantity > 0),
    status     TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +migrate Down
DROP TABLE orders;
