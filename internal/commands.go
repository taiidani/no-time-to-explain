package internal

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type applicationCommand struct {
	Command           *discordgo.ApplicationCommand
	Autocomplete      func(s *discordgo.Session, i *discordgo.InteractionCreate, o *discordgo.ApplicationCommandInteractionDataOption) []*discordgo.ApplicationCommandOptionChoice
	MessageComponents map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)
	Handler           func(s *discordgo.Session, i *discordgo.InteractionCreate)
}

type Commands struct {
	commands []applicationCommand
	registry []*discordgo.ApplicationCommand
	s        *discordgo.Session
}

func NewCommands(session *discordgo.Session) *Commands {
	ret := Commands{
		commands: []applicationCommand{
			{
				Command: &discordgo.ApplicationCommand{
					Name:        "time",
					Description: "Render a Discord-style timestamp for sharing with others",
					Type:        discordgo.ChatApplicationCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "date",
							Description: "The date for the timestamp, formatted as YYYYMMDD",
							Required:    false,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "time",
							Description: "The time for the timestamp, formatted as HH:MM:SS",
							Required:    false,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "tz",
							Description: "The timezone for the timestamp",
							Required:    false,
						},
					},
				},
				Handler: timeHandler,
				MessageComponents: map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
					tzHandlerCustomID:       tzHandler,
					changeTimeCustomID:      changeTimeHandler,
					changeTimeModalCustomID: changeTimeSubmitHandler,
				},
			},
		},
		registry: []*discordgo.ApplicationCommand{},
		s:        session,
	}

	return &ret
}

func (c *Commands) AddHandlers() {
	c.s.AddHandler(c.handleReady)
	c.s.AddHandler(c.handleCommand)
}

func (c *Commands) handleReady(s *discordgo.Session, event *discordgo.Ready) {
	for _, g := range event.Guilds {
		for _, cmd := range c.commands {
			fmt.Printf("Registering application command %q for bot user %q in guild %q\n", cmd.Command.Name, s.State.User.ID, g.ID)
			ccmd, err := s.ApplicationCommandCreate(s.State.User.ID, g.ID, cmd.Command)
			if err != nil {
				log.Printf("Unable to set application command %q: %s", cmd.Command.Name, err)
				continue
			}

			c.registry = append(c.registry, ccmd)
		}
	}
}

func (c *Commands) handleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	for _, cmd := range c.commands {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			if cmd.Command.Name == i.ApplicationCommandData().Name {
				cmd.Handler(s, i)
			}
		case discordgo.InteractionApplicationCommandAutocomplete:
			if cmd.Autocomplete != nil && cmd.Command.Name == i.ApplicationCommandData().Name {
				for _, opt := range i.ApplicationCommandData().Options {
					if opt.Focused {
						choices := cmd.Autocomplete(s, i, opt)
						_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
							Type: discordgo.InteractionApplicationCommandAutocompleteResult,
							Data: &discordgo.InteractionResponseData{Choices: choices},
						})
					}
				}
			}
		case discordgo.InteractionMessageComponent:
			if cmd.MessageComponents != nil {
				log.Println(i.MessageComponentData().CustomID)
				for customID, fn := range cmd.MessageComponents {
					if strings.HasPrefix(i.MessageComponentData().CustomID, customID) {
						fn(s, i)
					}
				}
			}
		case discordgo.InteractionModalSubmit:
			if cmd.MessageComponents != nil {
				log.Println(i.ModalSubmitData().CustomID)
				for customID, fn := range cmd.MessageComponents {
					if strings.HasPrefix(i.ModalSubmitData().CustomID, customID) {
						fn(s, i)
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
		err := c.s.ApplicationCommandDelete(cmd.ApplicationID, cmd.GuildID, cmd.ID)
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
