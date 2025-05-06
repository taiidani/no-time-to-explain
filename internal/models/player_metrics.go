package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type PlayerMetric struct {
	ID              int
	MetricID        uint32
	PlayerID        int
	ObjectiveHash   uint32
	Progress        *int32
	CompletionValue int32
	Complete        bool
	Visible         bool

	// This will be populated with the time of the refresh scan where the player's
	// completion flipped from 0 to 1
	CompletedAt *time.Time

	// The last time this player record was updated by the refresh script
	// If this is old then the player has likely left the clan
	UpdatedAt time.Time

	// This is the first time the record was created, when the refresh
	// script discovered them.
	CreatedAt time.Time
}

func (m *PlayerMetric) Validate() error {
	return nil
}

func GetPlayerMetrics(ctx context.Context, playerID string) ([]PlayerMetric, error) {
	rows, err := db.QueryContext(ctx, `
SELECT
	id,
	player_id,
	metric_id,
    objective_hash,
	progress,
	completion_value,
	complete,
	completed_at,
	visible,
	updated_at,
	created_at
FROM player_metric
WHERE player_id = $1
`, playerID)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}

	ret := []PlayerMetric{}
	for rows.Next() {
		add := PlayerMetric{}
		err = rows.Scan(add.ID, add.PlayerID, add.MetricID,
			add.ObjectiveHash, add.Progress, add.CompletionValue,
			add.Complete, add.CompletedAt, add.Visible,
			add.UpdatedAt, add.CreatedAt,
		)
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("data not found")
		} else if err != nil {
			return nil, fmt.Errorf("error accessing row: %w", err)
		}

		ret = append(ret, add)
	}

	return ret, nil
}

func GetPlayerMetric(ctx context.Context, playerID int, metricID string) (*PlayerMetric, error) {
	row := db.QueryRowContext(ctx, `
SELECT
	id,
	player_id,
	metric_id,
    objective_hash,
	progress,
	completion_value,
	complete,
	completed_at,
	visible,
	updated_at,
	created_at
FROM player_metric
WHERE player_id = $1 AND metric_id = $2
`, playerID, metricID)

	ret := PlayerMetric{}
	err := row.Scan(ret.ID, ret.PlayerID, ret.MetricID,
		ret.ObjectiveHash, ret.Progress, ret.CompletionValue,
		ret.Complete, ret.CompletedAt, ret.Visible,
		ret.UpdatedAt, ret.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("data not found")
	} else if err != nil {
		return nil, fmt.Errorf("error accessing row: %w", err)
	}

	return &ret, nil
}

func BulkUpdateMetrics(ctx context.Context, playerMetrics []PlayerMetric) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	updateTimestamp := time.Now()

	tableName := fmt.Sprintf("player_metric_incoming_%d", updateTimestamp.Unix())
	_, err = tx.ExecContext(ctx, `
CREATE TEMPORARY TABLE `+tableName+`
ON COMMIT DROP
AS TABLE player_metric
WITH NO DATA`)
	if err != nil {
		return errors.Join(tx.Rollback(), fmt.Errorf("creating temp table %q: %w", tableName, err))
	}

	// TODO Can we optimize this?
	for _, metric := range playerMetrics {
		_, err = tx.ExecContext(ctx, `
INSERT INTO `+tableName+`
(
    player_id,
	metric_id,
    objective_hash,
	progress,
	completion_value,
	complete,
	visible,
	updated_at,
	created_at
)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
			metric.PlayerID, metric.MetricID, metric.ObjectiveHash,
			metric.Progress, metric.CompletionValue, metric.Complete,
			metric.Visible, updateTimestamp, updateTimestamp,
		)
		if err != nil {
			return errors.Join(tx.Rollback(), fmt.Errorf("adding player metric %q: %w", metric.PlayerID, err))
		}
	}

	// Note - we're ignoring `created_at` when matched so that it continues to represent
	// the first time this row was inserted.
	_, err = tx.ExecContext(ctx, `
MERGE INTO player_metric AS p
USING `+tableName+` AS i
ON (i.player_id = p.player_id AND i.metric_id = p.metric_id)
WHEN MATCHED THEN
    UPDATE SET
		player_id = i.player_id,
		metric_id = i.metric_id,
		objective_hash = i.objective_hash,
		progress = i.progress,
		completion_value = i.completion_value,
		complete = i.complete,
		visible = i.visible,
		updated_at = i.updated_at
WHEN NOT MATCHED THEN
    INSERT (
		player_id,
		metric_id,
		objective_hash,
		progress,
		completion_value,
		complete,
		visible,
		updated_at,
		created_at
	) VALUES (
		i.player_id,
		i.metric_id,
		i.objective_hash,
		i.progress,
		i.completion_value,
		i.complete,
		i.visible,
		i.updated_at,
		i.created_at
	)
`)
	if err != nil {
		return errors.Join(tx.Rollback(), fmt.Errorf("merge failure: %w", err))
	}

	// Before committing, determine if any completions occurred in this run
	_, err = tx.ExecContext(ctx, `
UPDATE player_metric SET completed_at = $1
WHERE completed_at IS NULL AND complete IS TRUE`, updateTimestamp)
	if err != nil {
		return errors.Join(tx.Rollback(), fmt.Errorf("completion recording failure: %w", err))
	}

	return tx.Commit()
}
