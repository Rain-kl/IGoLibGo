-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS w_custom_greetings (
    id          BIGSERIAL PRIMARY KEY,
    recipient   VARCHAR(64) NOT NULL,
    message     VARCHAR(255) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_w_custom_greetings_recipient ON w_custom_greetings(recipient);

-- 允许跨插件向共享表插入定制化业务配置 (DML INSERT)
-- INSERT INTO w_settings (key, value, created_at, updated_at)
-- VALUES ('custom_example.welcome_message', '"Hello from custom plugin"', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
-- ON CONFLICT (key) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS w_custom_greetings;
-- +goose StatementEnd
