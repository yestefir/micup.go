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
	if token == "" {
		return errors.New("set MICUP_BOT_TOKEN before running this example")
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	b, err := micup.New(ctx, micup.Config{Token: token})
	if err != nil {
		return fmt.Errorf("create bot: %w", err)
	}
	if b.MessageContentEnabled != nil && !*b.MessageContentEnabled {
		return errors.New("enable the bot's message-content grant to use !ping")
	}
	mux := micup.NewEventMux()
	mux.On(micup.EventMessageCreate, micup.Listen(func(ctx context.Context, message micup.MessageCreate) error {
		if message.Author.Bot || message.Content != "!ping" {
			return nil
		}
		_, err := b.REST.SendMessage(ctx, message.ChannelID, "pong", message.ID)
		return err
	}))
	s, err := b.NewSession(micup.SessionOptions{Handler: mux.Handle})
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	log.Print("Starting ping bot; send !ping in a channel the bot can read. Ctrl+C stops it.")
	if err := s.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("run session: %w", err)
	}
	return nil
}
