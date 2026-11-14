package session

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/yestefir/micup.go/src/events"
	"github.com/yestefir/micup.go/src/gateway"
)

const (
	maxCountGuilds   = 100
	maxCountChannels = 25
	countTimeout     = 12 * time.Second
)

var ErrNoResponse = errors.New("Micup Gateway did not answer the request")

func (s *Session) RequestGuildCounts(ctx context.Context, guildIDs []string) ([]events.GuildCount, error) {
	ids := UniqueIDs(guildIDs)
	if len(ids) == 0 || len(ids) > maxCountGuilds {
		return nil, fmt.Errorf("a guild count request takes 1 to %d guild IDs", maxCountGuilds)
	}
	data, err := s.request(ctx, gateway.OpRequestGuildCounts, map[string]any{"guild_ids": ids})
	if err != nil {
		return nil, err
	}
	update, err := events.As[events.GuildCountsUpdate](gateway.Event{Name: gateway.EventGuildCountsUpdate, Data: data})
	return update.Counts, err
}

func (s *Session) RequestChannelMemberCounts(ctx context.Context, guildID string, channelIDs []string) ([]events.ChannelMemberCount, error) {
	ids := UniqueIDs(channelIDs)
	if guildID == "" || len(ids) == 0 || len(ids) > maxCountChannels {
		return nil, fmt.Errorf("a channel member count request takes a guild ID and 1 to %d channel IDs", maxCountChannels)
	}
	data, err := s.request(ctx, gateway.OpRequestChannelMemberCounts, map[string]any{"guild_id": guildID, "channel_ids": ids})
	if err != nil {
		return nil, err
	}
	update, err := events.As[events.ChannelMemberCountsUpdate](gateway.Event{Name: gateway.EventChannelMemberCountsUpdate, Data: data})
	return update.Counts, err
}

func (s *Session) Subscribe(ctx context.Context, subscriptions map[string]gateway.Subscription) error {
	if err := gateway.ValidateSubscriptions(subscriptions); err != nil {
		return err
	}
	if !s.Ready() {
		return ErrNotReady
	}
	return s.Send(ctx, gateway.OpLazyRequest, map[string]any{"subscriptions": subscriptions})
}

func (s *Session) request(ctx context.Context, opcode int, payload map[string]any) ([]byte, error) {
	if !s.Ready() {
		return nil, ErrNotReady
	}
	select {
	case s.countSlots <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	defer func() { <-s.countSlots }()
	nonce, err := newNonce()
	if err != nil {
		return nil, err
	}
	answer := make(chan []byte, 1)
	s.mu.Lock()
	s.countRequests[nonce] = answer
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.countRequests, nonce)
		s.mu.Unlock()
	}()
	payload["nonce"] = nonce
	if err = s.Send(ctx, opcode, payload); err != nil {
		return nil, err
	}
	timer := time.NewTimer(countTimeout)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-timer.C:
		return nil, ErrNoResponse
	case data := <-answer:
		return data, nil
	}
}

func (s *Session) handleCounts(event gateway.Event) {
	var envelope struct {
		Nonce string `json:"nonce"`
	}
	if event.Decode(&envelope) != nil || envelope.Nonce == "" {
		return
	}
	s.mu.Lock()
	answer := s.countRequests[envelope.Nonce]
	s.mu.Unlock()
	if answer == nil {
		return
	}
	select {
	case answer <- event.Data:
	default:
	}
}

func UniqueIDs(ids []string) []string {
	unique := slices.DeleteFunc(slices.Clone(ids), func(id string) bool { return id == "" })
	slices.Sort(unique)
	return slices.Compact(unique)
}
