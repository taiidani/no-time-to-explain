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
	hub.Scope().SetTags(map[string]string{
		"channel-id": m.ChannelID,
		"trigger":    m.Content,
	})

	// Determine the response based on the given content
	messages, err := models.LoadMessages(ctx)
	if err != nil {
		hub.CaptureException(err)
		slog.Error("Could not get messages from DB", "err", err)
	}

	ref := m.Message.Reference()
	response := c.responseForTrigger(ctx, messages, m.Content)
	if response != "" {
		log = log.With("response", response)

		var err error
		if ref != nil {
			log.Info("Sending message reply")
			_, err = s.ChannelMessageSendReply(m.ChannelID, response, ref)
		} else {
			log.Info("Sending message")
			_, err = s.ChannelMessageSend(m.ChannelID, response)
		}

		if err != nil {
			sentry.GetHubFromContext(ctx).CaptureException(err)
			log.Error("Could not send channel response", "err", err)
		}
	}
}

func (c *Commands) responseForTrigger(ctx context.Context, messages []models.Message, input string) string {
	for _, message := range messages {
		re := regexp.MustCompile(message.Trigger)
		if re.MatchString(input) {
			return message.Response
		}
	}

	return ""
}
