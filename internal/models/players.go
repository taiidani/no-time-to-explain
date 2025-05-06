package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Player struct {
	ID                int
	DisplayName       string
	MembershipType    int
	MembershipId      string
	GlobalDisplayName string
	GlobalDisplayCode int
	GroupId           string
	GroupJoinDate     time.Time
	LastOnline        time.Time

	// The last time this player record was updated by the refresh script
	// If this is old then the player has likely left the clan
	UpdatedAt time.Time

	// This is the first time the record was created, when the refresh
	// script discovered them.
	CreatedAt time.Time
}

func (p *Player) Validate() error {
	return nil
}

func GetPlayers(ctx context.Context) ([]Player, error) {
	rows, err := db.QueryContext(ctx, `
SELECT
	id,
	display_name,
	membership_type,
	global_display_name,
	global_display_code,
	group_id,
	membership_id,
	last_online,
	group_join_date,
	updated_at,
	created_at
FROM player
`)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}

	ret := []Player{}
	for rows.Next() {
		add := Player{}
		err = rows.Scan(&add.ID, &add.DisplayName, &add.MembershipType, &add.GlobalDisplayName,
			&add.GlobalDisplayCode, &add.GroupId, &add.MembershipId,
			&add.LastOnline, &add.GroupJoinDate, &add.UpdatedAt, &add.CreatedAt,
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

func BulkUpdatePlayers(ctx context.Context, players []Player) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	updateTimestamp := time.Now()

	tableName := fmt.Sprintf("player_incoming_%d", updateTimestamp.Unix())
	_, err = tx.ExecContext(ctx, `
CREATE TEMPORARY TABLE `+tableName+`
ON COMMIT DROP
AS TABLE player
WITH NO DATA`)
	if err != nil {
		return errors.Join(tx.Rollback(), fmt.Errorf("creating temp table %q: %w", tableName, err))
	}

	// TODO Can we optimize this?
	for _, player := range players {
		_, err = tx.ExecContext(ctx, `
INSERT INTO `+tableName+`
(
	display_name,
	membership_type,
	global_display_name,
	global_display_code,
	group_id,
	membership_id,
	last_online,
	group_join_date,
	updated_at,
	created_at
)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
			player.DisplayName, player.MembershipType, player.GlobalDisplayName,
			player.GlobalDisplayCode, player.GroupId, player.MembershipId,
			player.LastOnline, player.GroupJoinDate, updateTimestamp, updateTimestamp,
		)
		if err != nil {
			return errors.Join(tx.Rollback(), fmt.Errorf("adding player %q: %w", player.MembershipId, err))
		}
	}

	// Note - we're ignoring `created_at` when matched so that it continues to represent
	// the first time this row was inserted.
	_, err = tx.ExecContext(ctx, `
MERGE INTO player AS p
USING `+tableName+` AS i
ON (i.membership_id = p.membership_id)
WHEN MATCHED THEN
    UPDATE SET
		display_name = i.display_name,
		membership_type = i.membership_type,
		global_display_name = i.global_display_name,
		global_display_code = i.global_display_code,
		group_id = i.group_id,
		membership_id = i.membership_id,
		last_online = i.last_online,
		group_join_date = i.group_join_date,
		updated_at = i.updated_at
WHEN NOT MATCHED THEN
    INSERT (
		display_name,
		membership_type,
		global_display_name,
		global_display_code,
		group_id,
		membership_id,
		last_online,
		group_join_date,
		updated_at,
		created_at
	) VALUES (
		i.display_name,
		i.membership_type,
		i.global_display_name,
		i.global_display_code,
		i.group_id,
		i.membership_id,
		i.last_online,
		i.group_join_date,
		i.updated_at,
		i.created_at
	)
`)
	if err != nil {
		return errors.Join(tx.Rollback(), fmt.Errorf("merge failure: %w", err))
	}

	return tx.Commit()
}
