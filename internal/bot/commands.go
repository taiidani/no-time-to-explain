package bot

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/getsentry/sentry-go"
	"github.com/taiidani/no-time-to-explain/internal/data"
)

const (
	defaultFooter     = "Written with 💙 for Unknown Space by @taiidani"
	defaultColor      = 0x05FF05
	defaultErrorColor = 0xFF5050
)

type applicationCommand struct {
	Command           *discordgo.ApplicationCommand
	Autocomplete      func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate, o *discordgo.ApplicationCommandInteractionDataOption) []*discordgo.ApplicationCommandOptionChoice
	MessageComponents map[string]func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate)
	Handler           func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate)
}

type Commands struct {
	commands []applicationCommand
	registry []*discordgo.ApplicationCommand
	s        *discordgo.Session
	db       data.Cache
}

func NewCommands(session *discordgo.Session, db data.Cache) *Commands {
	ret := Commands{
		commands: []applicationCommand{
			{
				Command: &discordgo.ApplicationCommand{
					Name:        "time",
					Description: "Render a Discord-style timestamp for sharing with others",
					Type:        discordgo.ChatApplicationCommand,
					Options:     []*discordgo.ApplicationCommandOption{},
				},
				Handler: timeHandler,
				MessageComponents: map[string]func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate){
					changeTimeCustomID:      changeTimeHandler,
					nowTimeCustomID:         nowTimeHandler,
					changeTimeModalCustomID: changeTimeSubmitHandler,
				},
			},
			{
				Command: &discordgo.ApplicationCommand{
					// Parse Charlemagne events and generate exportable calendar items
					Name:    "Event Calendar",
					Type:    discordgo.MessageApplicationCommand,
					Options: []*discordgo.ApplicationCommandOption{},
				},
				Handler: eventCalendarHandler,
			},
		},
		registry: []*discordgo.ApplicationCommand{},
		s:        session,
		db:       db,
	}

	return &ret
}

func (c *Commands) AddHandlers() {
	c.s.Identify.Intents = discordgo.IntentsGuildMessages

	c.s.AddHandler(c.handleReady)
	c.s.AddHandler(c.handleCommand)
	c.s.AddHandler(c.handleMessage)
}

func (c *Commands) handleReady(s *discordgo.Session, event *discordgo.Ready) {
	for _, cmd := range c.commands {
		fmt.Printf("Registering global application command %q for bot user %q\n", cmd.Command.Name, s.State.User.ID)
		ccmd, err := s.ApplicationCommandCreate(s.State.User.ID, "", cmd.Command)
		if err != nil {
			log.Printf("Unable to set application command %q: %s", cmd.Command.Name, err)
		}

		c.registry = append(c.registry, ccmd)
	}
}

func (c *Commands) handleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Set up the Sentry transaction
	hub := sentry.CurrentHub().Clone()
	addSentry(i, hub)
	ctx := sentry.SetHubOnContext(context.Background(), hub)

	transaction := sentry.StartTransaction(ctx, "command")
	defer transaction.Finish()
	ctx = transaction.Context()

	for _, cmd := range c.commands {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			if cmd.Command.Name == i.ApplicationCommandData().Name {
				transaction.Name = cmd.Command.Name

				cmd.Handler(ctx, s, i)
			}
		case discordgo.InteractionApplicationCommandAutocomplete:
			if cmd.Autocomplete != nil && cmd.Command.Name == i.ApplicationCommandData().Name {
				transaction.Name = cmd.Command.Name + "-autocomplete"

				for _, opt := range i.ApplicationCommandData().Options {
					if opt.Focused {
						choices := cmd.Autocomplete(ctx, s, i, opt)
						_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
							Type: discordgo.InteractionApplicationCommandAutocompleteResult,
							Data: &discordgo.InteractionResponseData{Choices: choices},
						})
					}
				}
			}
		case discordgo.InteractionMessageComponent:
			if cmd.MessageComponents != nil {
				transaction.Name = cmd.Command.Name + "-component-interact"
				transaction.SetData("custom-id", i.MessageComponentData().CustomID)

				transaction.Name = i.MessageComponentData().CustomID
				log.Println(i.MessageComponentData().CustomID)
				for customID, fn := range cmd.MessageComponents {
					if strings.HasPrefix(i.MessageComponentData().CustomID, customID) {
						fn(ctx, s, i)
					}
				}
			}
		case discordgo.InteractionModalSubmit:
			if cmd.MessageComponents != nil {
				transaction.Name = cmd.Command.Name + "-modal-submit"
				transaction.SetData("custom-id", i.ModalSubmitData().CustomID)

				log.Println(i.ModalSubmitData().CustomID)
				for customID, fn := range cmd.MessageComponents {
					if strings.HasPrefix(i.ModalSubmitData().CustomID, customID) {
						fn(ctx, s, i)
					}
				}
			}
		default:
			log.Println("Unknown interaction type encountered: ", i.Type)
		}
	}
}

func (c *Commands) Teardown() {
	for _, cmd := range c.registry {
		err := c.s.ApplicationCommandDelete(cmd.ApplicationID, "", cmd.ID)
		if err != nil {
			log.Printf("Cannot delete slash command %q: %v", cmd.Name, err)
		}
	}
}

func commandError(s *discordgo.Session, i *discordgo.Interaction, message error) {
	_ = s.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf(":warning: %s", message),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func addSentry(evt interface{}, hub *sentry.Hub) {
	hub.ConfigureScope(func(scope *sentry.Scope) {
		// Add user information to Sentry
		user := sentry.User{}
		switch i := evt.(type) {
		case *discordgo.InteractionCreate:
			if i.Member != nil && i.Member.User != nil {
				user.ID = i.Member.User.ID
				user.Username = i.Member.User.Username + "#" + i.Member.User.Discriminator
			} else if i.User != nil {
				user.ID = i.User.ID
				user.Username = i.User.Username + "#" + i.User.Discriminator
			}
		case *discordgo.MessageCreate:
			if i.Author != nil {
				user.ID = i.Author.ID
				user.Username = i.Author.Username + "#" + i.Author.Discriminator
			}
		default:
			slog.Warn("Uninstrumented event received, could not populate Sentry", "event", evt)
		}

		if !user.IsEmpty() {
			scope.SetUser(user)
		}
	})
}
