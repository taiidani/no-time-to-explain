package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

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

	// TODO convert code to Discord information
	code := query.Get("code")
	sess.DiscordID = code

	// Set the session
	const defaultSessionExpiration = time.Duration(time.Hour * 168)
	sessionKey := s.buildSessionKey()
	err := s.backend.Set(r.Context(), "session:"+sessionKey, sess, defaultSessionExpiration)
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
			var sess *models.Session
			err = s.backend.Get(r.Context(), "session:"+cookie.Value, &sess)
			if err != nil {
				slog.Warn("Unable to retrieve session", "key", cookie.Value, "error", err)
			} else if sess != nil && sess.DiscordID != "" {
				newContext := context.WithValue(r.Context(), sessionKey, sess)
				r = r.WithContext(newContext)
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) buildSessionKey() string {
	key := uuid.New()
	return key.String()
}
