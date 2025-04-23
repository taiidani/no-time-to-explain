-- +goose Up
-- +goose StatementBegin
CREATE TABLE feed (
    id SERIAL PRIMARY KEY,
    source VARCHAR(255) NOT NULL,
    author_source_id VARCHAR(255) NOT NULL,
    author VARCHAR(255) NOT NULL,
    last_message TIMESTAMP NOT NULL,
    CONSTRAINT unique_source_author UNIQUE (source, author)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE feed;
-- +goose StatementEnd
