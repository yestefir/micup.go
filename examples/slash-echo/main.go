package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	micup "github.com/yestefir/micup.go"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	token := strings.TrimSpace(os.Getenv("MICUP_BOT_TOKEN"))
	guildID := strings.TrimSpace(os.Getenv("MICUP_GUILD_ID"))
	if token == "" || guildID == "" {
		return errors.New("set MICUP_BOT_TOKEN and MICUP_GUILD_ID before running this example")
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	b, err := micup.New(ctx, micup.Config{Token: token})
	if err != nil {
		return fmt.Errorf("create bot: %w", err)
	}
	maxLength := 1000
	command, err := b.REST.CreateGuildCommand(ctx, b.ApplicationID, guildID, micup.ApplicationCommand{
		Name:        "echo",
		Description: "Repeat text privately",
		Options: []micup.CommandOption{
			{Type: micup.OptionTypeString, Name: "text", Description: "What to say", Required: true, MaxLength: &maxLength},
		},
	})
	if err != nil {
		return fmt.Errorf("register echo command: %w", err)
	}
	mux := micup.NewEventMux()
	mux.On(micup.EventInteractionCreate, micup.Listen(func(ctx context.Context, interaction micup.Interaction) error {
		if !interaction.IsCommand() || interaction.GuildID != guildID {
			return nil
		}
		data, ok := interaction.CommandData()
		if !ok || data.ID != command.ID {
			return nil
		}
		text, ok := data.String("text")
		if !ok || strings.TrimSpace(text) == "" {
			text = "Please provide non-empty text."
		}
		return b.REST.RespondToInteraction(ctx, interaction, micup.MessageOptions{Content: text}, true)
	}))
	s, err := b.NewSession(micup.SessionOptions{Handler: mux.Handle})
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	log.Print("Starting slash bot; use /echo text:hello in your test guild. Ctrl+C stops it.")
	if err := s.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("run session: %w", err)
	}
	return nil
}
