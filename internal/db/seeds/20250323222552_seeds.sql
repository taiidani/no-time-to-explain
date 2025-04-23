-- +goose Up
-- +goose StatementBegin
DELETE FROM message;
ALTER SEQUENCE message_id_seq RESTART WITH 1;

INSERT INTO message (trigger, response) VALUES
('test', 'Response'),
('ping', 'pong');

DELETE FROM feed;
ALTER SEQUENCE feed_id_seq RESTART WITH 1;

INSERT INTO feed (source, author, author_source_id, last_message) VALUES
('bluesky', 'destinythegame.bungie.net', 'did:plc:lakwqi74b3kcqzk6mpk4kqvy', NOW()),
('bluesky', 'bungieserverstatus.bungie.net', 'did:plc:pekfvt52gjy5qunf3jcdvze4', NOW());

-- +goose StatementEnd

-- +goose Down
