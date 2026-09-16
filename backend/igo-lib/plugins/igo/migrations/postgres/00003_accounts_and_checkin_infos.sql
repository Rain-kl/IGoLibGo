-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS igo_accounts (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    name VARCHAR(128) NOT NULL,
    cookie TEXT NOT NULL DEFAULT '',
    cookie_expires_at TIMESTAMPTZ,
    cookie_source VARCHAR(32) NOT NULL DEFAULT '',
    checkin_token TEXT NOT NULL DEFAULT '',
    checkin_expires_at TIMESTAMPTZ,
    nickname VARCHAR(128) NOT NULL DEFAULT '',
    school VARCHAR(128) NOT NULL DEFAULT '',
    student_name VARCHAR(128) NOT NULL DEFAULT '',
    student_number VARCHAR(64) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_igo_accounts_user ON igo_accounts (user_id);

CREATE TABLE IF NOT EXISTS igo_checkin_infos (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    name VARCHAR(128) NOT NULL,
    beacon_uuid VARCHAR(64) NOT NULL DEFAULT '',
    major INTEGER NOT NULL DEFAULT 0,
    minor INTEGER NOT NULL DEFAULT 0,
    latitude VARCHAR(32) NOT NULL DEFAULT '',
    longitude VARCHAR(32) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_igo_checkin_infos_user ON igo_checkin_infos (user_id);

ALTER TABLE igo_pipeline_configs ADD COLUMN account_id BIGINT NOT NULL DEFAULT 0;
ALTER TABLE igo_pipeline_configs ADD COLUMN checkin_account_id BIGINT NOT NULL DEFAULT 0;
ALTER TABLE igo_pipeline_configs ADD COLUMN checkin_info_id BIGINT NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE igo_pipeline_configs DROP COLUMN IF EXISTS checkin_info_id;
ALTER TABLE igo_pipeline_configs DROP COLUMN IF EXISTS checkin_account_id;
ALTER TABLE igo_pipeline_configs DROP COLUMN IF EXISTS account_id;
DROP TABLE IF EXISTS igo_checkin_infos;
DROP TABLE IF EXISTS igo_accounts;
-- +goose StatementEnd
