package models

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/taiidani/no-time-to-explain/internal/bluesky"
)

type Feed struct {
	ID             string
	Source         string
	Author         string
	SourceAuthorID string
	LastMessage    time.Time
}

func (f *Feed) Validate() error {
	var ret error

	client := bluesky.NewBlueskyClient()
	user, err := client.GetUser(f.Author)
	if err != nil {
		ret = errors.Join(fmt.Errorf("could not look up user %q: %w", f.Author, err))
	} else {
		f.SourceAuthorID = user.DID
	}

	return ret
}

func (f *Feed) URL() string {
	return fmt.Sprintf("https://bsky.app/profile/%s", f.SourceAuthorID)
}

func LoadFeeds(ctx context.Context) ([]Feed, error) {
	rows, err := db.QueryContext(ctx, `
SELECT id, source, author, author_source_id, last_message
FROM feed
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

func AddFeed(ctx context.Context, msg Feed) error {
	_, err := db.ExecContext(ctx, `
INSERT INTO feed (source, author, author_source_id, last_message)
VALUES ($1, $2, $3, $4)`, msg.Source, msg.Author, msg.SourceAuthorID, msg.LastMessage)
	return err
}

func UpdateFeed(ctx context.Context, msg Feed) error {
	_, err := db.ExecContext(ctx, `
UPDATE feed SET
  source = $2,
  author = $3,
  author_source_id = $4,
  last_message = $5
WHERE id = $1`, msg.ID, msg.Source, msg.Author, msg.SourceAuthorID, msg.LastMessage)
	return err
}

func DeleteFeed(ctx context.Context, id string) error {
	_, err := db.ExecContext(ctx, "DELETE FROM feed WHERE id = $1", id)
	return err
}
