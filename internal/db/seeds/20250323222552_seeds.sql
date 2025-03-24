-- +goose Up
-- +goose StatementBegin
DELETE FROM message;
ALTER SEQUENCE message_id_seq RESTART WITH 1;

INSERT INTO message (trigger, response) VALUES
('test', 'Response'),
('ping', 'pong');
-- +goose StatementEnd

-- +goose Down
