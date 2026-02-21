package server

import (
	"log/slog"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/taiidani/no-time-to-explain/internal/models"
)

const (
	unknownSpaceServerID    = "570720951373922304"
	taiidaniTestingServerID = "372591705754566656"
)

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	type indexBag struct {
		baseBag
		Channels      []*discordgo.Channel
		MessageGroups map[string][]models.Message
		Bluesky       struct {
			Feeds []models.Feed
		}
	}

	bag := indexBag{baseBag: s.newBag(r)}

	// Load the current messages that the bot is listening to
	messages, err := models.LoadMessages(r.Context())
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	// Group messages by trigger and sender
	bag.MessageGroups = map[string][]models.Message{}
	for _, msg := range messages {
		bag.MessageGroups[msg.Trigger] = append(bag.MessageGroups[msg.Trigger], msg)
		// if _, exists := bag.MessageGroups[msg.Trigger]; exists {

		// } else {
		// 	bag.MessageGroups[msg.Trigger] = []models.Message{msg}
		// }
	}

	// Sort groups by trigger for consistent display
	// sort.Slice(bag.MessageGroups, func(i, j int) bool {
	// 	if bag.MessageGroups[i].Trigger != bag.MessageGroups[j].Trigger {
	// 		return bag.MessageGroups[i].Trigger < bag.MessageGroups[j].Trigger
	// 	}
	// 	return bag.MessageGroups[i].Sender < bag.MessageGroups[j].Sender
	// })

	// Load all currently subscribed feeds
	feeds, err := models.LoadFeeds(r.Context())
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}
	bag.Bluesky.Feeds = feeds

	// Load all channels in Unknown Space, fall back upon internal testing
	for _, guildID := range []string{unknownSpaceServerID, taiidaniTestingServerID} {
		channels, err := s.discord.GuildChannels(guildID, discordgo.WithContext(r.Context()))
		if err != nil {
			slog.Warn("Skipping guild", "id", guildID, "err", err.Error())
			continue
		}
		bag.Channels = append(bag.Channels, channels...)
	}

	sort.Slice(bag.Channels, func(i, j int) bool {
		left := strings.ToLower(bag.Channels[i].Name)
		right := strings.ToLower(bag.Channels[j].Name)
		return left < right
	})

	template := "index.gohtml"
	renderHtml(w, http.StatusOK, template, bag)
}

func (s *Server) channelsHandler(w http.ResponseWriter, r *http.Request) {
	type indexBag struct {
		baseBag
		Channels []*discordgo.Channel
	}

	bag := indexBag{baseBag: s.newBag(r)}

	// Load all channels in Unknown Space, fall back upon internal testing
	for _, guildID := range []string{unknownSpaceServerID, taiidaniTestingServerID} {
		channels, err := s.discord.GuildChannels(guildID, discordgo.WithContext(r.Context()))
		if err != nil {
			slog.Warn("Skipping guild", "id", guildID, "err", err.Error())
			continue
		}
		bag.Channels = append(bag.Channels, channels...)
	}

	sort.Slice(bag.Channels, func(i, j int) bool {
		left := strings.ToLower(bag.Channels[i].Name)
		right := strings.ToLower(bag.Channels[j].Name)
		return left < right
	})

	template := "channels.gohtml"
	renderHtml(w, http.StatusOK, template, bag)
}

func (s *Server) usersHandler(w http.ResponseWriter, r *http.Request) {
	type indexBag struct {
		baseBag
		Users []discordgo.User
	}

	bag := indexBag{baseBag: s.newBag(r)}

	// Load all recent senders from the cache
	userKeys, err := s.backend.Keys(r.Context(), "recent-senders:*")
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	for _, key := range userKeys {
		key = strings.TrimPrefix(key, "no-time-to-explain:")

		var user discordgo.User
		err = s.backend.Get(r.Context(), key, &user)
		if err != nil {
			errorResponse(r.Context(), w, http.StatusInternalServerError, err)
			return
		}

		bag.Users = append(bag.Users, user)
	}

	sort.Slice(bag.Users, func(i, j int) bool {
		left := strings.ToLower(bag.Users[i].Username)
		right := strings.ToLower(bag.Users[j].Username)
		return left < right
	})

	template := "users.gohtml"
	renderHtml(w, http.StatusOK, template, bag)
}

func (s *Server) feedAddHandler(w http.ResponseWriter, r *http.Request) {
	newFeed := models.Feed{
		Source:      "bluesky",
		Author:      r.FormValue("author"),
		LastMessage: time.Now(),
	}

	// Validate inputs
	if err := newFeed.Validate(); err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	// Save the new Feed
	err := models.AddFeed(r.Context(), newFeed)
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) feedDeleteHandler(w http.ResponseWriter, r *http.Request) {
	err := models.DeleteFeed(r.Context(), r.FormValue("id"))
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}
