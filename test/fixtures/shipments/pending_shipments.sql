-- Two shipments that wait for their dispatch, and one that was dispatched.
INSERT INTO shipments (order_id, sku, quantity, status, dispatched_at) VALUES
    ('0199f0a0-0000-7000-8000-000000000001', 'a-1', 2, 'pending', NULL),
    ('0199f0a0-0000-7000-8000-000000000002', 'b-2', 1, 'pending', NULL),
    ('0199f0a0-0000-7000-8000-000000000003', 'c-3', 5, 'dispatched', CURRENT_TIMESTAMP - INTERVAL '1 day');
