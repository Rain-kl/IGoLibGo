-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS igo_sessions (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    cookie TEXT NOT NULL,
    source VARCHAR(32) NOT NULL DEFAULT '',
    saved_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ,
    can_auto_restore BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_igo_sessions_user ON igo_sessions (user_id);

CREATE TABLE IF NOT EXISTS igo_venues (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    library_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    floor VARCHAR(64) NOT NULL DEFAULT '',
    is_open BOOLEAN NOT NULL DEFAULT FALSE,
    total_seats INTEGER NOT NULL DEFAULT 0,
    used_seats INTEGER NOT NULL DEFAULT 0,
    booked_seats INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_igo_venues_user ON igo_venues (user_id);

CREATE TABLE IF NOT EXISTS igo_favorites (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    library_id INTEGER NOT NULL,
    seat_key VARCHAR(64) NOT NULL,
    seat_name VARCHAR(128) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_igo_favorites_user_lib_seat ON igo_favorites (user_id, library_id, seat_key);
CREATE INDEX IF NOT EXISTS idx_igo_favorites_user_lib ON igo_favorites (user_id, library_id);

CREATE TABLE IF NOT EXISTS igo_seat_labels (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    library_id INTEGER NOT NULL,
    seat_key VARCHAR(64) NOT NULL,
    seat_name VARCHAR(128) NOT NULL DEFAULT '',
    label_text VARCHAR(64) NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_igo_seat_labels_user_lib_seat ON igo_seat_labels (user_id, library_id, seat_key);
CREATE INDEX IF NOT EXISTS idx_igo_seat_labels_user_lib ON igo_seat_labels (user_id, library_id);

CREATE TABLE IF NOT EXISTS igo_protocol_overrides (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    overrides JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_igo_protocol_overrides_user ON igo_protocol_overrides (user_id);

CREATE TABLE IF NOT EXISTS igo_settings (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_igo_settings_user ON igo_settings (user_id);

CREATE TABLE IF NOT EXISTS igo_task_runs (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    kind VARCHAR(32) NOT NULL,
    state VARCHAR(32) NOT NULL,
    title VARCHAR(128) NOT NULL DEFAULT '',
    message TEXT NOT NULL DEFAULT '',
    plan_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    started_at TIMESTAMPTZ,
    last_updated_at TIMESTAMPTZ,
    last_request_at TIMESTAMPTZ,
    poll_count INTEGER NOT NULL DEFAULT 0,
    request_count INTEGER NOT NULL DEFAULT 0,
    reason VARCHAR(64) NOT NULL DEFAULT ''
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_igo_task_runs_user_kind ON igo_task_runs (user_id, kind);

CREATE TABLE IF NOT EXISTS igo_task_launch_history (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    record_id VARCHAR(64) NOT NULL,
    kind VARCHAR(32) NOT NULL,
    fingerprint VARCHAR(128) NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL,
    payload_json JSONB NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_igo_task_history_record ON igo_task_launch_history (user_id, record_id);
CREATE UNIQUE INDEX IF NOT EXISTS uq_igo_task_history_fingerprint ON igo_task_launch_history (user_id, kind, fingerprint);
CREATE INDEX IF NOT EXISTS idx_igo_task_history_user_kind ON igo_task_launch_history (user_id, kind);

CREATE TABLE IF NOT EXISTS igo_global_leak_targets (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    library_id INTEGER NOT NULL,
    library_name VARCHAR(255) NOT NULL DEFAULT '',
    floor VARCHAR(64) NOT NULL DEFAULT '',
    scan_priority INTEGER NOT NULL DEFAULT 0
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_igo_leak_targets_user_lib ON igo_global_leak_targets (user_id, library_id);
CREATE INDEX IF NOT EXISTS idx_igo_leak_targets_user_prio ON igo_global_leak_targets (user_id, scan_priority);

CREATE TABLE IF NOT EXISTS igo_global_leak_blacklist (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    library_id INTEGER NOT NULL,
    seat_key VARCHAR(64) NOT NULL,
    seat_name VARCHAR(128) NOT NULL DEFAULT ''
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_igo_leak_blacklist_user_lib_seat ON igo_global_leak_blacklist (user_id, library_id, seat_key);
CREATE INDEX IF NOT EXISTS idx_igo_leak_blacklist_user_lib ON igo_global_leak_blacklist (user_id, library_id);

CREATE TABLE IF NOT EXISTS igo_checkin_sessions (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    token TEXT NOT NULL,
    saved_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ,
    can_auto_restore BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_igo_checkin_sessions_user ON igo_checkin_sessions (user_id);

CREATE TABLE IF NOT EXISTS igo_dashboard_metrics (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    historical_success_count INTEGER NOT NULL DEFAULT 0,
    total_guard_seconds BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_igo_dashboard_metrics_user ON igo_dashboard_metrics (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS igo_dashboard_metrics;
DROP TABLE IF EXISTS igo_checkin_sessions;
DROP TABLE IF EXISTS igo_global_leak_blacklist;
DROP TABLE IF EXISTS igo_global_leak_targets;
DROP TABLE IF EXISTS igo_task_launch_history;
DROP TABLE IF EXISTS igo_task_runs;
DROP TABLE IF EXISTS igo_settings;
DROP TABLE IF EXISTS igo_protocol_overrides;
DROP TABLE IF EXISTS igo_seat_labels;
DROP TABLE IF EXISTS igo_favorites;
DROP TABLE IF EXISTS igo_venues;
DROP TABLE IF EXISTS igo_sessions;
-- +goose StatementEnd
