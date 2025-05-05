-- +goose Up
-- +goose StatementBegin
CREATE TABLE player (
    id SERIAL PRIMARY KEY,
    display_name VARCHAR(255) NOT NULL,
    membership_type SMALLINT NOT NULL,
    global_display_name VARCHAR(255) NOT NULL,
    global_display_code INT NOT NULL,
    group_id VARCHAR(255) NOT NULL,
    group_join_date TIMESTAMP NOT NULL,
    membership_id VARCHAR(255) NOT NULL UNIQUE,
    last_online TIMESTAMP NOT NULL,
    last_updated TIMESTAMP NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE player;
-- +goose StatementEnd
