package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type errorBag struct {
	baseBag
	Title   string
	Message error
}

func (s *Server) errorNotFoundHandler(w http.ResponseWriter, r *http.Request) {
	err := errors.New("this page does not exist")
	data := errorBag{
		Title:   "404 Page Not Found",
		Message: err,
	}

	slog.Warn("Displaying error page", "error", err)
	renderHtml(w, http.StatusNotFound, "error.gohtml", data)
}

func errorResponse(ctx context.Context, writer http.ResponseWriter, code int, err error) {
	title := "Error"
	switch code {
	case http.StatusNotFound:
		title = "404 Page Not Found"
	case http.StatusInternalServerError:
		title = "500 Internal Server Error"
	case http.StatusBadRequest:
		title = "400 Bad Request"
	}

	data := errorBag{
		Title:   title,
		Message: err,
	}

	// Mark the active request span as errored so the failure surfaces in the
	// trace backend.
	span := trace.SpanFromContext(ctx)
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())

	slog.ErrorContext(ctx, "Displaying error page", "error", err)
	renderHtml(writer, code, "error.gohtml", data)
}
