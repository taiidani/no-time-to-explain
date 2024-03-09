package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
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
	tm, err := parseTimestamp(opts)
	if err != nil {
		return nil, err
	}

	customID := strings.Builder{}
	if err := json.NewEncoder(&customID).Encode(opts); err != nil {
		return nil, fmt.Errorf("could not encode timestamp data: %w", err)
	}

	fields := []*discordgo.MessageEmbedField{}
	types := []string{"d", "f", "t", "D", "T", "R", "F"}
	for _, typ := range types {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("<t:%d:%s>", tm.Unix(), typ),
			Value:  fmt.Sprintf("```<t:%d:%s>```", tm.Unix(), typ),
			Inline: true,
		})
	}

	ret := &discordgo.InteractionResponseData{
		Embeds: []*discordgo.MessageEmbed{
			{
				Title:       "Timestamp rendered!",
				Color:       0xFDFDFD,
				Description: "Please change the field below to your local timezone for more accurate results.",
				Fields:      fields,
				Footer:      &discordgo.MessageEmbedFooter{Text: "Written with 💙 for Unknown Space by @taiidani"},
			},
		},
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.SelectMenu{
						MenuType:    discordgo.StringSelectMenu,
						CustomID:    tzHandlerCustomID + customID.String(),
						Options:     tzOptions(opts.TZ),
						Placeholder: "Change Timezone",
					},
				},
			},
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						CustomID: "time-btn" + customID.String(),
						Label:    "Change Time",
						Style:    discordgo.PrimaryButton,
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
		Date: now.Format("2006-01-02"),
		Time: now.Format("15:04:05"),
		TZ:   "UTC",
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
	var err error
	var year int
	var month time.Month
	var day int
	var hour int
	var minute int
	var second int = 0

	d := strings.ReplaceAll(opts.Date, "-", "")
	if len(d) != 8 {
		return time.Time{}, fmt.Errorf("error: %q is not in YYYYMMDD format", d)
	}

	v, err := strconv.ParseInt(d[0:4], 10, 32)
	if err != nil {
		return time.Time{}, fmt.Errorf("error: Unable to parse %s as year", d[0:4])
	}
	year = int(v)

	v, err = strconv.ParseInt(d[4:6], 10, 32)
	if err != nil {
		return time.Time{}, fmt.Errorf("error: Unable to parse %s as month", d[4:6])
	}
	month = time.Month(v)

	v, err = strconv.ParseInt(d[6:8], 10, 32)
	if err != nil {
		return time.Time{}, fmt.Errorf("error: Unable to parse %s as day", d[6:8])
	}
	day = int(v)

	t := strings.Split(opts.Time, ":")
	if len(t) < 2 || len(t) > 3 {
		return time.Time{}, fmt.Errorf("error: %q is not in HH:MM:SS format", t)
	}

	v, err = strconv.ParseInt(t[0], 10, 32)
	if err != nil {
		return time.Time{}, fmt.Errorf("error: Unable to parse %s as hour", t[0])
	}
	hour = int(v)

	v, err = strconv.ParseInt(t[1], 10, 32)
	if err != nil {
		return time.Time{}, fmt.Errorf("error: Unable to parse %s as minute", t[1])
	}
	minute = int(v)

	if len(t) == 3 {
		v, err = strconv.ParseInt(t[2], 10, 32)
		if err != nil {
			return time.Time{}, fmt.Errorf("error: Unable to parse %s as second", t[2])
		}
		second = int(v)
	}

	loc, err := time.LoadLocation(opts.TZ)
	if err != nil {
		return time.Time{}, fmt.Errorf("error: Unable to parse %s as timezone", opts.TZ)
	}

	// January 2, 15:04:05, 2006
	return time.Date(year, month, day, hour, minute, second, 0, loc), nil
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
