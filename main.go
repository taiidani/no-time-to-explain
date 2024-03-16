package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/taiidani/no-time-to-explain/internal"
)

func main() {
	// Handle signal interrupts.
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer cancel()

	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("Please set a DISCORD_TOKEN environment variable to your bot token")
	}

	b, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal(err)
	}
	defer b.Close()

	internal.InitDB()

	commands := internal.NewCommands(b)
	commands.AddHandlers()
	defer commands.Teardown()

	// Begin listening for events
	err = b.Open()
	if err != nil {
		log.Fatal("Could not connect to discord", err)
	}
	defer b.Close()

	// Wait until the application is shutting down
	fmt.Println("Bot is now running. Check out Discord!")
	<-ctx.Done()
	log.Println("Graceful shutdown")
}
