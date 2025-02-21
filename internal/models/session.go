package models

import (
	"time"
)

type Session struct {
	Auth        *DiscordAuth
	DiscordUser *DiscordUser
}

type DiscordUser struct {
	ID       string
	Username string
}

type DiscordAuth struct {
	AccessToken  string
	ExpiresAt    time.Time
	RefreshToken string
}
