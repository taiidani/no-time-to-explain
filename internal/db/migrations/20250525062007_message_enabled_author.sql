-- +goose Up
-- +goose StatementBegin
ALTER TABLE message ADD COLUMN created_at TIMESTAMP NOT NULL DEFAULT NOW();
ALTER TABLE message ADD COLUMN enabled BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE message ADD COLUMN sender VARCHAR(255) NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE message DROP COLUMN created_at;
ALTER TABLE message DROP COLUMN enabled;
ALTER TABLE message DROP COLUMN sender;
-- +goose StatementEnd
