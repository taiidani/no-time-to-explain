package models

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type Message struct {
	ID        string
	Enabled   bool
	Sender    string
	Trigger   string
	Response  string
	CreatedAt time.Time
}

func (m *Message) Validate() error {
	var ret error

	if len(m.Trigger) < 4 || len(m.Response) < 4 {
		ret = errors.Join(ret, fmt.Errorf("provided inputs need to be at least 4 characters"))
	}
	return ret
}

func LoadMessages(ctx context.Context) ([]Message, error) {
	rows, err := db.QueryContext(ctx, `
SELECT id, enabled, sender, trigger, response, created_at
FROM message
ORDER BY trigger, response`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ret := []Message{}
	for rows.Next() {
		var row Message
		if err := rows.Scan(&row.ID, &row.Enabled, &row.Sender, &row.Trigger, &row.Response, &row.CreatedAt); err != nil {
			return nil, err
		}

		ret = append(ret, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ret, nil
}

func GetMessage(ctx context.Context, id int) (*Message, error) {
	row := db.QueryRowContext(ctx, `
SELECT id, enabled, sender, trigger, response, created_at
FROM message
WHERE id = $1`, id)

	var ret Message
	if err := row.Scan(&ret.ID, &ret.Enabled, &ret.Sender, &ret.Trigger, &ret.Response, &ret.CreatedAt); err != nil {
		return nil, err
	}

	return &ret, nil
}

func AddMessage(ctx context.Context, msg Message) error {
	_, err := db.ExecContext(ctx, `
INSERT INTO message (enabled, sender, trigger, response)
VALUES ($1, $2, $3, $4)
`, msg.Enabled, msg.Sender, msg.Trigger, msg.Response)
	return err
}

func UpdateMessage(ctx context.Context, msg Message) error {
	if msg.ID == "" {
		return fmt.Errorf("cannot update a message without an id")
	}

	_, err := db.ExecContext(ctx, `UPDATE message SET
enabled = $2,
sender = $3,
trigger = $4,
response = $5
WHERE id = $1`, msg.ID, msg.Enabled, msg.Sender, msg.Trigger, msg.Response)
	return err
}

func DeleteMessage(ctx context.Context, id string) error {
	_, err := db.ExecContext(ctx, "DELETE FROM message WHERE id = $1", id)
	return err
}
