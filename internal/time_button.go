package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const (
	changeTimeCustomID      string = "time-btn"
	changeTimeModalCustomID string = "time-modal"
)

func changeTimeHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.MessageComponentData()
	optsJson := strings.TrimPrefix(data.CustomID, changeTimeCustomID)

	opts := &interactionState{}
	if err := json.Unmarshal([]byte(optsJson), opts); err != nil {
		errorMessage(s, i.Interaction, fmt.Errorf("could not decode timestamp data: %w", err))
		return
	}

	customID := strings.Builder{}
	if err := json.NewEncoder(&customID).Encode(opts); err != nil {
		errorMessage(s, i.Interaction, fmt.Errorf("could not encode timestamp data: %w", err))
		return
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: changeTimeModalCustomID + customID.String(),
			Title:    "Change Time",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "txt-date",
							Style:       discordgo.TextInputShort,
							Required:    true,
							Label:       "Date",
							Value:       opts.Date,
							Placeholder: "YYYYMMDD",
							MinLength:   8,
							MaxLength:   10,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "txt-time",
							Style:       discordgo.TextInputShort,
							Required:    true,
							Label:       "Time",
							Value:       opts.Time,
							Placeholder: "HH:MM:SS",
							MinLength:   5,
							MaxLength:   8,
						},
					},
				},
			},
		},
	})
	if err != nil {
		log.Println("Could not respond to user interaction:", err)
		commandError(s, i.Interaction, err)
		return
	}
}

func changeTimeSubmitHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ModalSubmitData()
	optsJson := strings.TrimPrefix(data.CustomID, changeTimeModalCustomID)

	opts := &interactionState{}
	if err := json.Unmarshal([]byte(optsJson), opts); err != nil {
		errorMessage(s, i.Interaction, fmt.Errorf("could not decode timestamp data: %w", err))
		return
	}
	opts.Date = data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	opts.Time = data.Components[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

	msg, err := responseMessage(*opts)
	if err != nil {
		errorMessage(s, i.Interaction, err)
		return
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: msg,
	})
	if err != nil {
		log.Println("Could not respond to user button submission:", err)
		commandError(s, i.Interaction, err)
		return
	}
}
