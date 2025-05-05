package internal

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/getsentry/sentry-go"
	"github.com/taiidani/no-time-to-explain/internal/bluesky"
	"github.com/taiidani/no-time-to-explain/internal/destiny"
	"github.com/taiidani/no-time-to-explain/internal/models"
)

func Refresh(ctx context.Context, client *destiny.Client, discord *discordgo.Session) error {
	span := sentry.StartSpan(ctx, "refresh")
	defer span.Finish()

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
	span := sentry.StartSpan(ctx, "refresh-destiny")
	defer span.Finish()

	helper := destiny.NewHelper(client)

	// First, grab the latest information for all players
	err := refreshDestinyPlayerData(ctx, client)
	if err != nil {
		return fmt.Errorf("players error: %w", err)
	}

	// Next get fishy with it
	err = refreshDestinyPlayerFishData(ctx, helper)
	if err != nil {
		return fmt.Errorf("players error: %w", err)
	}

	return nil
}

func refreshDestinyPlayerData(ctx context.Context, client *destiny.Client) error {
	span := sentry.StartSpan(ctx, "refresh-destiny-titles")
	defer span.Finish()

	helper := destiny.NewHelper(client)

	members, err := helper.GetClan(ctx, destiny.UnknownSpaceGroupID)
	if err != nil {
		return fmt.Errorf("unable to get clan: %w", err)
	}

	return models.BulkUpdatePlayers(ctx, members.Members)
}

func refreshDestinyPlayerFishData(ctx context.Context, helper *destiny.Helper) error {
	span := sentry.StartSpan(ctx, "refresh-destiny-fish")
	defer span.Finish()

	_, _, err := helper.GetClanFish(ctx)
	if err != nil {
		return fmt.Errorf("fish error: %w", err)
	}

	return nil
}

// refreshBlueskyFeeds will post all Bluesky messages since the last processing time
// to the associated Discord channel.
func refreshBlueskyFeeds(ctx context.Context, discord *discordgo.Session) error {
	span := sentry.StartSpan(ctx, "refresh-bluesky")
	defer span.Finish()

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

		newPosts := filterPosts(feed, userFeed.Feed)
		if len(newPosts) == 0 {
			slog.Info("no new bluesky posts since the last processing time", "author", feed.Author)
			continue
		}

		latestPostTime := feed.LastMessage
		for _, post := range newPosts {
			channelID := os.Getenv("BLUESKY_FEED_CHANNEL_ID")
			_, err := discord.ChannelMessageSend(channelID, post.Post.URL(), discordgo.WithContext(ctx))
			if err != nil {
				return fmt.Errorf("posting error: %w", err)
			}

			// Mark this as the most recent post we've processed
			if post.Post.Record.CreatedAt.After(latestPostTime) {
				latestPostTime = post.Post.Record.CreatedAt
			}
		}

		// Record the most recent post into the DB for the next run
		feed.LastMessage = latestPostTime
		err = models.UpdateFeed(ctx, feed)
		if err != nil {
			return fmt.Errorf("failed to update feed %s in db: %w", feed.Author, err)
		}
	}

	return nil
}

func filterPosts(feed models.Feed, posts []bluesky.FeedPostEntry) []bluesky.FeedPostEntry {
	ret := []bluesky.FeedPostEntry{}

	for _, post := range posts {
		// Skip posts from within the last minute
		// If we display before the embeds have been processed then they might not get
		// added to the message.
		if post.Post.Record.CreatedAt.After(time.Now().Add(time.Minute * -1)) {
			continue
		}

		// Has the post already been processed?
		if post.Post.Record.CreatedAt.Before(feed.LastMessage) || post.Post.Record.CreatedAt == feed.LastMessage {
			continue
		}

		ret = append(ret, post)
	}

	return ret
}
