package session

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/yestefir/micup.go/src/gateway"
	"github.com/yestefir/micup.go/src/model"
	"github.com/yestefir/micup.go/src/state"
)

const (
	DefaultMinBackoff = time.Second
	DefaultMaxBackoff = 15 * time.Second
)

var ErrNotReady = errors.New("Micup Gateway session is not ready")

type ChannelLookup interface {
	Channel(ctx context.Context, channelID string) (model.Channel, error)
}

type Config struct {
	Endpoint       string
	Token          string
	BotID          string
	ShardID        int
	ShardCount     int
	Channels       ChannelLookup
	Members        MemberLookup
	BeforeIdentify func(context.Context) error
}

type Options struct {
	Handler       gateway.Handler
	TrackMembers  bool
	OnVoiceLost   func(guildID, connectionID string)
	OnVoiceGrant  func(grant model.VoiceServerUpdate)
	IgnoredEvents []string
	Presence      *model.Presence
	Flags         int
	Compress      bool
	UserAgent     string
	Browser       string
	MinBackoff    time.Duration
	MaxBackoff    time.Duration
}

type Session struct {
	inventory      memberInventory
	config         Config
	options        Options
	cache          *state.Cache
	resume         gateway.Resume
	ready          atomic.Bool
	mu             sync.Mutex
	conn           *gateway.Conn
	grants         map[string]model.VoiceServerUpdate
	pending        map[string]chan joinResult
	joinOrder      []string
	presence       *model.Presence
	memberRequests map[string]chan memberChunk
}

func New(config Config, options Options) (*Session, error) {
	if config.Endpoint == "" || config.Token == "" {
		return nil, errors.New("session endpoint and token are required")
	}
	if config.ShardCount == 0 {
		config.ShardCount = 1
	}
	if err := gateway.ValidateShard(config.ShardID, config.ShardCount); err != nil {
		return nil, err
	}
	if err := gateway.ValidateIgnoredEvents(options.IgnoredEvents); err != nil {
		return nil, err
	}
	if options.TrackMembers && config.Members == nil {
		return nil, errors.New("member tracking requires a member lookup")
	}
	if options.Presence != nil && !model.ValidStatus(options.Presence.Status) {
		return nil, fmt.Errorf("invalid presence status %q", options.Presence.Status)
	}
	if options.MinBackoff <= 0 {
		options.MinBackoff = DefaultMinBackoff
	}
	if options.MaxBackoff < options.MinBackoff {
		options.MaxBackoff = max(DefaultMaxBackoff, options.MinBackoff)
	}
	return &Session{
		config:         config,
		options:        options,
		cache:          state.New(config.BotID),
		grants:         map[string]model.VoiceServerUpdate{},
		pending:        map[string]chan joinResult{},
		memberRequests: map[string]chan memberChunk{},
		presence:       options.Presence,
	}, nil
}

func (s *Session) State() *state.Cache { return s.cache }

func (s *Session) Ready() bool { return s.ready.Load() }

func (s *Session) ShardID() int { return s.config.ShardID }

func (s *Session) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	var requests sync.WaitGroup
	defer func() { cancel(); requests.Wait() }()
	if s.options.TrackMembers {
		for range 2 {
			requests.Add(1)
			go func() { defer requests.Done(); s.enumerateMembers(ctx) }()
		}
	}
	delay := s.options.MinBackoff
	for ctx.Err() == nil {
		established := false
		err := gateway.Run(ctx, s.gatewayOptions(), func(ctx context.Context, event gateway.Event) error {
			if event.Name == gateway.EventReady || event.Name == gateway.EventResumed {
				established = true
			}
			return s.dispatch(ctx, event)
		})
		s.ready.Store(false)
		var closed *gateway.CloseError
		if errors.As(err, &closed) && closed.Fatal() {
			return closed
		}
		if established {
			delay = s.options.MinBackoff
		}
		if sleep(ctx, delay) != nil {
			return nil
		}
		delay = min(delay*2, s.options.MaxBackoff)
	}
	return nil
}

func (s *Session) Send(ctx context.Context, opcode int, data any) error {
	s.mu.Lock()
	conn := s.conn
	s.mu.Unlock()
	if conn == nil {
		return gateway.ErrClosed
	}
	return conn.Send(ctx, opcode, data)
}

func (s *Session) ResolveGuild(ctx context.Context, channelID string) (string, error) {
	if guild, ok := s.cache.GuildForChannel(channelID); ok {
		return guild, nil
	}
	if s.config.Channels == nil {
		return "", errors.New("channel is not cached and no channel lookup is configured")
	}
	channel, err := s.config.Channels.Channel(ctx, channelID)
	if err != nil {
		return "", err
	}
	s.cache.Remember(channelID, channel.GuildID)
	return channel.GuildID, nil
}

func (s *Session) gatewayOptions() gateway.Options {
	return gateway.Options{
		URL:          s.config.Endpoint,
		Token:        s.config.Token,
		ShardID:      s.config.ShardID,
		ShardCount:   s.config.ShardCount,
		UserAgent:    s.options.UserAgent,
		Browser:      s.options.Browser,
		Resume:       &s.resume,
		OnConnect:    s.attach,
		OnDisconnect: func() { s.attach(nil) },
		OnSessionInvalidated: func() {
			s.ready.Store(false)
			s.dropAllGrants()
		},
		OnGatewayError: s.handleGatewayError,
		BeforeIdentify: s.config.BeforeIdentify,
		IgnoredEvents:  s.options.IgnoredEvents,
		PresenceFunc:   s.currentPresence,
		Flags:          s.options.Flags,
		Compress:       s.options.Compress,
	}
}

func (s *Session) attach(conn *gateway.Conn) {
	s.mu.Lock()
	s.conn = conn
	s.mu.Unlock()
}

func (s *Session) dispatch(ctx context.Context, event gateway.Event) error {
	if s.options.TrackMembers {
		s.inventory.observe(event.Name, event.Data)
	}
	move := s.cache.Apply(event.Name, event.Data)
	switch event.Name {
	case gateway.EventReady:
		s.dropAllGrants()
		s.ready.Store(true)
	case gateway.EventResumed:
		s.ready.Store(true)
	case gateway.EventVoiceStateUpdate:
		s.handleBotVoiceMove(move)
	case gateway.EventVoiceServerUpdate:
		s.handleVoiceServerUpdate(event)
	case gateway.EventGuildMembersChunk:
		s.handleMembersChunk(event)
	case gateway.EventRateLimited:
		s.handleRateLimited(event)
	}
	if s.options.Handler == nil {
		return nil
	}
	return s.options.Handler(ctx, event)
}

func sleep(ctx context.Context, wait time.Duration) error {
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (s *Session) UpdatePresence(ctx context.Context, presence model.Presence) error {
	if !model.ValidStatus(presence.Status) {
		return fmt.Errorf("invalid presence status %q", presence.Status)
	}
	s.mu.Lock()
	s.presence = &presence
	s.mu.Unlock()
	return s.Send(ctx, gateway.OpPresenceUpdate, presence)
}

func (s *Session) currentPresence() *model.Presence {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.presence == nil {
		return nil
	}
	presence := *s.presence
	return &presence
}
