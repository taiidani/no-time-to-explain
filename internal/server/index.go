package server

import (
	"fmt"
	"net/http"

	"github.com/taiidani/no-time-to-explain/internal/models"
)

type indexBag struct {
	baseBag
	Messages []models.Message
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	bag := indexBag{baseBag: s.newBag(r)}

	messages, err := models.LoadMessages(r.Context())
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}
	bag.Messages = messages

	template := "index.gohtml"
	renderHtml(w, http.StatusOK, template, bag)
}

func (s *Server) indexPostHandler(w http.ResponseWriter, r *http.Request) {
	newMessage := models.Message{
		Trigger:  r.FormValue("trigger"),
		Response: r.FormValue("response"),
	}

	// Validate inputs
	if len(newMessage.Trigger) < 4 || len(newMessage.Response) < 4 {
		errorResponse(r.Context(), w, http.StatusInternalServerError, fmt.Errorf("provided inputs need to be at least 4 characters"))
		return
	}

	// Save the new message
	err := models.AddMessage(r.Context(), newMessage)
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) indexDeleteHandler(w http.ResponseWriter, r *http.Request) {
	err := models.DeleteMessage(r.Context(), r.FormValue("id"))
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}
