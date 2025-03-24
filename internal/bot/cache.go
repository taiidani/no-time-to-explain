package bot

import (
	"github.com/bwmarrin/discordgo"
	"github.com/taiidani/no-time-to-explain/internal/data"
)

// cache is a singleton holding either a Redis or Memory backed database
var cache data.Cache

func InitCache(newCache data.Cache) {
	cache = newCache
}

// state represents the internal persistence layer between each user's invocation.
type state struct {
	TZ string `json:"tz"`
}

const dbUserPrefix = "user:"

func generateStateKey(i *discordgo.InteractionCreate) string {
	var userID string
	if i.User != nil {
		userID = i.User.ID
	}
	if i.Member != nil {
		userID = i.Member.User.ID
	}

	return dbUserPrefix + userID
}
