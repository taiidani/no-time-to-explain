package authz

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

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

	r, err := http.NewRequestWithContext(ctx, http.MethodGet, oauthDiscordEndpoint+"/@me", nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(r)
	if err != nil {
		return nil, err
	} else {
		switch resp.StatusCode {
		case 200:
			// Nada
		case 401:
			body, _ := io.ReadAll(resp.Body)
			slog.Warn("Authorization failure from Discord", "code", resp.StatusCode, "body", body)
			return nil, fmt.Errorf("you are not authorized")
		default:
			body, _ := io.ReadAll(resp.Body)
			slog.Warn("Unexpected response code from Discord encountered", "code", resp.StatusCode, "body", body)
			return nil, fmt.Errorf("unexpected response from Discord")
		}
	}

	type meResponse struct {
		User struct {
			ID            string `json:"id"`
			Username      string `json:"username"`
			Avatar        string `json:"avatar"`
			Discriminator string `json:"discriminator"`
			GlobalName    string `json:"global_name"`
			PublicFlags   int    `json:"public_flags"`
		} `json:"user"`
	}

	// Parse the response containing the access token, expiration, and refresh token
	data := meResponse{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("could not parse Discord @me response: %w", err)
	}
	defer resp.Body.Close()

	// Convert the response into an internal object
	user := &models.DiscordUser{}
	user.ID = data.User.ID
	user.Username = data.User.Username + "#" + data.User.Discriminator

	return user, nil
}

func oauth2Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("DISCORD_CLIENT_ID"),
		ClientSecret: os.Getenv("DISCORD_CLIENT_SECRET"),
		Scopes:       []string{"identify"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://discord.com/oauth2/authorize",
			TokenURL: oauthDiscordEndpoint + "/token",
		},
		RedirectURL: os.Getenv("URL") + "/oauth/callback",
	}
}
