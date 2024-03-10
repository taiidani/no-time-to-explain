package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	tzHandlerCustomID string = "time-tz"
)

var zones = []string{}

func init() {
	candidateZones := []string{
		"UTC",
		"US/Alaska",
		"US/Aleutian",
		"US/Arizona",
		"US/Central",
		"US/East-Indiana",
		"US/Eastern",
		"US/Hawaii",
		"US/Indiana-Starke",
		"US/Michigan",
		"US/Mountain",
		"US/Pacific",
		"US/Samoa",
	}

	for _, zone := range candidateZones {
		if _, err := time.LoadLocation(zone); err == nil {
			zones = append(zones, zone)
		}
	}

	if len(zones) > 25 {
		log.Println("Truncating available timezones to 25 due to Discord Select choice limitations")
		zones = zones[0:25]
	}
	log.Println("Loaded timezones:", strings.Join(zones, ", "))
}

func tzHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.MessageComponentData()
	optsJson := strings.TrimPrefix(data.CustomID, tzHandlerCustomID)

	opts := &interactionState{}
	if err := json.Unmarshal([]byte(optsJson), opts); err != nil {
		errorMessage(s, i.Interaction, fmt.Errorf("could not decode timestamp data: %w", err))
		return
	}
	opts.TZ = data.Values[0]

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
		log.Println("Could not respond to user interaction:", err)
		commandError(s, i.Interaction, err)
		return
	}
}

func tzOptions(def string) []discordgo.SelectMenuOption {
	ret := []discordgo.SelectMenuOption{}

	for _, zone := range zones {
		if _, err := time.LoadLocation(zone); err == nil {
			ret = append(ret, discordgo.SelectMenuOption{
				Label:   zone,
				Value:   zone,
				Default: def == zone,
			})
		}
	}

	return ret
}
