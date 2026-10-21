-- Two messages that the subscriber shipping failed, one it completed, and one that the subscriber order
-- failed.
INSERT INTO inbox_events (message_id, subscriber_name, status, attempts, created_at, updated_at) VALUES
    ('m1', 'shipping', 2, 5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('m2', 'shipping', 2, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('m3', 'shipping', 1, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('m4', 'order', 2, 5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);
