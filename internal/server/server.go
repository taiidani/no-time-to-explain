package server

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"regexp"

	"github.com/bwmarrin/discordgo"
	"github.com/taiidani/go-lib/authz"
	"github.com/taiidani/go-lib/cache"
	"github.com/taiidani/no-time-to-explain/internal/models"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Server struct {
	backend        cache.Cache
	sessionManager authz.Session
	discord        *discordgo.Session
	publicURL      string
	port           string
	*http.Server
}

//go:embed templates
var templates embed.FS

// DevMode can be toggled to pull rendered files from the filesystem or the embedded FS.
var DevMode = os.Getenv("DEV") == "true"

func NewServer(backend cache.Cache, b *discordgo.Session, port string) *Server {
	mux := http.NewServeMux()

	publicURL := os.Getenv("PUBLIC_URL")
	if publicURL == "" {
		publicURL = "http://localhost:" + port
	}

	sess := authz.NewSession(backend)
	sess.Secure = !DevMode

	srv := &Server{
		Server: &http.Server{
			Addr:    fmt.Sprintf(":%s", port),
			Handler: mux,
		},
		publicURL:      publicURL,
		port:           port,
		backend:        backend,
		discord:        b,
		sessionManager: sess,
	}
	srv.addRoutes(mux)

	return srv
}

func (s *Server) addRoutes(mux *http.ServeMux) {
	// handle registers a route, wrapping its handler so each request produces a
	// server span named after the matched route. otelhttp also extracts any
	// incoming W3C trace context for distributed tracing.
	handle := func(pattern string, h http.Handler) {
		mux.Handle(pattern, otelhttp.NewHandler(h, pattern))
	}

	handle("GET /{$}", s.sessionMiddleware(http.HandlerFunc(s.indexHandler)))
	handle("GET /channels", s.sessionMiddleware(http.HandlerFunc(s.channelsHandler)))
	handle("GET /users", s.sessionMiddleware(http.HandlerFunc(s.usersHandler)))
	handle("GET /auth", http.HandlerFunc(s.auth))
	handle("GET /oauth/callback", http.HandlerFunc(s.authCallback))
	handle("GET /login", http.HandlerFunc(s.login))
	handle("GET /logout", http.HandlerFunc(s.logout))
	handle("POST /feed/add", s.sessionMiddleware(http.HandlerFunc(s.feedAddHandler)))
	handle("POST /feed/delete", s.sessionMiddleware(http.HandlerFunc(s.feedDeleteHandler)))
	handle("POST /message/add", s.sessionMiddleware(http.HandlerFunc(s.messageAddHandler)))
	handle("POST /message/edit", s.sessionMiddleware(http.HandlerFunc(s.messageEditHandler)))
	handle("POST /message/delete", s.sessionMiddleware(http.HandlerFunc(s.messageDeleteHandler)))
	handle("POST /message/send", s.sessionMiddleware(http.HandlerFunc(s.messageSendHandler)))
	handle("GET /message/{id}", s.sessionMiddleware(http.HandlerFunc(s.messageGetHandler)))
	handle("/assets/", http.HandlerFunc(s.assetsHandler))
	handle("/", http.HandlerFunc(s.errorNotFoundHandler))
}

// templateFuncs are made available to every template rendered by renderHtml.
var templateFuncs = template.FuncMap{
	"linkify": linkify,
}

// urlPattern matches bare URLs so they can be rendered as clickable links.
var urlPattern = regexp.MustCompile(`https?://[^\s<>"']+`)

// linkify escapes the given plain text and turns any URLs it contains into
// clickable <a> tags. This allows bot message responses (which are plain
// text, e.g. as sent to Discord) to render as hyperlinks in the web UI
// without allowing arbitrary HTML injection.
func linkify(text string) template.HTML {
	var buf bytes.Buffer

	last := 0
	for _, loc := range urlPattern.FindAllStringIndex(text, -1) {
		start, end := loc[0], loc[1]
		buf.WriteString(template.HTMLEscapeString(text[last:start]))

		url := text[start:end]
		buf.WriteString(`<a href="`)
		buf.WriteString(template.HTMLEscapeString(url))
		buf.WriteString(`" target="_blank" rel="noopener noreferrer">`)
		buf.WriteString(template.HTMLEscapeString(url))
		buf.WriteString(`</a>`)

		last = end
	}
	buf.WriteString(template.HTMLEscapeString(text[last:]))

	return template.HTML(buf.String())
}

func renderHtml(writer http.ResponseWriter, code int, file string, data any) {
	log := slog.With("name", file, "code", code)

	var t *template.Template
	var err error
	if DevMode {
		t, err = template.New("").Funcs(templateFuncs).ParseGlob("internal/server/templates/**")
	} else {
		t, err = template.New("").Funcs(templateFuncs).ParseFS(templates, "templates/**")
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

	if sess, ok := r.Context().Value(sessionKey).(models.Session); ok {
		ret.Username = sess.DiscordUser.Username
	}

	return ret
}
