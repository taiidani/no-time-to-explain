package models

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type Player struct {
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
	LastUpdated time.Time
}

func (p *Player) Validate() error {
	return nil
}

func LoadPlayers(ctx context.Context) ([]Feed, error) {
	rows, err := db.QueryContext(ctx, `
SELECT id, source, author, author_source_id, last_message
FROM player
ORDER BY source, author`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ret := []Feed{}
	for rows.Next() {
		var row Feed
		if err := rows.Scan(&row.ID, &row.Source, &row.Author, &row.SourceAuthorID, &row.LastMessage); err != nil {
			return nil, err
		}

		ret = append(ret, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ret, nil
}

func BulkUpdatePlayers(ctx context.Context, players []Player) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	tableName := fmt.Sprintf("player_incoming_%d", time.Now().Unix())
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
	last_updated
)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NOW())`,
			player.DisplayName, player.MembershipType, player.GlobalDisplayName,
			player.GlobalDisplayCode, player.GroupId, player.MembershipId,
			player.LastOnline, player.GroupJoinDate,
		)
		if err != nil {
			return errors.Join(tx.Rollback(), fmt.Errorf("adding player %q: %w", player.MembershipId, err))
		}
	}

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
		last_updated = i.last_updated
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
		last_updated
	) VALUES (
		i.display_name,
		i.membership_type,
		i.global_display_name,
		i.global_display_code,
		i.group_id,
		i.membership_id,
		i.last_online,
		i.group_join_date,
		i.last_updated
	)
`)
	if err != nil {
		return errors.Join(tx.Rollback(), fmt.Errorf("merge failure: %w", err))
	}

	return tx.Commit()
}
