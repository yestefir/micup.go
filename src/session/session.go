package session

import (
	"context"
	"errors"
	"slices"
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
	OnReady       func(resumed bool)
	OnDisconnect  func(err error)
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
	pending        map[string]voicePending
	joinOrder      []string
	presence       *model.Presence
	presenceMu     sync.Mutex
	memberRequests map[string]chan memberChunk
	memberSlots    chan struct{}
	countRequests  map[string]chan []byte
	countSlots     chan struct{}
	sessionID      atomic.Value
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
	if options.Presence != nil {
		if err := options.Presence.Validate(time.Now()); err != nil {
			return nil, err
		}
		presence := copyPresence(*options.Presence)
		options.Presence = &presence
	}
	if options.MinBackoff <= 0 {
		options.MinBackoff = DefaultMinBackoff
	}
	if options.MaxBackoff < options.MinBackoff {
		options.MaxBackoff = max(DefaultMaxBackoff, options.MinBackoff)
	}
	options.IgnoredEvents = slices.Clone(options.IgnoredEvents)
	return &Session{
		config:         config,
		options:        options,
		cache:          state.New(config.BotID),
		grants:         map[string]model.VoiceServerUpdate{},
		pending:        map[string]voicePending{},
		memberRequests: map[string]chan memberChunk{},
		memberSlots:    make(chan struct{}, maxMemberRequestsInFlight),
		countRequests:  map[string]chan []byte{},
		countSlots:     make(chan struct{}, 1),
		presence:       options.Presence,
	}, nil
}

func (s *Session) State() *state.Cache { return s.cache }

func (s *Session) Ready() bool { return s.ready.Load() }

func (s *Session) ShardID() int { return s.config.ShardID }

func (s *Session) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	var requests sync.WaitGroup
	defer func() {
		cancel()
		s.stopMemberRequests()
		s.stopCountRequests()
		s.stopVoiceJoins()
		requests.Wait()
	}()
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
		if s.options.OnDisconnect != nil && ctx.Err() == nil {
			s.options.OnDisconnect(err)
		}
		var closed *gateway.CloseError
		if errors.As(err, &closed) {
			if closed.ResetsSession() {
				s.invalidateSession()
			}
			if closed.Fatal() {
				return closed
			}
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
		URL:                  s.config.Endpoint,
		Token:                s.config.Token,
		ShardID:              s.config.ShardID,
		ShardCount:           s.config.ShardCount,
		UserAgent:            s.options.UserAgent,
		Browser:              s.options.Browser,
		Resume:               &s.resume,
		OnConnect:            s.attach,
		OnDisconnect:         func() { s.attach(nil) },
		OnSessionInvalidated: s.invalidateSession,
		OnGatewayError:       s.handleGatewayError,
		BeforeIdentify:       s.config.BeforeIdentify,
		IgnoredEvents:        s.options.IgnoredEvents,
		PresenceFunc:         s.currentPresence,
		Flags:                s.options.Flags,
		Compress:             s.options.Compress,
	}
}

func (s *Session) invalidateSession() {
	s.ready.Store(false)
	s.stopMemberRequests()
	s.stopCountRequests()
	s.dropAllGrants()
}

// Latency is the Gateway heartbeat round trip of the live connection. It is
// 0 while disconnected and until the first heartbeat is acknowledged, one
// heartbeat interval after connecting.
func (s *Session) Latency() time.Duration {
	s.mu.Lock()
	conn := s.conn
	s.mu.Unlock()
	if conn == nil {
		return 0
	}
	return conn.Latency()
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
		s.invalidateSession()
		s.rememberSession(event)
		s.ready.Store(true)
		s.announceReady(false)
	case gateway.EventResumed:
		s.ready.Store(true)
		s.announceReady(true)
	case gateway.EventGuildCountsUpdate, gateway.EventChannelMemberCountsUpdate:
		s.handleCounts(event)
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

func (s *Session) SessionID() string {
	id, _ := s.sessionID.Load().(string)
	return id
}

func (s *Session) rememberSession(event gateway.Event) {
	var ready struct {
		SessionID string `json:"session_id"`
	}
	if event.Decode(&ready) == nil {
		s.sessionID.Store(ready.SessionID)
	}
}

func (s *Session) announceReady(resumed bool) {
	if s.options.OnReady != nil {
		go s.options.OnReady(resumed)
	}
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

// UpdatePresence replaces the bot's presence on this session and remembers it
// for later IDENTIFYs. A nil CustomStatus clears the custom status.
func (s *Session) UpdatePresence(ctx context.Context, presence model.Presence) error {
	return s.changePresence(ctx, func(*model.Presence) model.Presence { return presence })
}

// SetStatus changes only the status and keeps the custom status.
func (s *Session) SetStatus(ctx context.Context, status string) error {
	return s.changePresence(ctx, func(p *model.Presence) model.Presence {
		p.Status = status
		return *p
	})
}

// SetCustomStatus changes only the custom status and keeps the status.
func (s *Session) SetCustomStatus(ctx context.Context, custom model.CustomStatus) error {
	return s.changePresence(ctx, func(p *model.Presence) model.Presence {
		p.CustomStatus = &custom
		return *p
	})
}

// ClearCustomStatus removes the custom status and keeps the status.
func (s *Session) ClearCustomStatus(ctx context.Context) error {
	return s.changePresence(ctx, func(p *model.Presence) model.Presence {
		p.CustomStatus = nil
		return *p
	})
}

// Presence returns the presence this session publishes. Before any update
// it is online with no custom status.
func (s *Session) Presence() model.Presence {
	if p := s.currentPresence(); p != nil {
		return *p
	}
	return model.Presence{Status: model.StatusOnline}
}

// changePresence serialises read-modify-send so concurrent helpers do not
// overwrite each other's field.
func (s *Session) changePresence(ctx context.Context, change func(*model.Presence) model.Presence) error {
	s.presenceMu.Lock()
	defer s.presenceMu.Unlock()
	current := s.Presence()
	presence := copyPresence(change(&current))
	if err := presence.Validate(time.Now()); err != nil {
		return err
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
	presence := copyPresence(*s.presence)
	return &presence
}

func copyPresence(presence model.Presence) model.Presence {
	if presence.CustomStatus != nil {
		custom := *presence.CustomStatus
		presence.CustomStatus = &custom
	}
	return presence
}
