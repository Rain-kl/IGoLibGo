-- +goose Up
-- +goose StatementBegin
ALTER TABLE w_access_tokens ADD COLUMN IF NOT EXISTS token VARCHAR(255) NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE w_access_tokens DROP COLUMN IF EXISTS token;
-- +goose StatementEnd
