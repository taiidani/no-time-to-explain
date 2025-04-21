package cron

import (
	"context"
	"encoding/json"
	"os"

	"github.com/taiidani/no-time-to-explain/internal/destiny"
)

func Refresh(ctx context.Context, client *destiny.Client) error {
	helper := destiny.NewHelper(client)

	_, _, err := helper.GetClanFish(ctx)
	if err != nil {
		return err
	}

	titles, err := helper.GetClanTitles(ctx)
	if err != nil {
		return err
	}

	_ = json.NewEncoder(os.Stdout).Encode(titles)
	return nil
}
