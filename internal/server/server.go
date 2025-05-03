package server

import (
	"embed"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"

	"github.com/bwmarrin/discordgo"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/taiidani/no-time-to-explain/internal/data"
	"github.com/taiidani/no-time-to-explain/internal/models"
)

type Server struct {
	backend   data.Cache
	discord   *discordgo.Session
	publicURL string
	port      string
	*http.Server
}

//go:embed templates
var templates embed.FS

// DevMode can be toggled to pull rendered files from the filesystem or the embedded FS.
var DevMode = os.Getenv("DEV") == "true"

func NewServer(backend data.Cache, b *discordgo.Session, port string) *Server {
	mux := http.NewServeMux()

	publicURL := os.Getenv("PUBLIC_URL")
	if publicURL == "" {
		publicURL = "http://localhost:" + port
	}

	srv := &Server{
		Server: &http.Server{
			Addr:    fmt.Sprintf(":%s", port),
			Handler: mux,
		},
		publicURL: publicURL,
		port:      port,
		backend:   backend,
		discord:   b,
	}
	srv.addRoutes(mux)

	return srv
}

func (s *Server) addRoutes(mux *http.ServeMux) {
	sentryHandler := sentryhttp.New(sentryhttp.Options{})

	mux.Handle("GET /{$}", sentryHandler.Handle(s.sessionMiddleware(http.HandlerFunc(s.indexHandler))))
	mux.Handle("GET /auth", sentryHandler.Handle(http.HandlerFunc(s.auth)))
	mux.Handle("GET /oauth/callback", sentryHandler.Handle(http.HandlerFunc(s.authCallback)))
	mux.Handle("GET /login", sentryHandler.Handle(http.HandlerFunc(s.login)))
	mux.Handle("GET /logout", sentryHandler.Handle(http.HandlerFunc(s.logout)))
	mux.Handle("POST /feed/add", sentryHandler.Handle(s.sessionMiddleware(http.HandlerFunc(s.feedAddHandler))))
	mux.Handle("POST /feed/delete", sentryHandler.Handle(s.sessionMiddleware(http.HandlerFunc(s.feedDeleteHandler))))
	mux.Handle("POST /message/add", sentryHandler.Handle(s.sessionMiddleware(http.HandlerFunc(s.messageAddHandler))))
	mux.Handle("POST /message/delete", sentryHandler.Handle(s.sessionMiddleware(http.HandlerFunc(s.messageDeleteHandler))))
	mux.Handle("/assets/", sentryHandler.Handle(http.HandlerFunc(s.assetsHandler)))
	mux.Handle("/", sentryHandler.Handle(http.HandlerFunc(s.errorNotFoundHandler)))
}

func renderHtml(writer http.ResponseWriter, code int, file string, data any) {
	log := slog.With("name", file, "code", code)

	var t *template.Template
	var err error
	if DevMode {
		t, err = template.ParseGlob("internal/server/templates/**")
	} else {
		t, err = template.ParseFS(templates, "templates/**")
	}
	if err != nil {
		log.Error("Could not parse templates", "error", err)
		return
	}

	log.Debug("Rendering file", "dev", DevMode)
	writer.WriteHeader(code)
	err = t.ExecuteTemplate(writer, file, data)
	if err != nil {
		log.Error("Could not render template", "error", err)
	}
}

type baseBag struct {
	Username string
}

func (s *Server) newBag(r *http.Request) baseBag {
	ret := baseBag{}

	if sess, ok := r.Context().Value(sessionKey).(*models.Session); ok {
		ret.Username = sess.DiscordUser.Username
	}

	return ret
}
