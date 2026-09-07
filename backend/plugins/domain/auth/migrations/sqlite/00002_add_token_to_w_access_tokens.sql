-- +goose Up
-- +goose StatementBegin
ALTER TABLE w_access_tokens ADD COLUMN token TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE w_access_tokens DROP COLUMN token;
-- +goose StatementEnd
