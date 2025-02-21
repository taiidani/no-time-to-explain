package server

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/taiidani/no-time-to-explain/internal/data"
	"github.com/taiidani/no-time-to-explain/internal/models"
)

type indexBag struct {
	baseBag
	Messages models.Messages

	// Used when logging in
	Redirect string
	State    string
	ClientID string
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	bag := indexBag{baseBag: s.newBag(r, "home")}

	// Session challenge
	// TODO challenge all endpoints
	if _, ok := r.Context().Value(sessionKey).(*models.Session); !ok {
		template := "login.gohtml"
		bag.ClientID = os.Getenv("DISCORD_CLIENT_ID")
		bag.Redirect = os.Getenv("URL") + "/oauth/callback"
		// TODO Add proper state
		bag.State = os.Getenv("DISCORD_CLIENT_ID")
		renderHtml(w, http.StatusOK, template, bag)
		return
	}

	var messages models.Messages
	err := s.backend.Get(r.Context(), models.MessagesDBKey, &messages)
	if err != nil {
		if !errors.Is(err, data.ErrKeyNotFound) {
			errorResponse(r.Context(), w, http.StatusInternalServerError, err)
			return
		}
	}
	bag.Messages = messages

	template := "index.gohtml"
	renderHtml(w, http.StatusOK, template, bag)
}

func (s *Server) indexPostHandler(w http.ResponseWriter, r *http.Request) {
	var messages models.Messages
	err := s.backend.Get(r.Context(), models.MessagesDBKey, &messages)
	if err != nil {
		if !errors.Is(err, data.ErrKeyNotFound) {
			errorResponse(r.Context(), w, http.StatusInternalServerError, err)
			return
		}
	}

	newMessage := models.Message{
		ID:       base64.StdEncoding.EncodeToString([]byte(r.FormValue("trigger"))),
		Trigger:  r.FormValue("trigger"),
		Response: r.FormValue("response"),
	}

	// Validate inputs
	if len(newMessage.Trigger) < 4 || len(newMessage.Response) < 4 {
		errorResponse(r.Context(), w, http.StatusInternalServerError, fmt.Errorf("provided inputs need to be at least 4 characters"))
		return
	}

	// Check for existing messages
	for _, msg := range messages.Messages {
		if msg.ID == newMessage.ID {
			errorResponse(r.Context(), w, http.StatusInternalServerError, fmt.Errorf("message already found"))
			return
		}
	}

	// Save the new message
	messages.Messages = append(messages.Messages, newMessage)
	err = s.backend.Set(r.Context(), models.MessagesDBKey, messages, time.Hour*8760)
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) indexDeleteHandler(w http.ResponseWriter, r *http.Request) {
	var messages models.Messages
	err := s.backend.Get(r.Context(), models.MessagesDBKey, &messages)
	if err != nil {
		if !errors.Is(err, data.ErrKeyNotFound) {
			errorResponse(r.Context(), w, http.StatusInternalServerError, err)
			return
		}
	}

	for i, msg := range messages.Messages {
		if msg.ID == r.FormValue("id") {
			messages.Messages = append(messages.Messages[0:i], messages.Messages[i+1:]...)
		}
	}
	err = s.backend.Set(r.Context(), models.MessagesDBKey, messages, time.Hour*8760)
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}
