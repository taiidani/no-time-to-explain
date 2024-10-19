package internal

import (
	"context"
	"log/slog"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/getsentry/sentry-go"
)

var messageResponses = map[string]string{
	"jesus christ": "You mean Bees-us?",
}

func (c *Commands) handleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Set up the Sentry transaction
	transaction := sentry.StartTransaction(context.Background(), "message")
	defer transaction.Finish()
	// ctx := transaction.Context()
	addSentry(m)

	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	log := slog.With("channel-id", m.ChannelID, "user-id", m.Author.ID)
	content := strings.ToLower(m.Content)

	for trigger, response := range messageResponses {
		if strings.Contains(content, trigger) {
			log = log.With("trigger", trigger, "response", response)
			log.Info("Message received")

			_, err := s.ChannelMessageSend(m.ChannelID, response)
			if err != nil {
				log.Error("Could not send channel response", "response", response)
			}
		}
	}
}
