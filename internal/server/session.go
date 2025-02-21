package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
	"github.com/taiidani/no-time-to-explain/internal/models"
)

type contextKey string

var sessionKey contextKey = "session"

func (s *Server) authCallback(w http.ResponseWriter, r *http.Request) {
	sess, ok := r.Context().Value(sessionKey).(models.Session)
	if !ok {
		sess = models.Session{}
	}

	query := r.URL.Query()

	// TODO Add proper state handling
	state := query.Get("state")
	if state != os.Getenv("DISCORD_CLIENT_ID") {
		errorResponse(w, http.StatusInternalServerError, fmt.Errorf("improper state"))
		return
	}

	// Convert code to Discord information
	code := query.Get("code")
	user, auth, err := oauthToken(r.Context(), code)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, fmt.Errorf("unable to validate OAuth code from Discord: %w", err))
		return
	}

	sess.Auth = auth
	sess.DiscordUser = user

	// Set the session
	const defaultSessionExpiration = time.Duration(time.Hour * 168)
	sessionKey := s.buildSessionKey()
	err = s.backend.Set(r.Context(), "session:"+sessionKey, sess, defaultSessionExpiration)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err)
		return
	}

	cookie := http.Cookie{
		Name:     "session",
		Value:    sessionKey,
		Secure:   !DevMode,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   int(defaultSessionExpiration.Seconds()),
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (s *Server) sessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s %s\n", r.Method, r.URL.Path)

		// Is the user ID in the session?
		cookie, err := r.Cookie("session")
		if err == nil {
			newRequest, err := s.loadSession(r, cookie.Value)
			if err != nil {
				slog.Warn("Unable to retrieve session", "key", cookie.Value, "error", err)
			} else {
				r = newRequest
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) buildSessionKey() string {
	key := uuid.New()
	return key.String()
}

func (s *Server) loadSession(r *http.Request, sessionID string) (*http.Request, error) {
	var sess *models.Session
	err := s.backend.Get(r.Context(), "session:"+sessionID, &sess)
	if err != nil {
		return nil, fmt.Errorf("failed to load session from backend: %w", err)
	}

	if sess == nil || sess.Auth == nil || sess.DiscordUser == nil {
		return nil, fmt.Errorf("empty session found")
	}

	// Is the refreshToken expiring? Get a new set!
	if time.Now().Add(time.Minute * 5).After(sess.Auth.ExpiresAt) {
		sess.DiscordUser, sess.Auth, err = oauthRefresh(r.Context(), *sess.Auth)
		if err != nil {
			return nil, fmt.Errorf("failed to refresh oauth token: %w", err)
		}

		err = s.backend.Set(r.Context(), "session:"+sessionID, &sess, time.Hour*24)
		if err != nil {
			return nil, fmt.Errorf("failed to update backend with new session: %w", err)
		}
	} else if time.Now().After(sess.Auth.ExpiresAt) {
		// Already expired :( back to the login page
		return nil, fmt.Errorf("session expired")
	}

	newContext := context.WithValue(r.Context(), sessionKey, sess)
	hub := sentry.GetHubFromContext(newContext)
	if hub == nil {
		hub = sentry.CurrentHub().Clone()
		newContext = sentry.SetHubOnContext(newContext, hub)
	}
	hub.Scope().SetUser(sentry.User{
		ID:       sess.DiscordUser.ID,
		Username: sess.DiscordUser.Username,
	})

	return r.WithContext(newContext), nil
}

const oauthDiscordEndpoint = "https://discord.com/api/v10/oauth2"

// oauthToken will exchange a given OAuth code with access & refresh tokens
// url: https://discord.com/developers/docs/topics/oauth2#authorization-code-grant-access-token-response
func oauthToken(ctx context.Context, code string) (user *models.DiscordUser, auth *models.DiscordAuth, err error) {
	params := url.Values{}
	params.Set("grant_type", "authorization_code")
	params.Set("code", code)
	params.Set("redirect_uri", os.Getenv("URL")+"/oauth/callback")
	return oauthExchange(ctx, params)
}

// oauthToken will exchange a given refresh token with new access & refresh tokens
// url: https://discord.com/developers/docs/topics/oauth2#authorization-code-grant-access-token-response
func oauthRefresh(ctx context.Context, stale models.DiscordAuth) (user *models.DiscordUser, auth *models.DiscordAuth, err error) {
	params := url.Values{}
	params.Set("grant_type", "refresh_token")
	params.Set("refresh_token", stale.RefreshToken)
	return oauthExchange(ctx, params)
}

type oauthDiscordAccessTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

// oauthExchange will perform the actual OAuth negotiation with Discord
func oauthExchange(ctx context.Context, params url.Values) (user *models.DiscordUser, auth *models.DiscordAuth, err error) {
	paramsReader := strings.NewReader(params.Encode())
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, oauthDiscordEndpoint+"/token", paramsReader)
	if err != nil {
		return nil, nil, err
	}
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.SetBasicAuth(os.Getenv("DISCORD_CLIENT_ID"), os.Getenv("DISCORD_CLIENT_SECRET"))

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, nil, err
	} else {
		switch resp.StatusCode {
		case 200:
			// Nada
		case 401:
			body, _ := io.ReadAll(resp.Body)
			slog.Warn("Authorization failure from Discord", "code", resp.StatusCode, "body", body)
			return nil, nil, fmt.Errorf("you are not authorized")
		default:
			body, _ := io.ReadAll(resp.Body)
			slog.Warn("Unexpected response code from Discord encountered", "code", resp.StatusCode, "body", body)
			return nil, nil, fmt.Errorf("unexpected response from Discord")
		}
	}

	// Parse the response containing the access token, expiration, and refresh token
	data := oauthDiscordAccessTokenResponse{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, nil, fmt.Errorf("could not parse Discord OAuth token response: %w", err)
	} else if data.AccessToken == "" {
		return nil, nil, fmt.Errorf("empty Discord OAuth token response encountered")
	}
	defer resp.Body.Close()

	// Convert the response into an internal object
	auth = &models.DiscordAuth{}
	auth.AccessToken = data.AccessToken
	auth.RefreshToken = data.RefreshToken
	auth.ExpiresAt = time.Now().Add(time.Second * time.Duration(data.ExpiresIn))

	// Get the Discord user information
	user, err = userInformation(ctx, auth.AccessToken)
	if err != nil {
		return nil, nil, fmt.Errorf("could not extract information about the current Discord user: %w", err)
	}

	return user, auth, nil
}

func userInformation(ctx context.Context, accessToken string) (*models.DiscordUser, error) {
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, oauthDiscordEndpoint+"/@me", nil)
	if err != nil {
		return nil, err
	}
	r.Header.Add("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(r)
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
