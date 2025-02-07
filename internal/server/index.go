package server

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/taiidani/no-time-to-explain/internal/data"
	"github.com/taiidani/no-time-to-explain/internal/models"
)

type indexBag struct {
	baseBag
	Messages models.Messages
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	bag := indexBag{baseBag: s.newBag(r, "home")}

	var messages models.Messages
	err := s.backend.Get(r.Context(), models.MessagesDBKey, &messages)
	if err != nil {
		if !errors.Is(err, data.ErrKeyNotFound) {
			errorResponse(w, http.StatusInternalServerError, err)
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
			errorResponse(w, http.StatusInternalServerError, err)
			return
		}
	}

	newMessage := models.Message{
		ID:       base64.StdEncoding.EncodeToString([]byte(r.FormValue("trigger"))),
		Trigger:  r.FormValue("trigger"),
		Response: r.FormValue("response"),
	}

	// Check for existing messages
	for _, msg := range messages.Messages {
		if msg.ID == newMessage.ID {
			errorResponse(w, http.StatusInternalServerError, fmt.Errorf("message already found"))
			return
		}
	}

	// Save the new message
	messages.Messages = append(messages.Messages, newMessage)
	err = s.backend.Set(r.Context(), models.MessagesDBKey, messages, time.Hour*8760)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) indexDeleteHandler(w http.ResponseWriter, r *http.Request) {
	var messages models.Messages
	err := s.backend.Get(r.Context(), models.MessagesDBKey, &messages)
	if err != nil {
		if !errors.Is(err, data.ErrKeyNotFound) {
			errorResponse(w, http.StatusInternalServerError, err)
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
		errorResponse(w, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}
