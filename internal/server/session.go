package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/taiidani/no-time-to-explain/internal/authz"
	"github.com/taiidani/no-time-to-explain/internal/models"
)

type contextKey string

var sessionKey contextKey = "session"

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	bag := s.newBag(r)
	template := "login.gohtml"
	renderHtml(w, http.StatusOK, template, bag)
}

func (s *Server) auth(w http.ResponseWriter, r *http.Request) {
	// Generate the OAuth2 URL and verification string
	url, verifier, err := authz.NewOAuth2Config()
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, fmt.Errorf("could not create new oauth2 config: %w", err))
		return
	}

	sess := models.Session{State: verifier}
	cookie, err := s.sessionManager.Create(r.Context(), &sess)
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, fmt.Errorf("could not create new session: %w", err))
		return
	}

	http.SetCookie(w, cookie)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	cookie := s.sessionManager.Delete()
	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (s *Server) authCallback(w http.ResponseWriter, r *http.Request) {
	sess := models.Session{}
	err := s.sessionManager.Get(r, &sess)
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, fmt.Errorf("unable to retrieve session: %w", err))
		return
	}

	// First, validate the request
	query := r.URL.Query()
	if query.Get("state") != sess.State {
		slog.Warn("Session state and OAuth2 callback state did not match", "session", sess.State, "request", query.Get("state"))
		errorResponse(r.Context(), w, http.StatusBadRequest, fmt.Errorf("unable to verify oauth2 request"))
		return
	}
	sess.State = ""

	// And see if it has an error
	if query.Get("error_description") != "" {
		errorResponse(r.Context(), w, http.StatusBadRequest, errors.New(query.Get("error_description")))
		return
	}

	// Next, exchange the OAuth code for a token
	sess.Auth, err = authz.OAuth2Callback(r.Context(), query.Get("code"))
	if err != nil {
		errorResponse(r.Context(), w, http.StatusBadRequest, fmt.Errorf("unable to validate OAuth code from Discord: %w", err))
		return
	}

	// Convert token to Discord information
	sess.DiscordUser, err = authz.OAuth2UserInformation(r.Context(), sess.Auth)
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, fmt.Errorf("unable to look up user information from Discord: %w", err))
		return
	}

	// Set the session
	err = s.sessionManager.Update(r, sess)
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (s *Server) sessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info(r.Method, "path", r.URL.Path)

		// Do we have a session already?
		sess := models.Session{}
		err := s.sessionManager.Get(r, &sess)
		if err != nil {
			if !errors.Is(err, http.ErrNoCookie) {
				slog.Warn("Failed to retrieve session", "error", err)
			}

			// No session! Login page
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		// Embed the user information for Sentry
		if hub := sentry.GetHubFromContext(r.Context()); hub != nil {
			hub.Scope().SetUser(sentry.User{
				ID:       sess.DiscordUser.ID,
				Username: sess.DiscordUser.Username,
			})
		}

		ctx := context.WithValue(r.Context(), sessionKey, sess)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
