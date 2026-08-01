package server

import (
	"net/http"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/taiidani/no-time-to-explain/internal/db/models"
)

func (s *Server) messageGetHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 32)
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	message, err := s.queries.GetMessage(r.Context(), int32(id))
	if err != nil {
		errorResponse(r.Context(), w, http.StatusBadRequest, err)
		return
	}

	template := "fragment_message.gohtml"
	renderHtml(w, http.StatusOK, template, message)
}

func (s *Server) messageAddHandler(w http.ResponseWriter, r *http.Request) {
	newMessage := models.Message{
		Enabled:  r.FormValue("enabled") == "enabled",
		Sender:   r.FormValue("sender"),
		Trigger:  r.FormValue("trigger"),
		Response: r.FormValue("response"),
	}

	// Validate inputs
	if err := s.queries.ValidateMessage(newMessage); err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	// Save the new Message
	_, err := s.queries.CreateMessage(r.Context(), models.CreateMessageParams{
		Enabled:  newMessage.Enabled,
		Sender:   newMessage.Sender,
		Trigger:  newMessage.Trigger,
		Response: newMessage.Response,
	})
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) messageEditHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.FormValue("id"), 10, 32)
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	newMessage := models.Message{
		ID:       int32(id),
		Enabled:  r.FormValue("enabled") == "enabled",
		Sender:   r.FormValue("sender"),
		Trigger:  r.FormValue("trigger"),
		Response: r.FormValue("response"),
	}

	// Validate inputs
	if err := s.queries.ValidateMessage(newMessage); err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	// Save the Message
	_, err = s.queries.UpdateMessage(r.Context(), models.UpdateMessageParams{
		ID:       newMessage.ID,
		Enabled:  newMessage.Enabled,
		Sender:   newMessage.Sender,
		Trigger:  newMessage.Trigger,
		Response: newMessage.Response,
	})
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) messageDeleteHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	err = s.queries.DeleteMessage(r.Context(), int32(id))
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
