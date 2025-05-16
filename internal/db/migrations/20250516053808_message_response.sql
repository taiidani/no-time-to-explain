-- +goose Up
-- +goose StatementBegin
ALTER TABLE message DROP CONSTRAINT message_trigger_key;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE message ADD CONSTRAINT message_trigger_key UNIQUE (trigger);
-- +goose StatementEnd
