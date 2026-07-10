package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/taiidani/go-lib/cache"
	"github.com/taiidani/no-time-to-explain/internal"
	"github.com/taiidani/no-time-to-explain/internal/bot"
	"github.com/taiidani/no-time-to-explain/internal/models"
	"github.com/taiidani/no-time-to-explain/internal/server"
	"github.com/taiidani/no-time-to-explain/internal/telemetry"
)

func main() {
	flag.Parse()

	// Handle signal interrupts.
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer cancel()

	// Set up logging. Records are wrapped so that any log emitted with an
	// active span context is annotated with trace_id/span_id for correlation
	// with traces in the observability backend.
	var handler slog.Handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{})
	handler = telemetry.NewTraceHandler(handler)
	slog.SetDefault(slog.New(handler))

	// Set up OpenTelemetry tracing
	shutdownTelemetry, err := telemetry.Init(ctx)
	if err != nil {
		log.Fatalf("telemetry init: %s", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdownTelemetry(shutdownCtx); err != nil {
			slog.Error("telemetry shutdown error", "err", err)
		}
	}()

	// Set up the Redis/Memory cache
	cache, err := cache.NewRedis("no-time-to-explain:")
	if err != nil {
		log.Fatalf("cache init: %s", err)
	}
	bot.InitCache(cache)

	// Set up the relational database
	err = models.InitDB(ctx)
	if err != nil {
		log.Fatalf("database init: %s", err)
	}

	// Set up the Discord client
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("Please set a DISCORD_TOKEN environment variable to your bot token")
	}

	d, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal(err)
	}

	// Handle the arguments
	wg := sync.WaitGroup{}

	wg.Go(func() {
		// Start the web UI
		if err := initServer(ctx, cache, d); err != nil {
			log.Fatal(err)
		}
	})

	wg.Go(func() {
		// Start the Discord bot
		if err := initBot(ctx, cache, d); err != nil {
			log.Fatal(err)
		}
	})

	wg.Go(func() {
		// Start the Refresh loop
		for {
			select {
			case <-ctx.Done():
				slog.Info("Refresh loop shutting down")
				return
			case <-time.After(5 * time.Minute):
				err := internal.Refresh(ctx, d)
				if err != nil {
					log.Fatal(err)
				}

				slog.Info("Refresh successful")
			}
		}
	})

	wg.Wait()

	slog.Info("Shutdown successful")
}

func initBot(ctx context.Context, cache cache.Cache, b *discordgo.Session) error {
	commands := bot.NewCommands(b, cache)
	commands.AddHandlers()
	defer commands.Teardown()

	// Begin listening for events
	err := b.Open()
	if err != nil {
		return fmt.Errorf("could not connect to discord: %w", err)
	}
	defer b.Close()

	// Wait until the application is shutting down
	slog.Info("Bot is now running. Check out Discord!")
	<-ctx.Done()
	slog.Info("Bot shutdown successful")
	return nil
}

func initServer(ctx context.Context, cache cache.Cache, b *discordgo.Session) error {
	port := os.Getenv("PORT")
	if port == "" {
		return fmt.Errorf("required PORT environment variable not present")
	}

	srv := server.NewServer(cache, b, port)

	go func() {
		slog.Info("Server starting", "port", port)
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			slog.Error("Unclean server shutdown encountered", "error", err)
		}
	}()

	<-ctx.Done()

	// Gracefully shut down over 60 seconds
	slog.Info("Server shutting down")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Minute)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}

	slog.Info("Server shutdown successful")
	return nil
}
