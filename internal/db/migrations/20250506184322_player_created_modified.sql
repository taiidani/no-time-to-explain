-- +goose Up
-- +goose StatementBegin
ALTER TABLE player RENAME COLUMN last_updated TO updated_at;
ALTER TABLE player ADD COLUMN created_at TIMESTAMP NOT NULL DEFAULT NOW();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE player DROP COLUMN created_at;
ALTER TABLE player RENAME COLUMN updated_at TO last_updated;
-- +goose StatementEnd
