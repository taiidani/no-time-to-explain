package internal

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/taiidani/no-time-to-explain/internal/bluesky"
	"github.com/taiidani/no-time-to-explain/internal/destiny"
	"github.com/taiidani/no-time-to-explain/internal/models"
)

func Refresh(ctx context.Context, client *destiny.Client, discord *discordgo.Session) error {
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

		slog.Info("starting destiny api refresh")
		err := refreshDestinyAPI(ctx, client)
		if err != nil {
			slog.Error("destiny refresh error", "err", err)
			return
		}
		slog.Info("destiny api refresh complete")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		slog.Info("starting bluesky refresh")
		err := refreshBlueskyFeeds(ctx, discord)
		if err != nil {
			slog.Error("bluesky refresh error", "err", err)
			return
		}
		slog.Info("bluesky refresh complete")
	}()

	wg.Wait()
	return nil
}

func refreshDestinyAPI(ctx context.Context, client *destiny.Client) error {
	helper := destiny.NewHelper(client)

	_, _, err := helper.GetClanFish(ctx)
	if err != nil {
		return fmt.Errorf("fish error: %w", err)
	}

	return nil
}

func refreshBlueskyFeeds(ctx context.Context, discord *discordgo.Session) error {
	// Examine the Bluesky posts
	bs := bluesky.NewBlueskyClient()

	feeds, err := models.LoadFeeds(ctx)
	if err != nil {
		return fmt.Errorf("feed load error: %w", err)
	}

	for _, feed := range feeds {
		userFeed, err := bs.GetUserFeed(feed.SourceAuthorID)
		if err != nil {
			return fmt.Errorf("user feed error: %w", err)
		}

		for _, post := range userFeed.Feed {
			if post.Post.Record.CreatedAt.Before(feed.LastMessage) || post.Post.Record.CreatedAt == feed.LastMessage {
				continue
			}

			// Got a new one!
			channelID := os.Getenv("BLUESKY_FEED_CHANNEL_ID")
			_, err := discord.ChannelMessageSend(channelID, post.Post.URL(), discordgo.WithContext(ctx))
			if err != nil {
				return fmt.Errorf("posting error: %w", err)
			}

			feed.LastMessage = post.Post.Record.CreatedAt
			err = models.UpdateFeed(ctx, feed)
			if err != nil {
				return fmt.Errorf("failed to update feed %s in db: %w", feed.Author, err)
			}
		}
	}

	return nil
}
