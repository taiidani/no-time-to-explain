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
	"github.com/getsentry/sentry-go"
	"github.com/taiidani/no-time-to-explain/internal"
	"github.com/taiidani/no-time-to-explain/internal/bot"
	"github.com/taiidani/no-time-to-explain/internal/data"
	"github.com/taiidani/no-time-to-explain/internal/destiny"
	"github.com/taiidani/no-time-to-explain/internal/models"
	"github.com/taiidani/no-time-to-explain/internal/server"
)

func main() {
	flag.Parse()

	// Handle signal interrupts.
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer cancel()

	teardown, err := initSentry()
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}

	// Flush buffered Sentry events before the program terminates.
	defer teardown()

	// Set up the Redis/Memory cache
	cache := data.NewCache()
	bot.InitCache(cache)

	// Set up the relational database
	err = models.InitDB(ctx)
	if err != nil {
		log.Fatalf("database init: %s", err)
	}

	// Oh hey Bungie
	d2 := destiny.NewTokenClient(cache, os.Getenv("BUNGIE_API_KEY"))
	bot.InitDestinyClient(d2)

	// Set up the Discord client
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("Please set a DISCORD_TOKEN environment variable to your bot token")
	}

	d, err := discordgo.New("Bot " + token)
	if err != nil {
		sentry.CaptureException(err)
		log.Fatal(err)
	}

	// Handle the arguments
	switch flag.Arg(0) {
	case "refresh":
		err := internal.Refresh(ctx, d2, d)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Refresh successful")
	default:
		wg := sync.WaitGroup{}

		wg.Add(1)
		go func() {
			defer wg.Done()
			// Start the web UI
			if err := initServer(ctx, cache, d); err != nil {
				log.Fatal(err)
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			// Start the Discord bot
			if err := initBot(ctx, cache, d); err != nil {
				log.Fatal(err)
			}
		}()

		wg.Wait()

		fmt.Println("Shutdown successful")
	}
}

func initSentry() (func(), error) {
	// Set up Sentry
	err := sentry.Init(sentry.ClientOptions{
		SampleRate:       1.0,
		EnableTracing:    true,
		TracesSampleRate: 1.0,
	})
	if err != nil {
		return func() {}, err
	}

	return func() {
		sentry.Flush(2 * time.Second)
	}, nil
}

func initBot(ctx context.Context, cache data.Cache, b *discordgo.Session) error {
	commands := bot.NewCommands(b, cache)
	commands.AddHandlers()
	defer commands.Teardown()

	// Begin listening for events
	err := b.Open()
	if err != nil {
		sentry.CaptureException(err)
		return fmt.Errorf("could not connect to discord: %w", err)
	}
	defer b.Close()

	// Wait until the application is shutting down
	fmt.Println("Bot is now running. Check out Discord!")
	<-ctx.Done()
	log.Println("Bot shutdown successful")
	return nil
}

func initServer(ctx context.Context, cache data.Cache, b *discordgo.Session) error {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("Required PORT environment variable not present")
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
