package authz

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/taiidani/no-time-to-explain/internal/data"
	"github.com/taiidani/no-time-to-explain/internal/models"
)

const defaultSessionExpiration = time.Duration(time.Hour * 168)

func NewSession(ctx context.Context, sess models.Session, backend data.Cache) (*http.Cookie, error) {
	sessionKey := uuid.New().String()
	err := backend.Set(ctx, "session:"+sessionKey, sess, defaultSessionExpiration)
	if err != nil {
		return nil, err
	}

	cookie := http.Cookie{
		Name:     "session",
		Value:    sessionKey,
		Secure:   os.Getenv("DEV") != "true",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   int(defaultSessionExpiration.Seconds()),
	}
	return &cookie, nil
}

func DeleteSession() *http.Cookie {
	cookie := http.Cookie{
		Name:     "session",
		Value:    "",
		Secure:   os.Getenv("DEV") != "true",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	}
	return &cookie
}

func GetSession(r *http.Request, backend data.Cache) (*models.Session, error) {
	var sess *models.Session
	cookie, err := r.Cookie("session")
	if err != nil {
		// No cookie 🍪
		return nil, nil
	}

	_, err = backend.Get(r.Context(), "session:"+cookie.Value, &sess)
	if err != nil {
		return nil, fmt.Errorf("failed to load session from backend: %w", err)
	}

	return sess, nil
}

func UpdateSession(r *http.Request, sess *models.Session, backend data.Cache) error {
	cookie, err := r.Cookie("session")
	if err != nil {
		// No cookie 🍪
		return fmt.Errorf("no session found to update")
	}

	err = backend.Set(r.Context(), "session:"+cookie.Value, &sess, defaultSessionExpiration)
	if err != nil {
		return fmt.Errorf("failed to update session in backend: %w", err)
	}

	return nil
}
