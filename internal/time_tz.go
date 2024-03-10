package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	tzHandlerCustomID string = "time-tz"
	zonePath          string = "/usr/share/zoneinfo/"
)

var zones = []string{}

func init() {
	// The maximum amount of avaiable choices in a Select is 25
	// Limiting scope to only UTC+US
	zones = append(zones, "UTC")
	loadTimezones("")
	log.Println("Available timezones:", strings.Join(zones, ","))
}

func loadTimezones(path string) {
	files, _ := os.ReadDir(filepath.Join(zonePath, path))
	for _, f := range files {
		if f.Name() != strings.ToUpper(f.Name()[:1])+f.Name()[1:] {
			continue
		}
		if f.IsDir() {
			loadTimezones(filepath.Join(path, f.Name()))
		} else {
			zones = append(zones, filepath.Join(path, f.Name()))
		}
	}
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
