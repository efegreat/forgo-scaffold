-- +migrate Up
-- The shipments of internal/domain/shipping, one per order: the worker records one for each order.placed
-- message, and the dispatch job dispatches the pending ones. No foreign key to orders, since the API and
-- the worker may own separate databases.
CREATE TABLE shipments (
    order_id      TEXT        PRIMARY KEY,
    sku           TEXT        NOT NULL,
    quantity      INT         NOT NULL,
    status        TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    dispatched_at TIMESTAMPTZ
);

-- The dispatch job reads the pending shipments only.
CREATE INDEX shipments_pending ON shipments (created_at) WHERE status = 'pending';

-- +migrate Down
DROP TABLE shipments;
