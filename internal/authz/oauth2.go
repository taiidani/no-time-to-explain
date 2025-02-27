package authz

import (
	"context"
	"fmt"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/taiidani/no-time-to-explain/internal/models"
	"golang.org/x/oauth2"
)

const oauthDiscordEndpoint = "https://discord.com/api/v10/oauth2"

func NewOAuth2Config() (string, string, error) {
	conf := oauth2Config()
	if conf.ClientID == "" {
		return "", "", fmt.Errorf("missing client_id in configuration")
	}

	verifier := oauth2.GenerateVerifier()
	url := conf.AuthCodeURL(verifier, oauth2.AccessTypeOffline)

	return url, verifier, nil
}

func OAuth2Callback(ctx context.Context, code string) (*oauth2.Token, error) {
	conf := oauth2Config()
	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("unable to exchange Discord code for token: %w", err)
	}

	return tok, nil
}

func OAuth2UserInformation(ctx context.Context, token *oauth2.Token) (*models.DiscordUser, error) {
	conf := oauth2Config()
	client := conf.Client(ctx, token)

	dClient, err := discordgo.New(token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("could not instantiate Discord client: %w", err)
	}

	// Look up the user
	dUser, err := dClient.User("@me", discordgo.WithClient(client))
	if err != nil {
		return nil, err
	}

	// Convert the response into an internal object
	user := &models.DiscordUser{}
	user.ID = dUser.ID
	user.Username = dUser.Username + "#" + dUser.Discriminator

	// And verify the user is a member of the bot's servers
	dGuilds, err := dClient.UserGuilds(200, "", "", false, discordgo.WithClient(client))
	if err != nil {
		return nil, err
	} else if !isAuthorizedMember(dGuilds) {
		return nil, fmt.Errorf("you are not a member of a server this bot is installed on")
	}

	return user, nil
}

func oauth2Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("DISCORD_CLIENT_ID"),
		ClientSecret: os.Getenv("DISCORD_CLIENT_SECRET"),
		Scopes:       []string{"guilds", "identify"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://discord.com/oauth2/authorize",
			TokenURL: oauthDiscordEndpoint + "/token",
		},
		RedirectURL: os.Getenv("URL") + "/oauth/callback",
	}
}

func isAuthorizedMember(guilds []*discordgo.UserGuild) bool {
	for _, guild := range guilds {
		if guild.Name == "Unknown Space" {
			return true
		}
	}

	return false
}
