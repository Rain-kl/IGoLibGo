-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS igo_pipeline_configs (
    id VARCHAR(64) PRIMARY KEY,
    user_id BIGINT NOT NULL,
    name VARCHAR(128) NOT NULL,
    cookie TEXT NOT NULL,
    cookie_expires_at DATETIME,
    library_id INTEGER NOT NULL,
    library_name VARCHAR(255) NOT NULL,
    floor VARCHAR(64) NOT NULL DEFAULT '',
    seat_key VARCHAR(64) NOT NULL,
    seat_name VARCHAR(128) NOT NULL DEFAULT '',
    auto_checkin BOOLEAN NOT NULL DEFAULT 0,
    checkin_token TEXT,
    checkin_expires_at DATETIME,
    beacon_uuid VARCHAR(64) NOT NULL DEFAULT '',
    major INTEGER NOT NULL DEFAULT 0,
    minor INTEGER NOT NULL DEFAULT 0,
    latitude VARCHAR(32) NOT NULL DEFAULT '',
    longitude VARCHAR(32) NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_igo_pipeline_configs_user ON igo_pipeline_configs (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS igo_pipeline_configs;
-- +goose StatementEnd
