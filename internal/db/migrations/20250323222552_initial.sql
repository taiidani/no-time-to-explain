-- +goose Up
CREATE TABLE message (
    id SERIAL PRIMARY KEY,
    trigger VARCHAR(255) NOT NULL UNIQUE,
    response TEXT NOT NULL
);

-- +goose Down
DROP TABLE message;
