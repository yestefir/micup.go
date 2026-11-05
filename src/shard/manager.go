package shard

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync/atomic"
	"time"

	"github.com/yestefir/micup.go/src/gateway"
	"github.com/yestefir/micup.go/src/model"
	"github.com/yestefir/micup.go/src/session"
)

const DefaultIdentifyInterval = 5 * time.Second

var (
	ErrNotOwned = errors.New("Micup guild belongs to a shard this manager does not run")
	ErrRunning  = errors.New("Micup shard manager is already running")
)

type Options struct {
	ShardCount       int
	ShardIDs         []int
	IdentifyInterval time.Duration
	Session          session.Options
}

type Manager struct {
	count    int
	ids      []int
	sessions map[int]*session.Session
	channels session.ChannelLookup
	running  atomic.Bool
}

func NewManager(base session.Config, options Options) (*Manager, error) {
	count := options.ShardCount
	if count == 0 {
		count = max(base.ShardCount, 1)
	}
	if count < 1 {
		return nil, errors.New("shard count must be positive")
	}
	ids, err := shardIDs(options.ShardIDs, count)
	if err != nil {
		return nil, err
	}
	interval := options.IdentifyInterval
	if interval <= 0 {
		interval = DefaultIdentifyInterval
	}
	limiter := &identifyLimiter{interval: interval}
	m := &Manager{count: count, ids: ids, sessions: make(map[int]*session.Session, len(ids)), channels: base.Channels}
	for _, id := range ids {
		config := base
		config.ShardID = id
		config.ShardCount = count
		config.BeforeIdentify = limiter.wait
		s, err := session.New(config, options.Session)
		if err != nil {
			return nil, err
		}
		m.sessions[id] = s
	}
	return m, nil
}

func shardIDs(requested []int, count int) ([]int, error) {
	ids := slices.Clone(requested)
	if len(ids) == 0 {
		ids = make([]int, count)
		for i := range ids {
			ids[i] = i
		}
	}
	slices.Sort(ids)
	if len(slices.Compact(slices.Clone(ids))) != len(ids) {
		return nil, errors.New("shard IDs must be unique")
	}
	for _, id := range ids {
		if err := gateway.ValidateShard(id, count); err != nil {
			return nil, err
		}
	}
	return ids, nil
}

func (m *Manager) Run(ctx context.Context) error {
	if !m.running.CompareAndSwap(false, true) {
		return ErrRunning
	}
	defer m.running.Store(false)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	errs := make(chan error, len(m.ids))
	for _, id := range m.ids {
		s := m.sessions[id]
		go func() {
			err := s.Run(ctx)
			if err != nil {
				cancel()
				err = fmt.Errorf("Micup shard %d: %w", s.ShardID(), err)
			}
			errs <- err
		}()
	}
	var first error
	for range m.ids {
		if err := <-errs; err != nil && first == nil {
			first = err
		}
	}
	return first
}

func (m *Manager) ShardCount() int { return m.count }

func (m *Manager) Shards() []*session.Session {
	shards := make([]*session.Session, 0, len(m.ids))
	for _, id := range m.ids {
		shards = append(shards, m.sessions[id])
	}
	return shards
}

func (m *Manager) Shard(shardID int) (*session.Session, bool) {
	s, ok := m.sessions[shardID]
	return s, ok
}

func (m *Manager) ForGuild(guildID string) (*session.Session, error) {
	id, err := ForGuild(guildID, m.count)
	if err != nil {
		return nil, err
	}
	s, ok := m.sessions[id]
	if !ok {
		return nil, ErrNotOwned
	}
	return s, nil
}

func (m *Manager) Ready() bool {
	for _, s := range m.sessions {
		if !s.Ready() {
			return false
		}
	}
	return true
}

func (m *Manager) JoinVoice(ctx context.Context, guildID, channelID string, options session.VoiceJoin) (model.VoiceServerUpdate, error) {
	s, err := m.ForGuild(guildID)
	if err != nil {
		return model.VoiceServerUpdate{}, err
	}
	return s.JoinVoice(ctx, guildID, channelID, options)
}

func (m *Manager) LeaveVoice(guildID, connectionID string) error {
	s, err := m.ForGuild(guildID)
	if err != nil {
		return err
	}
	return s.LeaveVoice(guildID, connectionID)
}

func (m *Manager) Voice(guildID string) (model.VoiceServerUpdate, bool) {
	s, err := m.ForGuild(guildID)
	if err != nil {
		return model.VoiceServerUpdate{}, false
	}
	return s.Voice(guildID)
}

func (m *Manager) ResolveGuild(ctx context.Context, channelID string) (string, error) {
	for _, s := range m.sessions {
		if guild, ok := s.State().GuildForChannel(channelID); ok {
			return guild, nil
		}
	}
	if m.channels == nil {
		return "", errors.New("channel is not cached and no channel lookup is configured")
	}
	channel, err := m.channels.Channel(ctx, channelID)
	if err != nil {
		return "", err
	}
	if s, err := m.ForGuild(channel.GuildID); err == nil {
		s.State().Remember(channelID, channel.GuildID)
	}
	return channel.GuildID, nil
}

func (m *Manager) UpdatePresence(ctx context.Context, presence model.Presence) error {
	var errs []error
	for _, id := range m.ids {
		if err := m.sessions[id].UpdatePresence(ctx, presence); err != nil && !errors.Is(err, gateway.ErrClosed) {
			errs = append(errs, fmt.Errorf("Micup shard %d: %w", id, err))
		}
	}
	return errors.Join(errs...)
}
