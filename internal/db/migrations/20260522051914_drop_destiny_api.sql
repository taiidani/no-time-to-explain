-- +goose Up
-- +goose StatementBegin
DROP TABLE player_metric;
DROP TABLE player;
-- +goose StatementEnd

-- +goose Down
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
CREATE TABLE player_metric (
    id SERIAL PRIMARY KEY,
    player_id INT REFERENCES player,
    metric_id BIGINT NOT NULL,
    objective_hash BIGINT NOT NULL,
	progress INT,
	completion_value INT NOT NULL,
	complete BOOLEAN NOT NULL,
    completed_at TIMESTAMP,
	visible BOOLEAN NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL,
    CONSTRAINT unique_player_id_metric_id UNIQUE (player_id, metric_id)
);
-- +goose StatementEnd
