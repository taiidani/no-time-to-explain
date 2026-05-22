# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

"No Time To Explain" is a Discord bot written in Go that provides timestamp utilities, Destiny 2 clan leaderboards, and Bluesky feed integration. The application consists of two concurrent services:
- **Discord Bot**: Handles slash commands and user interactions
- **HTTP Server**: Provides OAuth authentication and a web UI for managing bot features

## Build & Run Commands

The project uses [mise](https://mise.jdx.dev) for task running and tool management:

```bash
# Run the application (builds, starts Docker services, runs app)
mise

# Build only (runs tests and linting first)
mise build

# Run tests
mise test

# Run linting (go vet + staticcheck)
mise lint

# Refresh data from external APIs (Destiny/Bluesky)
mise refresh
# OR: ./no-time-to-explain refresh

# Seed the database
mise seed
```

### Direct Go Commands

```bash
# Run without mise
go run main.go

# Run the refresh job
go run main.go refresh

# Run tests with race detection
go test -race -covermode=atomic ./...

# Run specific test
go test -run TestName ./internal/bot/

# Database migrations (via goose tool)
go tool goose -dir ./internal/db/migrations up
go tool goose -dir ./internal/db/migrations status
```

## Required Environment Variables

Set these in `.env` (loaded automatically by mise):
- `DISCORD_TOKEN`: Bot token from Discord Developer Portal
- `DATABASE_URL`: PostgreSQL connection string (default: `postgresql://postgres:root@127.0.0.1:5432/postgres`)
- `PORT`: HTTP server port (default: `3000`)
- `REDIS_HOST` / `REDIS_PORT`: Redis connection details (default: `localhost:6379`)
- `BLUESKY_FEED_CHANNEL_ID`: Discord channel ID for Bluesky posts
- `SENTRY_DSN` / `SENTRY_ENVIRONMENT`: Sentry error tracking (optional)

## Architecture Overview

### Dual-Service Design

The application runs two services concurrently via goroutines in `main.go`:
1. **Discord Bot** (`initBot`): Listens for Discord interactions using discordgo
2. **HTTP Server** (`initServer`): Serves web UI and handles OAuth callbacks

Both services share:
- Redis cache (via `taiidani/go-lib/cache`)
- PostgreSQL database (via `internal/models`)
- Destiny API client (via `internal/destiny`)

### Directory Structure

```
internal/
├── bot/              # Discord bot handlers and commands
│   ├── commands.go   # Command registration and routing
│   ├── time.go       # /time command for timestamp utilities
│   ├── event-calendar.go # Context menu for calendar exports
│   └── messages.go   # Scheduled message management
├── server/           # HTTP server for web UI
│   ├── server.go     # Routes and middleware
│   ├── session.go    # OAuth and session management
│   └── templates/    # HTML templates (embedded in binary)
├── bluesky/          # Bluesky API integration
├── models/           # Database models and queries
├── db/               # Database schema
│   ├── migrations/   # Goose SQL migrations
│   └── seeds/        # Seed data
└── refresh.go        # Background job for syncing external data
```

### Command Architecture

Discord commands are defined as `applicationCommand` structs in `internal/bot/commands.go`:
- `Command`: The Discord application command definition
- `Handler`: Main interaction handler function
- `MessageComponents`: Map of custom IDs to component handlers (buttons, modals)
- `Autocomplete`: Optional autocomplete function

Commands are registered globally on bot startup via `handleReady`. All interactions flow through `handleCommand` which routes to the appropriate handler based on interaction type.

### Database Patterns

- **Migrations**: Goose migrations in `internal/db/migrations/` are automatically applied on startup
- **Models**: Database access is centralized in `internal/models/` with functions like `BulkUpdatePlayers`, `LoadFeeds`
- **Connection**: Single `sql.DB` instance initialized in `models.InitDB()` from `DATABASE_URL`

### External API Integration

#### Bluesky
- Client: `bluesky.Client` in `internal/bluesky/bluesky.go`
- Refresh: Posts new feed items to Discord channel (configured via `BLUESKY_FEED_CHANNEL_ID`)

### Caching Strategy

Redis cache (or memory fallback) is used for:
- Session data (HTTP server)
- Bot interaction state

The cache is initialized once in `main.go` and shared across services.

## Testing

Tests are located alongside the code they test:
- `internal/bot/*_test.go`: Bot command tests
- `internal/destiny/*_test.go`: API helper tests
- `internal/refresh_test.go`: Refresh job tests

Run with race detection enabled (required for CI):
```bash
go test -race -covermode=atomic ./...
```

## CI/CD Pipeline

GitHub Actions workflow in `.github/workflows/build.yml`:
1. **Build**: Compiles binary via `mise build`, packages as `.tgz`
2. **Test**: Runs `go vet`, `staticcheck`, and `go test`
3. **Upload**: Publishes artifact to deployment system
4. **Deploy**: Deploys to production (main branch only)

## Development Notes

- The application uses structured logging via `log/slog` (JSON in production, stderr in dev)
- Sentry integration captures errors and performance metrics
- Database schema is embedded in the binary via `//go:embed` directives
- Templates are embedded but can be loaded from filesystem in dev mode (`DEV=true`)
- Goose migrations run automatically on startup (no manual migration step needed)
- The `refresh` command is intended to be run periodically (e.g., via cron) to sync external data
