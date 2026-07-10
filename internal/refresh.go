package internal

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/taiidani/no-time-to-explain/internal/bluesky"
	"github.com/taiidani/no-time-to-explain/internal/models"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

// tracer is the OpenTelemetry tracer for the background refresh job.
var tracer = otel.Tracer("github.com/taiidani/no-time-to-explain/internal")

func Refresh(ctx context.Context, discord *discordgo.Session) error {
	ctx, span := tracer.Start(ctx, "refresh")
	defer span.End()

	wg := sync.WaitGroup{}

	wg.Go(func() {
		slog.InfoContext(ctx, "starting bluesky refresh")
		err := refreshBlueskyFeeds(ctx, discord)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			slog.ErrorContext(ctx, "bluesky refresh error", "err", err)
			return
		}
		slog.InfoContext(ctx, "bluesky refresh complete")
	})

	wg.Wait()
	return nil
}

// refreshBlueskyFeeds will post all Bluesky messages since the last processing time
// to the associated Discord channel.
func refreshBlueskyFeeds(ctx context.Context, discord *discordgo.Session) (err error) {
	ctx, span := tracer.Start(ctx, "refresh-bluesky")
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	// Examine the Bluesky posts
	bs := bluesky.NewBlueskyClient()

	feeds, err := models.LoadFeeds(ctx)
	if err != nil {
		return fmt.Errorf("feed load error: %w", err)
	}

	for _, feed := range feeds {
		logger := slog.With("author", feed.Author)

		userFeed, err := bs.GetUserFeed(feed.SourceAuthorID)
		if err != nil {
			return fmt.Errorf("user feed error: %w", err)
		}

		newPosts := filterPosts(logger, feed, userFeed.Feed)
		if len(newPosts) == 0 {
			logger.Info("no new bluesky posts since the last processing time")
			continue
		}

		// Reverse the order of the posts so they are in chronological order
		slices.Reverse(newPosts)

		channelID := os.Getenv("BLUESKY_FEED_CHANNEL_ID")
		for _, post := range newPosts {
			_, err := discord.ChannelMessageSend(channelID, post.Post.URL(), discordgo.WithContext(ctx))
			if err != nil {
				return fmt.Errorf("posting error: %w", err)
			}

			// Mark this as the most recent feed entry we've processed
			if post.Post.IndexedAt.After(feed.LastMessage) {
				feed.LastMessage = post.Post.IndexedAt
			}
		}

		// Record the most recent post into the DB for the next run
		err = models.UpdateFeed(ctx, feed)
		if err != nil {
			return fmt.Errorf("failed to update feed %s in db: %w", feed.Author, err)
		}
	}

	return nil
}

func filterPosts(logger *slog.Logger, feed models.Feed, posts []bluesky.FeedPostEntry) []bluesky.FeedPostEntry {
	ret := []bluesky.FeedPostEntry{}

	for _, post := range posts {
		postLogger := logger.With(
			"post_uri", post.Post.URI,
			"post_indexed_at", post.Post.IndexedAt,
		)

		// Skip reposts. The author feed can include reposted entries, which causes the
		// same underlying post to be emitted again on later refreshes.
		if post.Reason != nil && post.Reason.Type == "app.bsky.feed.defs#reasonRepost" {
			postLogger.Debug("skipping bluesky repost entry",
				"reason_indexed_at", post.Reason.IndexedAt,
			)
			continue
		}

		// Skip posts from within the last minute
		// If we display before the embeds have been processed then they might not get
		// added to the message.
		if post.Post.IndexedAt.After(time.Now().Add(time.Minute * -1)) {
			postLogger.Debug("skipping recent bluesky post")
			continue
		}

		// Has the feed entry already been processed?
		if post.Post.IndexedAt.Before(feed.LastMessage) || post.Post.IndexedAt.Equal(feed.LastMessage) {
			postLogger.Debug("skipping already processed bluesky post",
				"last_message", feed.LastMessage,
			)
			continue
		}

		ret = append(ret, post)
	}

	return ret
}
