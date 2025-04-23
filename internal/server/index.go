package server

import (
	"net/http"
	"time"

	"github.com/taiidani/no-time-to-explain/internal/models"
)

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	type indexBag struct {
		baseBag
		Messages []models.Message
		Bluesky  struct {
			Feeds []models.Feed
		}
	}

	bag := indexBag{baseBag: s.newBag(r)}

	messages, err := models.LoadMessages(r.Context())
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}
	bag.Messages = messages

	feeds, err := models.LoadFeeds(r.Context())
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}
	bag.Bluesky.Feeds = feeds

	template := "index.gohtml"
	renderHtml(w, http.StatusOK, template, bag)
}

func (s *Server) feedAddHandler(w http.ResponseWriter, r *http.Request) {
	newFeed := models.Feed{
		Source:      "bluesky",
		Author:      r.FormValue("author"),
		LastMessage: time.Now(),
	}

	// Validate inputs
	if err := newFeed.Validate(); err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	// Save the new Feed
	err := models.AddFeed(r.Context(), newFeed)
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) feedDeleteHandler(w http.ResponseWriter, r *http.Request) {
	err := models.DeleteFeed(r.Context(), r.FormValue("id"))
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) messageAddHandler(w http.ResponseWriter, r *http.Request) {
	newMessage := models.Message{
		Trigger:  r.FormValue("trigger"),
		Response: r.FormValue("response"),
	}

	// Validate inputs
	if err := newMessage.Validate(); err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	// Save the new Message
	err := models.AddMessage(r.Context(), newMessage)
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) messageDeleteHandler(w http.ResponseWriter, r *http.Request) {
	err := models.DeleteMessage(r.Context(), r.FormValue("id"))
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}
