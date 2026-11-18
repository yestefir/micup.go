package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	micup "github.com/yestefir/micup.go"
	"github.com/yestefir/micup.go/voice"
)

type playback struct {
	update micup.VoiceServerUpdate
	room   voice.Room
	player *voice.Player
	cancel context.CancelFunc
}

type bot struct {
	b        *micup.Bot
	session  *micup.Session
	mu       sync.Mutex
	active   map[string]*playback
	ffmpeg   string
	maxIdle  time.Duration
	idleStop map[string]*time.Timer
}

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
	maxLength := 2000
	_, err = b.REST.SetGuildCommands(ctx, "", guildID, []micup.ApplicationCommand{
		{
			Name:        "play",
			Description: "Play a file or URL in your voice channel",
			Options: []micup.CommandOption{
				{Type: micup.OptionTypeString, Name: "source", Description: "File path or http(s) URL", Required: true, MaxLength: &maxLength},
			},
		},
		{Name: "pause", Description: "Pause playback"},
		{Name: "resume", Description: "Resume playback"},
		{Name: "skip", Description: "Skip the current track"},
		{Name: "stop", Description: "Leave the voice channel"},
	})
	if err != nil {
		return fmt.Errorf("register commands: %w", err)
	}

	app := &bot{b: b, active: map[string]*playback{}, idleStop: map[string]*time.Timer{}, maxIdle: 30 * time.Second, ffmpeg: os.Getenv("FFMPEG_PATH")}
	router := micup.NewEventRouter()
	router.Command("play", app.play)
	router.Command("pause", app.control(func(p *voice.Player) { p.Pause() }, "Paused."))
	router.Command("resume", app.control(func(p *voice.Player) { p.Resume() }, "Resumed."))
	router.Command("skip", app.control(func(p *voice.Player) { p.Skip() }, "Skipped."))
	router.Command("stop", app.stop)
	mux := micup.NewEventMux()
	mux.On(micup.EventInteractionCreate, router.Listener())

	app.session, err = b.NewSession(micup.SessionOptions{
		Handler:      mux.Handle,
		Logger:       slog.Default(),
		OnVoiceLost:  func(guildID, connectionID string) { app.leave(guildID, connectionID) },
		OnVoiceEmpty: app.scheduleLeave,
	})
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	log.Print("Starting voice player; join a voice channel and use /play in your test guild. Ctrl+C stops it.")
	err = app.session.Run(ctx)
	app.leaveAll()
	if err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("run session: %w", err)
	}
	return nil
}

func (a *bot) reply(ctx context.Context, i micup.Interaction, text string) error {
	return a.b.REST.RespondToInteraction(ctx, i, micup.MessageOptions{Content: text}, true)
}

func (a *bot) play(ctx context.Context, i micup.Interaction) error {
	data, _ := i.CommandData()
	source, _ := data.String("source")
	invoker := i.Invoker()
	if invoker == nil || i.GuildID == "" {
		return a.reply(ctx, i, "Use this command in a guild.")
	}
	channelID := a.session.State().VoiceChannel(i.GuildID, invoker.ID)
	if channelID == "" {
		return a.reply(ctx, i, "Join a voice channel first.")
	}
	if err := a.b.REST.DeferInteraction(ctx, i, true); err != nil {
		return err
	}
	if err := a.start(i.GuildID, channelID, source); err != nil {
		_, editErr := a.b.REST.EditInteractionResponse(ctx, i, micup.MessageEdit{Content: new(err.Error())})
		return errors.Join(err, editErr)
	}
	_, err := a.b.REST.EditInteractionResponse(ctx, i, micup.MessageEdit{Content: new("Playing.")})
	return err
}

func (a *bot) start(guildID, channelID, source string) error {
	a.cancelIdle(guildID)
	a.mu.Lock()
	current := a.active[guildID]
	a.mu.Unlock()
	if current != nil {
		current.player.Skip()
	}
	if current == nil {
		var err error
		if current, err = a.join(guildID, channelID); err != nil {
			return err
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	pcm, err := voice.OpenSource(ctx, source, voice.SourceOptions{FFmpeg: a.ffmpeg})
	if err != nil {
		cancel()
		return err
	}
	go func() {
		defer cancel()
		defer pcm.Close()
		if err := current.player.Play(ctx, pcm); err != nil && !errors.Is(err, voice.ErrSkipped) && !errors.Is(err, context.Canceled) {
			slog.Warn("playback ended with an error", "guild", guildID, "error", err)
		}
	}()
	return nil
}

func (a *bot) join(guildID, channelID string) (*playback, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	update, err := a.session.JoinVoice(ctx, guildID, channelID, micup.VoiceJoin{SelfDeaf: true})
	if err != nil {
		return nil, fmt.Errorf("join voice: %w", err)
	}
	room, err := (voice.LiveKit{FFmpeg: a.ffmpeg, Logger: slog.Default()}).Connect(ctx, voice.Grant{Endpoint: update.Endpoint, Token: update.Token})
	if err != nil {
		_ = a.session.LeaveVoice(guildID, update.ConnectionID)
		return nil, err
	}
	track, err := room.OpenTrack(context.Background())
	if err != nil {
		room.Close()
		_ = a.session.LeaveVoice(guildID, update.ConnectionID)
		return nil, err
	}
	p := &playback{update: update, room: room, player: voice.NewPlayer(track), cancel: func() { _ = track.Close() }}
	a.mu.Lock()
	a.active[guildID] = p
	a.mu.Unlock()
	return p, nil
}

func (a *bot) control(apply func(*voice.Player), done string) micup.InteractionHandler {
	return func(ctx context.Context, i micup.Interaction) error {
		a.mu.Lock()
		p := a.active[i.GuildID]
		a.mu.Unlock()
		if p == nil {
			return a.reply(ctx, i, "Nothing is playing.")
		}
		apply(p.player)
		return a.reply(ctx, i, done)
	}
}

func (a *bot) stop(ctx context.Context, i micup.Interaction) error {
	a.cancelIdle(i.GuildID)
	a.mu.Lock()
	p := a.active[i.GuildID]
	a.mu.Unlock()
	if p == nil {
		return a.reply(ctx, i, "Not in a voice channel.")
	}
	_ = a.session.LeaveVoice(i.GuildID, p.update.ConnectionID)
	a.leave(i.GuildID, p.update.ConnectionID)
	return a.reply(ctx, i, "Left the voice channel.")
}

func (a *bot) leave(guildID, connectionID string) {
	a.mu.Lock()
	p := a.active[guildID]
	if p == nil || (connectionID != "" && p.update.ConnectionID != connectionID) {
		a.mu.Unlock()
		return
	}
	delete(a.active, guildID)
	a.mu.Unlock()
	p.player.Skip()
	p.cancel()
	p.room.Close()
}

func (a *bot) leaveAll() {
	a.mu.Lock()
	guilds := make([]string, 0, len(a.active))
	for guildID := range a.active {
		guilds = append(guilds, guildID)
	}
	a.mu.Unlock()
	for _, guildID := range guilds {
		a.leave(guildID, "")
	}
}

func (a *bot) scheduleLeave(guildID, _ string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.active[guildID] == nil || a.idleStop[guildID] != nil {
		return
	}
	a.idleStop[guildID] = time.AfterFunc(a.maxIdle, func() {
		a.mu.Lock()
		delete(a.idleStop, guildID)
		p := a.active[guildID]
		a.mu.Unlock()
		if p != nil {
			_ = a.session.LeaveVoice(guildID, p.update.ConnectionID)
			a.leave(guildID, p.update.ConnectionID)
		}
	})
}

func (a *bot) cancelIdle(guildID string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if timer := a.idleStop[guildID]; timer != nil {
		timer.Stop()
		delete(a.idleStop, guildID)
	}
}
