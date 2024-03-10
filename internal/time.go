package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

type interactionState struct {
	Date string
	Time string
	TZ   string
}

func timeHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	opts := parseOptions(i.ApplicationCommandData().Options)

	msg, err := responseMessage(opts)
	if err != nil {
		errorMessage(s, i.Interaction, err)
		return
	}

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

func responseMessage(opts interactionState) (*discordgo.InteractionResponseData, error) {
	customID := strings.Builder{}
	if err := json.NewEncoder(&customID).Encode(opts); err != nil {
		return nil, fmt.Errorf("could not encode timestamp data: %w", err)
	}

	// If the necessary fields have not been provided, display a call to action
	// Otherwise, render the full message
	title := "Timestamp missing data"
	color := 0xFF5050
	description := "Please use the fields below to set your current timezone and desired time."
	fields := []*discordgo.MessageEmbedField{}
	if len(opts.Date) > 0 && len(opts.Time) > 0 && len(opts.TZ) > 0 {
		title = "Timestamp rendered!"
		color = 0x05FF05
		description = ""

		tm, err := parseTimestamp(opts)
		if err != nil {
			return nil, err
		}

		types := []string{"d", "f", "t", "D", "T", "R", "F"}
		for _, typ := range types {
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:   fmt.Sprintf("<t:%d:%s>", tm.Unix(), typ),
				Value:  fmt.Sprintf("```<t:%d:%s>```", tm.Unix(), typ),
				Inline: true,
			})
		}
	}

	ret := &discordgo.InteractionResponseData{
		Embeds: []*discordgo.MessageEmbed{
			{
				Title:       title,
				Color:       color,
				Description: description,
				Fields:      fields,
				Footer:      &discordgo.MessageEmbedFooter{Text: "Written with 💙 for Unknown Space by @taiidani"},
			},
		},
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						CustomID: changeTimeCustomID + customID.String(),
						Label:    "Change Time",
						Style:    discordgo.PrimaryButton,
					},
					discordgo.Button{
						CustomID: "time-now" + customID.String(),
						Label:    "Current Time",
						Style:    discordgo.SecondaryButton,
					},
				},
			},
		},
		Flags: discordgo.MessageFlagsEphemeral,
	}

	return ret, nil
}

func parseOptions(opts []*discordgo.ApplicationCommandInteractionDataOption) interactionState {
	now := time.Now()
	ret := interactionState{
		// January 2, 3:04:05PM, 2006 MST
		Date: now.Format("2006-01-02"),
		Time: now.Format("3:04:05 PM"),
		TZ:   now.Format("MST"),
	}
	for _, opt := range opts {
		switch opt.Name {
		case "date":
			ret.Date = opt.StringValue()
		case "time":
			ret.Time = opt.StringValue()
		case "tz":
			ret.TZ = opt.StringValue()
		}
	}

	return ret
}

func parseTimestamp(opts interactionState) (time.Time, error) {
	return time.Parse("2006-01-02 3:04:05 PM MST", fmt.Sprintf("%s %s %s", opts.Date, opts.Time, opts.TZ))
}

func errorMessage(s *discordgo.Session, i *discordgo.Interaction, msg error) {
	_ = s.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg.Error(),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
