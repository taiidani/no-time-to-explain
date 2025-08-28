package bot

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/getsentry/sentry-go"
	"github.com/taiidani/no-time-to-explain/internal/destiny"
)

func leaderboardHandler(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	span := sentry.StartSpan(ctx, "leaderboard")
	defer span.Finish()

	var msg *discordgo.InteractionResponseData
	var err error
	switch i.ApplicationCommandData().Options[0].Name {
	case "fish":
		msg, err = leaderboardFish(ctx)
		if err != nil {
			errorMessage(s, i.Interaction, err)
			return
		}
	default:
		slog.Warn("Unknown command provided", "err", err)
		commandError(s, i.Interaction, err)
		return
	}

	if msg != nil {
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: msg,
		})
		if err != nil {
			slog.Warn("Could not respond to user message", "err", err)
			commandError(s, i.Interaction, err)
			return
		}
	}
}

// Number of fish MetricDefinition index: 24768693
// Number of fish objective hash: 2773717662
func leaderboardFish(ctx context.Context) (*discordgo.InteractionResponseData, error) {
	span := sentry.StartSpan(ctx, "fish")
	defer span.Finish()

	helper := destiny.NewHelper(destinyClient)

	name, metrics, err := helper.GetFishMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get fish metrics: %w", err)
	}

	var totalFish int32
	for _, metric := range metrics {
		if metric.Progress != nil {
			totalFish += *metric.Progress
		}
	}

	fields := []*discordgo.MessageEmbedField{
		{
			Name:   name,
			Value:  fmt.Sprintf("%d", totalFish),
			Inline: true,
		},
	}

	ret := &discordgo.InteractionResponseData{
		Embeds: []*discordgo.MessageEmbed{
			{
				Title:       "Unknown Space Fish Leaderboard",
				Color:       defaultColor,
				Description: "Statistics about the clan's fishing habit",
				Fields:      fields,
				Footer:      &discordgo.MessageEmbedFooter{Text: defaultFooter},
			},
		},
		Flags: discordgo.MessageFlagsEphemeral,
	}

	return ret, nil
}
