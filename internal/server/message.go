package server

import (
	"net/http"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/taiidani/no-time-to-explain/internal/models"
)

func (s *Server) messageGetHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	message, err := models.GetMessage(r.Context(), id)
	if err != nil {
		errorResponse(r.Context(), w, http.StatusBadRequest, err)
		return
	}

	template := "fragment_message.gohtml"
	renderHtml(w, http.StatusOK, template, message)
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

func (s *Server) messageEditHandler(w http.ResponseWriter, r *http.Request) {
	newMessage := models.Message{
		ID:       r.FormValue("id"),
		Trigger:  r.FormValue("trigger"),
		Response: r.FormValue("response"),
	}

	// Validate inputs
	if err := newMessage.Validate(); err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	// Save the Message
	err := models.UpdateMessage(r.Context(), newMessage)
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

func (s *Server) messageSendHandler(w http.ResponseWriter, r *http.Request) {
	channelID := r.FormValue("channel")
	message := r.FormValue("message")

	_, err := s.discord.ChannelMessageSend(channelID, message, discordgo.WithContext(r.Context()))
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}
