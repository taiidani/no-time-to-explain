package bot

import (
	"context"
	"log/slog"
	"regexp"

	"github.com/bwmarrin/discordgo"
	"github.com/getsentry/sentry-go"
)

var messageResponses = map[string]string{
	`[jJ]esus`: "You mean Bees-us?",
	`^ping$`:   "pong",
}

func (c *Commands) handleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Set up the Sentry transaction
	transaction := sentry.StartTransaction(context.Background(), "message")
	defer transaction.Finish()
	// ctx := transaction.Context()
	addSentry(m)

	// Ignore all messages created by the bot itself or any other bot
	if m.Author.ID == s.State.User.ID || m.Author.Bot {
		return
	}

	log := slog.With(
		"channel-id", m.ChannelID,
		"user-id", m.Author.ID,
		"trigger", m.Content,
	)

	response := responseForTrigger(m.Content)
	if response != "" {
		log = log.With("response", response)
		log.Info("Message received")

		_, err := s.ChannelMessageSend(m.ChannelID, response)
		if err != nil {
			log.Error("Could not send channel response", "err", err)
		}
	}
}

func responseForTrigger(input string) string {
	for trigger, response := range messageResponses {
		re := regexp.MustCompile(trigger)
		if re.MatchString(input) {
			return response
		}
	}

	return ""
}
