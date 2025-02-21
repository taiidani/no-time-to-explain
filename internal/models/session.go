package models

import (
	"golang.org/x/oauth2"
)

type Session struct {
	State       string // Used in the OAuth2 flow to validate the request
	Auth        *oauth2.Token
	DiscordUser *DiscordUser
}

type DiscordUser struct {
	ID       string
	Username string
}
