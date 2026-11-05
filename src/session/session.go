package session

import (
	"context"
	"errors"
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
	BeforeIdentify func(context.Context) error
}

type Options struct {
	Handler     gateway.Handler
	OnVoiceLost func(guildID, connectionID string)
	UserAgent   string
	Browser     string
	MinBackoff  time.Duration
	MaxBackoff  time.Duration
}

type Session struct {
	config  Config
	options Options
	cache   *state.Cache
	resume  gateway.Resume
	ready   atomic.Bool
	mu      sync.Mutex
	conn    *gateway.Conn
	grants  map[string]model.VoiceServerUpdate
	pending map[string]chan model.VoiceServerUpdate
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
	if options.MinBackoff <= 0 {
		options.MinBackoff = DefaultMinBackoff
	}
	if options.MaxBackoff < options.MinBackoff {
		options.MaxBackoff = max(DefaultMaxBackoff, options.MinBackoff)
	}
	return &Session{
		config:  config,
		options: options,
		cache:   state.New(config.BotID),
		grants:  map[string]model.VoiceServerUpdate{},
		pending: map[string]chan model.VoiceServerUpdate{},
	}, nil
}

func (s *Session) State() *state.Cache { return s.cache }

func (s *Session) Ready() bool { return s.ready.Load() }

func (s *Session) ShardID() int { return s.config.ShardID }

func (s *Session) Run(ctx context.Context) error {
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

func (s *Session) Send(opcode int, data any) error {
	s.mu.Lock()
	conn := s.conn
	s.mu.Unlock()
	if conn == nil {
		return gateway.ErrClosed
	}
	return conn.Send(opcode, data)
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
		URL:            s.config.Endpoint,
		Token:          s.config.Token,
		ShardID:        s.config.ShardID,
		ShardCount:     s.config.ShardCount,
		UserAgent:      s.options.UserAgent,
		Browser:        s.options.Browser,
		Resume:         &s.resume,
		OnConnect:      s.attach,
		OnDisconnect:   func() { s.attach(nil) },
		BeforeIdentify: s.config.BeforeIdentify,
	}
}

func (s *Session) attach(conn *gateway.Conn) {
	s.mu.Lock()
	s.conn = conn
	s.mu.Unlock()
}

func (s *Session) dispatch(ctx context.Context, event gateway.Event) error {
	move := s.cache.Apply(event.Name, event.Data)
	switch event.Name {
	case gateway.EventReady:
		s.mu.Lock()
		s.grants = map[string]model.VoiceServerUpdate{}
		s.mu.Unlock()
		s.ready.Store(true)
	case gateway.EventResumed:
		s.ready.Store(true)
	case gateway.EventVoiceStateUpdate:
		s.handleBotVoiceMove(move)
	case gateway.EventVoiceServerUpdate:
		s.handleVoiceServerUpdate(event)
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
