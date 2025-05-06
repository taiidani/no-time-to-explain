package bot

import (
	"context"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/getsentry/sentry-go"
	"github.com/taiidani/no-time-to-explain/internal/destiny"
	"github.com/taiidani/no-time-to-explain/internal/models"
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
		log.Println("Unknown command provided:", err)
		commandError(s, i.Interaction, err)
		return
	}

	if msg != nil {
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: msg,
		})
		if err != nil {
			log.Println("Could not respond to user message:", err)
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

	const fishMetricDefinition = "24768693"

	helper := destiny.NewHelper(destinyClient)

	manifest, err := helper.GetManifestMetricEntry(ctx, fishMetricDefinition)
	if err != nil {
		return nil, err
	}

	players, err := models.GetPlayers(ctx)
	if err != nil {
		return nil, fmt.Errorf("get players error: %w", err)
	}

	var totalFish int32
	for _, player := range players {
		metric, err := models.GetPlayerMetric(ctx, player.ID, fishMetricDefinition)
		if err != nil {
			return nil, fmt.Errorf("player %q metric %q error: %w", player.ID, fishMetricDefinition, err)
		}

		if metric.Progress != nil {
			totalFish += *metric.Progress
		}
	}

	fields := []*discordgo.MessageEmbedField{
		{
			Name:   manifest.DisplayProperties.Name,
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
