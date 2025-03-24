package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/getsentry/sentry-go"
)

type interactionState struct {
	Date string
	Time string
	TZ   string
}

var defaultTimezone *time.Location = time.UTC

func init() {
	if cmdTz, found := os.LookupEnv("CMD_TZ"); found {
		tm, err := parseTimezone(cmdTz)
		if err != nil {
			panic(fmt.Errorf("timezone %q: %w", cmdTz, err))
		}

		defaultTimezone = tm
		slog.Info("Default timezone", "tz", defaultTimezone)
	}
}

func timeHandler(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	span := sentry.StartSpan(ctx, "function")
	defer span.Finish()

	opts, err := parseOptions(ctx, i)
	if err != nil {
		errorMessage(s, i.Interaction, err)
		return
	}

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
	color := defaultErrorColor
	description := "Please use the fields below to set your current timezone and desired time."
	content := ""
	fields := []*discordgo.MessageEmbedField{}
	if len(opts.Date) > 0 && len(opts.Time) > 0 && len(opts.TZ) > 0 {
		title = "Timestamp rendered!"
		color = defaultColor

		tm, err := parseTimestamp(opts)
		if err != nil {
			return nil, err
		}
		slog.Info("Timestamp generated", "time", tm, "opts", opts)

		content = fmt.Sprintf("<t:%d:f>", tm.Unix())
		description = `Click to copy the below text fields. When used in a message, these values will display the time in the reader's local timezone in a format based on the example shown.

Mobile users may have difficulty clicking to copy the fields. Long tap this message and select "Copy Text" to copy the ` + content + ` format to your clipboard. You may change the "f" in the text to modify the format according to the fields below.`

		types := []string{"d", "f", "t", "D", "T", "R", "F"}
		for _, typ := range types {
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:   fmt.Sprintf("<t:%d:%s>", tm.Unix(), typ),
				Value:  fmt.Sprintf("```<t:%d:%s>```", tm.Unix(), typ),
				Inline: true,
			})
		}

		// Special case for LFG Bot
		// Standard Go parsing format: January 2, 3:04:05PM, 2006 MST
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "<#614104443797110794>",
			Value:  fmt.Sprintf("```%s```", tm.Format("01/02/06 3:04PM MST")),
			Inline: true,
		})
	}

	ret := &discordgo.InteractionResponseData{
		Content: content,
		Embeds: []*discordgo.MessageEmbed{
			{
				Title:       title,
				Color:       color,
				Description: description,
				Fields:      fields,
				Footer:      &discordgo.MessageEmbedFooter{Text: defaultFooter},
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

func parseOptions(ctx context.Context, i *discordgo.InteractionCreate) (interactionState, error) {
	tz := defaultTimezone
	var st state
	key := generateStateKey(i)
	if _, err := cache.Get(ctx, key, &st); err == nil {
		if tz, err = parseTimezone(st.TZ); err != nil {
			slog.Warn("Unable to parse timezone", "tz", st.TZ, "key", key, "err", err)
			tz = defaultTimezone
		}
	}

	now := time.Now().In(tz)
	ret := interactionState{
		// Standard Go parsing format: January 2, 3:04:05PM, 2006 MST
		Date: now.Format("2006-01-02"),
		Time: now.Format("3:04:00 PM"),
		TZ:   now.Format("MST"),
	}

	return ret, nil
}

func parseTimestamp(opts interactionState) (time.Time, error) {
	tz, err := parseTimezone(opts.TZ)
	if err != nil {
		return time.Time{}, err
	}

	formats := []string{
		"2006-01-02 3:04 PM",
		"2006-01-02 3:04:05 PM",
		"2006-01-02 3:04:05PM",
	}

	for _, format := range formats {
		if tm, err := time.ParseInLocation(format, fmt.Sprintf("%s %s", opts.Date, opts.Time), tz); err == nil {
			return tm, nil
		}
	}

	return time.Time{}, fmt.Errorf("could not parse timezone. Format %q expected", formats[0])
}

func parseTimezone(tz string) (*time.Location, error) {
	switch tz {
	case "HST":
		return time.FixedZone(tz, -10*60*60), nil
	case "HDT", "AKST":
		return time.FixedZone(tz, -9*60*60), nil
	case "AKDT", "PST":
		return time.FixedZone(tz, -8*60*60), nil
	case "PDT", "MST":
		return time.FixedZone(tz, -7*60*60), nil
	case "MDT", "CST":
		return time.FixedZone(tz, -6*60*60), nil
	case "CDT", "EST":
		return time.FixedZone(tz, -5*60*60), nil
	case "EDT":
		return time.FixedZone(tz, -4*60*60), nil
	default:
		return time.LoadLocation(tz)
	}
}
