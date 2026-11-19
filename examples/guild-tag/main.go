package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

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
	mux := micup.NewEventMux()
	mux.OnGuildTagChange(func(ctx context.Context, change micup.GuildTagChange) error {
		if change.Cleared() {
			log.Printf("user %s removed the tag of guild %s", change.UserID, change.Previous)
			return nil
		}
		log.Printf("user %s now shows the tag of guild %s", change.UserID, change.Current)
		return nil
	})
	mux.On(micup.EventGuildUpdate, micup.Listen(func(ctx context.Context, update micup.GuildUpdate) error {
		guild := update.Guild
		if update.Properties != nil {
			guild = *update.Properties
		}
		if guild.GuildTag != nil {
			log.Printf("guild %s uses the tag %q", update.ID, *guild.GuildTag)
		}
		return nil
	}))
	s, err := b.NewSession(micup.SessionOptions{Handler: micup.Safe(mux.Handle, micup.SafeOptions{Timeout: 30 * time.Second})})
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	log.Print("Listening for guild tag changes. Ctrl+C stops it.")
	if err := s.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("run session: %w", err)
	}
	return nil
}
