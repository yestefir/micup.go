package shard

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync/atomic"
	"time"

	"github.com/yestefir/micup.go/src/events"
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

// UpdatePresence replaces the presence on every shard. A nil CustomStatus
// clears the custom status.
func (m *Manager) UpdatePresence(ctx context.Context, presence model.Presence) error {
	if err := presence.Validate(time.Now()); err != nil {
		return err
	}
	return m.eachPresence(func(s *session.Session) error { return s.UpdatePresence(ctx, presence) })
}

// SetStatus changes the status on every shard and keeps the custom status.
func (m *Manager) SetStatus(ctx context.Context, status string) error {
	if !model.ValidStatus(status) {
		return fmt.Errorf("invalid presence status %q", status)
	}
	return m.eachPresence(func(s *session.Session) error { return s.SetStatus(ctx, status) })
}

// SetCustomStatus changes the custom status on every shard and keeps the status.
func (m *Manager) SetCustomStatus(ctx context.Context, custom model.CustomStatus) error {
	if err := custom.Validate(time.Now()); err != nil {
		return err
	}
	return m.eachPresence(func(s *session.Session) error { return s.SetCustomStatus(ctx, custom) })
}

// ClearCustomStatus removes the custom status on every shard.
func (m *Manager) ClearCustomStatus(ctx context.Context) error {
	return m.eachPresence(func(s *session.Session) error { return s.ClearCustomStatus(ctx) })
}

// Presence returns the presence of the first shard; helpers keep all shards equal.
func (m *Manager) Presence() model.Presence {
	return m.sessions[m.ids[0]].Presence()
}

// eachPresence applies a change to every shard. A shard that is disconnected
// still remembers the presence and sends it on its next IDENTIFY.
func (m *Manager) eachPresence(update func(*session.Session) error) error {
	var errs []error
	for _, id := range m.ids {
		if err := update(m.sessions[id]); err != nil && !errors.Is(err, gateway.ErrClosed) {
			errs = append(errs, fmt.Errorf("Micup shard %d: %w", id, err))
		}
	}
	return errors.Join(errs...)
}

func (m *Manager) RequestGuildMembers(ctx context.Context, request model.MembersRequest) ([]model.Member, error) {
	s, err := m.ForGuild(request.GuildID)
	if err != nil {
		return nil, err
	}
	return s.RequestGuildMembers(ctx, request)
}

func (m *Manager) UpdateVoiceState(ctx context.Context, guildID string, options session.VoiceJoin) error {
	s, err := m.ForGuild(guildID)
	if err != nil {
		return err
	}
	return s.UpdateVoiceState(ctx, guildID, options)
}

func (m *Manager) RequestGuildCounts(ctx context.Context, guildIDs []string) ([]events.GuildCount, error) {
	ids := session.UniqueIDs(guildIDs)
	if len(ids) == 0 {
		return nil, errors.New("a guild count request needs at least one guild ID")
	}
	byShard, err := m.groupByShard(ids)
	if err != nil {
		return nil, err
	}
	var counts []events.GuildCount
	for _, id := range m.ids {
		if len(byShard[id]) == 0 {
			continue
		}
		part, err := m.sessions[id].RequestGuildCounts(ctx, byShard[id])
		if err != nil {
			return counts, fmt.Errorf("Micup shard %d: %w", id, err)
		}
		counts = append(counts, part...)
	}
	return counts, nil
}

func (m *Manager) groupByShard(guildIDs []string) (map[int][]string, error) {
	byShard := map[int][]string{}
	for _, guildID := range guildIDs {
		id, err := ForGuild(guildID, m.count)
		if err != nil {
			return nil, err
		}
		if _, ok := m.sessions[id]; !ok {
			return nil, ErrNotOwned
		}
		byShard[id] = append(byShard[id], guildID)
	}
	return byShard, nil
}

func (m *Manager) RequestChannelMemberCounts(ctx context.Context, guildID string, channelIDs []string) ([]events.ChannelMemberCount, error) {
	s, err := m.ForGuild(guildID)
	if err != nil {
		return nil, err
	}
	return s.RequestChannelMemberCounts(ctx, guildID, channelIDs)
}

func (m *Manager) Subscribe(ctx context.Context, subscriptions map[string]gateway.Subscription) error {
	if err := gateway.ValidateSubscriptions(subscriptions); err != nil {
		return err
	}
	guildIDs := make([]string, 0, len(subscriptions))
	for guildID := range subscriptions {
		guildIDs = append(guildIDs, guildID)
	}
	byShard, err := m.groupByShard(guildIDs)
	if err != nil {
		return err
	}
	var errs []error
	for _, id := range m.ids {
		if len(byShard[id]) == 0 {
			continue
		}
		part := make(map[string]gateway.Subscription, len(byShard[id]))
		for _, guildID := range byShard[id] {
			part[guildID] = subscriptions[guildID]
		}
		if err := m.sessions[id].Subscribe(ctx, part); err != nil {
			errs = append(errs, fmt.Errorf("Micup shard %d: %w", id, err))
		}
	}
	return errors.Join(errs...)
}

func (m *Manager) Latency() time.Duration {
	var total time.Duration
	var measured int
	for _, s := range m.sessions {
		if latency := s.Latency(); latency > 0 {
			total += latency
			measured++
		}
	}
	if measured == 0 {
		return 0
	}
	return total / time.Duration(measured)
}

func (m *Manager) Guilds() int {
	total := 0
	for _, s := range m.sessions {
		total += s.State().Guilds()
	}
	return total
}
