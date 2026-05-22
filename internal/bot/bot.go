package bot

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

func errorMessage(s *discordgo.Session, i *discordgo.Interaction, msg error) {
	first := msg.Error()[0:1]
	rest := msg.Error()[1:]
	content := strings.ToUpper(first) + rest

	_ = s.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
			Embeds: []*discordgo.MessageEmbed{
				{
					Description: content,
					Color:       defaultErrorColor,
					Footer:      &discordgo.MessageEmbedFooter{Text: "For support, reach out to @taiidani"},
				},
			},
		},
	})
}
