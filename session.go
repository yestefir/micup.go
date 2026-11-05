package micup

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultMinBackoff = time.Second
	defaultMaxBackoff = 15 * time.Second
)

var ErrNotReady = errors.New("Micup Gateway session is not ready")

type SessionOptions struct {
	Handler     EventHandler
	OnVoiceLost func(guildID, connectionID string)
	UserAgent   string
	Browser     string
	MinBackoff  time.Duration
	MaxBackoff  time.Duration
}

type Session struct {
	client         *Client
	options        SessionOptions
	shardID        int
	shardCount     int
	beforeIdentify func(context.Context) error
	state          *State
	resume         resumeState
	ready          atomic.Bool
	mu             sync.Mutex
	current        *gatewaySession
	grants         map[string]VoiceServerUpdate
	pending        map[string]chan VoiceServerUpdate
}

func (c *Client) NewSession(options SessionOptions) *Session {
	return c.newSession(options, c.config.ShardID, c.config.ShardCount)
}

func (c *Client) newSession(options SessionOptions, shardID, shardCount int) *Session {
	if options.UserAgent == "" {
		options.UserAgent = defaultUserAgent
	}
	if options.Browser == "" {
		options.Browser = defaultBrowser
	}
	if options.MinBackoff <= 0 {
		options.MinBackoff = defaultMinBackoff
	}
	if options.MaxBackoff < options.MinBackoff {
		options.MaxBackoff = max(defaultMaxBackoff, options.MinBackoff)
	}
	return &Session{
		client:     c,
		options:    options,
		shardID:    shardID,
		shardCount: shardCount,
		state:      NewState(c.BotID),
		grants:     map[string]VoiceServerUpdate{},
		pending:    map[string]chan VoiceServerUpdate{},
	}
}

func (s *Session) State() *State { return s.state }

func (s *Session) Ready() bool { return s.ready.Load() }

func (s *Session) ShardID() int { return s.shardID }

func (s *Session) Run(ctx context.Context) error {
	delay := s.options.MinBackoff
	for ctx.Err() == nil {
		established := false
		err := runConnection(ctx, s.client, func(ctx context.Context, event Event) error {
			if event.Name == EventReady || event.Name == EventResumed {
				established = true
			}
			return s.dispatch(ctx, event)
		}, connectionHooks{
			userAgent:      s.options.UserAgent,
			browser:        s.options.Browser,
			shardID:        s.shardID,
			shardCount:     s.shardCount,
			attached:       s.attach,
			beforeIdentify: s.beforeIdentify,
			resume:         &s.resume,
		})
		s.ready.Store(false)
		var closed *CloseError
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

func (s *Session) attach(session *gatewaySession) {
	s.mu.Lock()
	s.current = session
	s.mu.Unlock()
}

func (s *Session) Send(opcode int, data any) error {
	s.mu.Lock()
	session := s.current
	s.mu.Unlock()
	if session == nil {
		return ErrGatewayClosed
	}
	return session.send(opcode, data)
}

func (s *Session) dispatch(ctx context.Context, event Event) error {
	move := s.state.apply(event.Name, event.Data)
	switch event.Name {
	case EventReady:
		s.mu.Lock()
		s.grants = map[string]VoiceServerUpdate{}
		s.mu.Unlock()
		s.ready.Store(true)
	case EventResumed:
		s.ready.Store(true)
	case EventVoiceStateUpdate:
		s.handleBotVoiceMove(move)
	case EventVoiceServerUpdate:
		s.handleVoiceServerUpdate(event)
	}
	if s.options.Handler == nil {
		return nil
	}
	return s.options.Handler(ctx, event)
}

func (s *Session) ResolveGuild(ctx context.Context, channelID string) (string, error) {
	if guild, ok := s.state.GuildForChannel(channelID); ok {
		return guild, nil
	}
	channel, err := s.client.Channel(ctx, channelID)
	if err != nil {
		return "", err
	}
	s.state.rememberChannel(channelID, channel.GuildID)
	return channel.GuildID, nil
}
