-- +goose Up
-- +goose StatementBegin
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

-- +goose Down
-- +goose StatementBegin
DROP TABLE player_metric;
-- +goose StatementEnd
