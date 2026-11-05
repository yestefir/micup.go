package micup

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const DefaultIdentifyInterval = 5 * time.Second

var (
	ErrShardNotOwned       = errors.New("Micup guild belongs to a shard this manager does not run")
	ErrShardManagerRunning = errors.New("Micup shard manager is already running")
)

func ShardForGuild(guildID string, shardCount int) (int, error) {
	if shardCount < 1 {
		return 0, errors.New("shard count must be positive")
	}
	id, err := strconv.ParseUint(guildID, 10, 64)
	if err != nil {
		return 0, errors.New("guild ID must be a numeric snowflake")
	}
	return int((id >> 22) % uint64(shardCount)), nil
}

func validateShard(shardID, shardCount int) error {
	if shardCount < 1 || shardID < 0 || shardID >= shardCount {
		return errors.New("invalid shard configuration")
	}
	return nil
}

type ShardManagerOptions struct {
	ShardCount       int
	ShardIDs         []int
	IdentifyInterval time.Duration
	Session          SessionOptions
}

type ShardManager struct {
	client   *Client
	count    int
	ids      []int
	sessions map[int]*Session
	running  atomic.Bool
}

func (c *Client) NewShardManager(options ShardManagerOptions) (*ShardManager, error) {
	count := options.ShardCount
	if count == 0 {
		count = c.config.ShardCount
	}
	if count < 1 {
		return nil, errors.New("shard count must be positive")
	}
	ids := slices.Clone(options.ShardIDs)
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
	interval := options.IdentifyInterval
	if interval <= 0 {
		interval = DefaultIdentifyInterval
	}
	limiter := &identifyLimiter{interval: interval}
	m := &ShardManager{client: c, count: count, ids: ids, sessions: make(map[int]*Session, len(ids))}
	for _, id := range ids {
		if err := validateShard(id, count); err != nil {
			return nil, err
		}
		session := c.newSession(options.Session, id, count)
		session.beforeIdentify = limiter.wait
		m.sessions[id] = session
	}
	return m, nil
}

func (m *ShardManager) Run(ctx context.Context) error {
	if !m.running.CompareAndSwap(false, true) {
		return ErrShardManagerRunning
	}
	defer m.running.Store(false)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	errs := make(chan error, len(m.ids))
	for _, id := range m.ids {
		session := m.sessions[id]
		go func() {
			err := session.Run(ctx)
			if err != nil {
				cancel()
				err = fmt.Errorf("Micup shard %d: %w", session.shardID, err)
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

func (m *ShardManager) ShardCount() int { return m.count }

func (m *ShardManager) Shards() []*Session {
	shards := make([]*Session, 0, len(m.ids))
	for _, id := range m.ids {
		shards = append(shards, m.sessions[id])
	}
	return shards
}

func (m *ShardManager) Shard(shardID int) (*Session, bool) {
	session, ok := m.sessions[shardID]
	return session, ok
}

func (m *ShardManager) ForGuild(guildID string) (*Session, error) {
	id, err := ShardForGuild(guildID, m.count)
	if err != nil {
		return nil, err
	}
	session, ok := m.sessions[id]
	if !ok {
		return nil, ErrShardNotOwned
	}
	return session, nil
}

func (m *ShardManager) Ready() bool {
	for _, session := range m.sessions {
		if !session.Ready() {
			return false
		}
	}
	return true
}

func (m *ShardManager) JoinVoice(ctx context.Context, guildID, channelID string, options VoiceJoin) (VoiceServerUpdate, error) {
	session, err := m.ForGuild(guildID)
	if err != nil {
		return VoiceServerUpdate{}, err
	}
	return session.JoinVoice(ctx, guildID, channelID, options)
}

func (m *ShardManager) LeaveVoice(guildID, connectionID string) error {
	session, err := m.ForGuild(guildID)
	if err != nil {
		return err
	}
	return session.LeaveVoice(guildID, connectionID)
}

func (m *ShardManager) Voice(guildID string) (VoiceServerUpdate, bool) {
	session, err := m.ForGuild(guildID)
	if err != nil {
		return VoiceServerUpdate{}, false
	}
	return session.Voice(guildID)
}

func (m *ShardManager) ResolveGuild(ctx context.Context, channelID string) (string, error) {
	for _, session := range m.sessions {
		if guild, ok := session.state.GuildForChannel(channelID); ok {
			return guild, nil
		}
	}
	channel, err := m.client.Channel(ctx, channelID)
	if err != nil {
		return "", err
	}
	if session, err := m.ForGuild(channel.GuildID); err == nil {
		session.state.rememberChannel(channelID, channel.GuildID)
	}
	return channel.GuildID, nil
}

type identifyLimiter struct {
	mu       sync.Mutex
	interval time.Duration
	next     time.Time
}

func (l *identifyLimiter) wait(ctx context.Context) error {
	l.mu.Lock()
	now := time.Now()
	at := l.next
	if at.Before(now) {
		at = now
	}
	l.next = at.Add(l.interval)
	l.mu.Unlock()
	return sleep(ctx, time.Until(at))
}
