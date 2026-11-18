package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"

	micup "github.com/yestefir/micup.go"
)

const countPrefix = "panel:count:"

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
	_, err = b.REST.SetGuildCommands(ctx, b.ApplicationID, guildID, []micup.ApplicationCommand{
		{Name: "panel", Description: "Post a counter panel"},
		{Name: "feedback", Description: "Send feedback"},
	})
	if err != nil {
		return fmt.Errorf("register commands: %w", err)
	}

	router := micup.NewEventRouter()
	router.Command("panel", func(ctx context.Context, i micup.Interaction) error {
		return b.REST.RespondToInteraction(ctx, i, panel(0), false)
	})
	router.Component(countPrefix+"*", func(ctx context.Context, i micup.Interaction) error {
		current, err := strconv.Atoi(strings.TrimPrefix(i.CustomID(), countPrefix))
		if err != nil {
			return err
		}
		options := panel(current + 1)
		return b.REST.UpdateInteractionMessage(ctx, i, micup.MessageEdit{Content: &options.Content, Components: &options.Components})
	})
	router.Command("feedback", func(ctx context.Context, i micup.Interaction) error {
		return b.REST.RespondWithModal(ctx, i, micup.Modal{
			CustomID: "feedback",
			Title:    "Feedback",
			Inputs:   []micup.TextInput{micup.ParagraphInput("message", "What should we improve?")},
		})
	})
	router.Modal("feedback", func(ctx context.Context, i micup.Interaction) error {
		data, _ := i.ModalData()
		message, _ := data.Value("message")
		log.Printf("feedback: %q", message)
		return b.REST.RespondToInteraction(ctx, i, micup.MessageOptions{Content: "Thanks for the feedback."}, true)
	})

	mux := micup.NewEventMux()
	mux.On(micup.EventInteractionCreate, router.Listener())
	s, err := b.NewSession(micup.SessionOptions{Handler: mux.Handle})
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	log.Print("Starting panel bot; use /panel and /feedback in your test guild. Ctrl+C stops it.")
	if err := s.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("run session: %w", err)
	}
	return nil
}

func panel(count int) micup.MessageOptions {
	return micup.MessageOptions{
		Content: fmt.Sprintf("Clicked %d times", count),
		Components: []micup.ActionRow{
			micup.NewActionRow(micup.PrimaryButton(countPrefix+strconv.Itoa(count), "Click me")),
		},
	}
}
