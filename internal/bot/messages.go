package bot

import (
	"context"
	"log/slog"
	"regexp"

	"github.com/bwmarrin/discordgo"
	"github.com/getsentry/sentry-go"
	"github.com/taiidani/no-time-to-explain/internal/models"
)

func (c *Commands) handleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Set up the Sentry transaction
	hub := sentry.CurrentHub().Clone()
	addSentry(m, hub)
	ctx := sentry.SetHubOnContext(context.Background(), hub)

	transaction := sentry.StartTransaction(ctx, "message")
	defer transaction.Finish()
	ctx = transaction.Context()

	// Ignore all messages created by the bot itself or any other bot
	if m.Author.ID == s.State.User.ID || m.Author.Bot {
		return
	}

	log := slog.With(
		"channel-id", m.ChannelID,
		"user-id", m.Author.ID,
		"trigger", m.Content,
	)

	response := c.responseForTrigger(ctx, m.Content)
	if response != "" {
		log = log.With("response", response)
		log.Info("Message received")

		_, err := s.ChannelMessageSend(m.ChannelID, response)
		if err != nil {
			sentry.GetHubFromContext(ctx).CaptureException(err)
			log.Error("Could not send channel response", "err", err)
		}
	}
}

func (c *Commands) responseForTrigger(ctx context.Context, input string) string {
	var messages models.Messages
	if err := c.db.Get(ctx, models.MessagesDBKey, &messages); err != nil {
		sentry.GetHubFromContext(ctx).CaptureException(err)
		slog.Error("Could not get messages from DB", "err", err)
	}

	if messages.Messages == nil {
		return ""
	}

	for _, message := range messages.Messages {
		re := regexp.MustCompile(message.Trigger)
		if re.MatchString(input) {
			return message.Response
		}
	}

	return ""
}
