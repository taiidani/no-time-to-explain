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
	ret := charlemagneEvent{ID: msg.ID}

	if len(msg.Embeds) == 0 {
		return ret, fmt.Errorf("message did not contain any embeds")
	}
	if len(msg.Embeds[0].Fields) == 0 {
		return ret, fmt.Errorf("message embed did not contain any fields")
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

	content := strings.Builder{}
	for _, evt := range events {
		content.WriteString(fmt.Sprintf("> %s. Join ID: %s\n", evt.Activity, evt.JoinID))
		content.WriteString(fmt.Sprintf("> <t:%d:F>\n", evt.StartTime.Unix()))
		content.WriteString(fmt.Sprintf("> Guardians: %s\n\n", strings.Join(evt.Guardians, ", ")))

		ret.Files = append(ret.Files, &discordgo.File{
			Name:        fmt.Sprintf("D2-%s-%s.ics", evt.Activity, evt.ID),
			ContentType: "text/calendar",
			Reader:      buildCalendarEvent(evt),
		})
	}

	ret.Content = content.String()
	return ret, nil
}

func buildCalendarEvent(evt charlemagneEvent) *strings.Reader {
	// For reference: 2006-01-02T15:04:05Z07:00
	const dtFormat = "20060102T150405Z"
	evt.StartTime = evt.StartTime.In(time.UTC)
	dtEnd := evt.StartTime.Add(time.Hour * 2)

	ical := strings.Builder{}
	ical.WriteString("BEGIN:VCALENDAR\n")
	ical.WriteString("VERSION:2.0\n")
	ical.WriteString("PRODID:-//hacksw/handcal//NONSGML v1.0//EN\n")
	ical.WriteString("BEGIN:VEVENT\n")
	ical.WriteString("UID:" + evt.ID + "\n")
	ical.WriteString("DTSTART:" + evt.StartTime.Format(dtFormat) + "\n")
	ical.WriteString("DTEND:" + dtEnd.Format(dtFormat) + "\n")
	ical.WriteString("SUMMARY:" + evt.Activity + "\n")
	ical.WriteString("DESCRIPTION:" + evt.Description + "\n")
	ical.WriteString("END:VEVENT\n")
	ical.WriteString("END:VCALENDAR\n")

	return strings.NewReader(ical.String())
}
