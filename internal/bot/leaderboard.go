package bot

import (
	"context"
	"fmt"
	"log"

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

func leaderboardFish(ctx context.Context) (*discordgo.InteractionResponseData, error) {
	helper := destiny.NewHelper(destinyClient)

	def, metric, err := helper.GetClanFish(ctx)
	if err != nil {
		return nil, err
	}

	fields := []*discordgo.MessageEmbedField{
		{
			Name:   def.DisplayProperties.Name,
			Value:  fmt.Sprintf("%d", metric.TotalFish),
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
