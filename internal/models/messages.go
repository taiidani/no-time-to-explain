package models

import "context"

const MessagesDBKey = "messages"

type Messages struct {
	Messages []Message
}

type Message struct {
	ID       string
	Trigger  string
	Response string
}

func LoadMessages(ctx context.Context) ([]Message, error) {
	rows, err := db.QueryContext(ctx, `
SELECT id, trigger, response
FROM message
ORDER BY trigger`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ret := []Message{}
	for rows.Next() {
		var row Message
		if err := rows.Scan(&row.ID, &row.Trigger, &row.Response); err != nil {
			return nil, err
		}

		ret = append(ret, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ret, nil
}

func AddMessage(ctx context.Context, msg Message) error {
	_, err := db.ExecContext(ctx, "INSERT INTO message (trigger, response) VALUES ($1, $2)", msg.Trigger, msg.Response)
	return err
}

func DeleteMessage(ctx context.Context, id string) error {
	_, err := db.ExecContext(ctx, "DELETE FROM message WHERE id = $1", id)
	return err
}
