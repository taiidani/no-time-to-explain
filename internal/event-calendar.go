package internal

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

type charlemagneEvent struct {
	ID          string
	URL         string
	GuildID     string
	ChannelID   string
	Activity    string
	StartTime   time.Time
	Description string
	JoinID      string
	Guardians   []string
}

func eventCalendarHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	var messages []*discordgo.Message
	if i.Message != nil {
		messages = append(messages, i.Message)
	} else {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			data := i.ApplicationCommandData()

			for _, msg := range data.Resolved.Messages {
				messages = append(messages, msg)
			}
		}
	}

	if len(messages) == 0 {
		errorMessage(s, i.Interaction, fmt.Errorf("message not found"))
		return
	}

	events := []charlemagneEvent{}

	for _, msg := range messages {
		if msg == nil {
			continue
		}
		if !msg.Author.Bot {
			continue
		}

		// GuildID is not always populated 🤷
		if msg.GuildID == "" {
			msg.GuildID = i.GuildID
		}

		evt, err := parseEvent(ctx, i.Member, *msg)
		if err != nil {
			slog.Warn("Unable to parse message", "author", msg.Author.Username)
			continue
		}

		events = append(events, evt)
	}

	if len(events) == 0 {
		errorMessage(s, i.Interaction, fmt.Errorf("no Charlemagne events found in message"))
		return
	}

	reply, err := eventCalendarResponse(events)
	if err != nil {
		errorMessage(s, i.Interaction, err)
		return
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: reply,
	})
	if err != nil {
		log.Println("Could not respond to user message:", err)
		commandError(s, i.Interaction, err)
		return
	}
}

func parseEvent(_ context.Context, _ *discordgo.Member, msg discordgo.Message) (charlemagneEvent, error) {
	ret := charlemagneEvent{
		ID:        msg.ID,
		ChannelID: msg.ChannelID,
		GuildID:   msg.GuildID,
	}

	if len(msg.Embeds) == 0 {
		return ret, fmt.Errorf("message did not contain any embeds")
	}
	if len(msg.Embeds[0].Fields) == 0 {
		return ret, fmt.Errorf("message embed did not contain any fields")
	}

	if ret.GuildID != "" && ret.ChannelID != "" && ret.ID != "" {
		// Discord links are https://discord.com/channels/372591705754566656/1241174079030169641/1241294928340844601
		ret.URL = fmt.Sprintf("https://discord.com/channels/%s/%s/%s", ret.GuildID, ret.ChannelID, ret.ID)
	}

	for _, f := range msg.Embeds[0].Fields {
		switch {
		case f.Name == "Activity":
			ret.Activity = f.Value
		case f.Name == "Join Id":
			ret.JoinID = f.Value
		case f.Name == "Start Time":
			ret.StartTime = parseEventCalendarStartTime(f.Value)
		case f.Name == "Description":
			ret.Description = f.Value
		case strings.HasPrefix(f.Name, "Guardians Joined"):
			ret.Guardians = strings.Split(f.Value, " | ")
		}
	}

	if ret.StartTime.IsZero() {
		return ret, fmt.Errorf("could not parse Start Time")
	}

	return ret, nil
}

func parseEventCalendarStartTime(tm string) time.Time {
	re := regexp.MustCompile(`\<t\:(\d+)\:`)
	matches := re.FindStringSubmatch(tm)
	if matches == nil || len(matches) == 1 {
		return time.Time{}
	}

	unixTime, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return time.Time{}
	}

	return time.Unix(unixTime, 0)
}

func eventCalendarResponse(events []charlemagneEvent) (*discordgo.InteractionResponseData, error) {
	ret := &discordgo.InteractionResponseData{
		Flags: discordgo.MessageFlagsEphemeral,
	}

	if len(events) == 0 {
		return ret, fmt.Errorf("no events were found")
	}

	for _, evt := range events {
		description := "Event information parsed:"
		if evt.URL != "" {
			description = "Event information parsed for " + evt.URL + ":"
		}

		ret.Embeds = append(ret.Embeds, &discordgo.MessageEmbed{
			Description: description,
			Color:       defaultColor,
			Footer:      &discordgo.MessageEmbedFooter{Text: "Event ID: " + evt.JoinID},
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Activity",
					Value:  evt.Activity,
					Inline: true,
				},
				{
					Name:   "Start Time",
					Value:  fmt.Sprintf("<t:%d:F>\n", evt.StartTime.Unix()),
					Inline: true,
				},
				{
					Name:   "Description",
					Value:  evt.Description,
					Inline: false,
				},
				{
					Name:   "Guardians",
					Value:  strings.Join(evt.Guardians, ", "),
					Inline: false,
				},
			},
		})

		ret.Files = append(ret.Files, &discordgo.File{
			Name:        fmt.Sprintf("D2-%s-%s-%s.ics", evt.Activity, evt.GuildID, evt.JoinID),
			ContentType: "text/calendar",
			Reader:      buildCalendarEvent(evt),
		})
	}

	return ret, nil
}

func buildCalendarEvent(evt charlemagneEvent) *strings.Reader {
	// Go date format for reference: 2006-01-02T15:04:05Z07:00
	const dtFormat = "20060102T150405Z"
	evt.StartTime = evt.StartTime.In(time.UTC)
	dtEnd := evt.StartTime.Add(time.Hour * 2)
	dtStamp := time.Now().In(time.UTC)

	ical := strings.Builder{}
	ical.WriteString("BEGIN:VCALENDAR\n")
	ical.WriteString("VERSION:2.0\n")
	ical.WriteString("PRODID:-//hacksw/handcal//NONSGML v1.0//EN\n")
	ical.WriteString("BEGIN:VEVENT\n")
	ical.WriteString(fmt.Sprintf("UID:%s-%s\n", evt.GuildID, evt.JoinID))
	if evt.URL != "" {
		ical.WriteString("URL:" + evt.URL + "\n")
	}
	ical.WriteString("DTSTAMP:" + dtStamp.Format(dtFormat) + "\n")
	ical.WriteString("DTSTART:" + evt.StartTime.Format(dtFormat) + "\n")
	ical.WriteString("DTEND:" + dtEnd.Format(dtFormat) + "\n")
	ical.WriteString("SUMMARY:" + evt.Activity + "\n")
	ical.WriteString("DESCRIPTION:" + evt.Description + "\n")
	ical.WriteString("END:VEVENT\n")
	ical.WriteString("END:VCALENDAR\n")

	return strings.NewReader(ical.String())
}
