-- +migrate Up
-- Store of forgo-adapter's liveconfig.NewMySQL. Reads and versioned saves go by name only, so the
-- primary key is the only index. utf8mb4_0900_bin keeps names case sensitive, as LiveConfig.Name is.
CREATE TABLE live_configs (
    name       VARCHAR(255)    NOT NULL PRIMARY KEY,
    data       JSON            NOT NULL,
    version    BIGINT UNSIGNED NOT NULL,
    updated_at DATETIME(6)     NOT NULL
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_0900_bin;

-- +migrate Down
DROP TABLE live_configs;
