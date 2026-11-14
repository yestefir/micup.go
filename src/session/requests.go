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
	event, err := s.request(ctx, gateway.OpRequestGuildCounts, map[string]any{"guild_ids": ids})
	if err != nil {
		return nil, err
	}
	update, err := events.As[events.GuildCountsUpdate](event)
	if err != nil {
		return update.Counts, err
	}
	for _, count := range update.Counts {
		if !slices.Contains(ids, count.GuildID) {
			return nil, errors.New("Micup guild count response contains an unrequested guild")
		}
	}
	return update.Counts, nil
}

func (s *Session) RequestChannelMemberCounts(ctx context.Context, guildID string, channelIDs []string) ([]events.ChannelMemberCount, error) {
	ids := UniqueIDs(channelIDs)
	if guildID == "" || len(ids) == 0 || len(ids) > maxCountChannels {
		return nil, fmt.Errorf("a channel member count request takes a guild ID and 1 to %d channel IDs", maxCountChannels)
	}
	event, err := s.request(ctx, gateway.OpRequestChannelMemberCounts, map[string]any{"guild_id": guildID, "channel_ids": ids})
	if err != nil {
		return nil, err
	}
	update, err := events.As[events.ChannelMemberCountsUpdate](event)
	if err != nil {
		return update.Counts, err
	}
	for _, count := range update.Counts {
		if count.GuildID != guildID || !slices.Contains(ids, count.ChannelID) {
			return nil, errors.New("Micup channel count response contains an unrequested guild or channel")
		}
	}
	return update.Counts, nil
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

func (s *Session) request(ctx context.Context, opcode int, payload map[string]any) (gateway.Event, error) {
	if !s.Ready() {
		return gateway.Event{}, ErrNotReady
	}
	select {
	case s.countSlots <- struct{}{}:
	case <-ctx.Done():
		return gateway.Event{}, ctx.Err()
	}
	defer func() { <-s.countSlots }()
	nonce, err := newNonce()
	if err != nil {
		return gateway.Event{}, err
	}
	answer := make(chan gateway.Event, 1)
	s.mu.Lock()
	if !s.Ready() {
		s.mu.Unlock()
		return gateway.Event{}, ErrNotReady
	}
	s.countRequests[nonce] = answer
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.countRequests, nonce)
		s.mu.Unlock()
	}()
	payload["nonce"] = nonce
	if err = s.Send(ctx, opcode, payload); err != nil {
		return gateway.Event{}, err
	}
	timer := time.NewTimer(countTimeout)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return gateway.Event{}, ctx.Err()
	case <-timer.C:
		return gateway.Event{}, ErrNoResponse
	case data, open := <-answer:
		if !open {
			return gateway.Event{}, gateway.ErrClosed
		}
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
	defer s.mu.Unlock()
	answer := s.countRequests[envelope.Nonce]
	if answer == nil {
		return
	}
	select {
	case answer <- gateway.Event{Name: event.Name, Data: slices.Clone(event.Data)}:
	default:
	}
}

func (s *Session) stopCountRequests() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for nonce, answer := range s.countRequests {
		delete(s.countRequests, nonce)
		// Each answer has one buffer slot. Discard an unconsumed response
		// so pending callers observe shutdown rather than stale success.
		select {
		case <-answer:
		default:
		}
		close(answer)
	}
}

func UniqueIDs(ids []string) []string {
	unique := slices.DeleteFunc(slices.Clone(ids), func(id string) bool { return id == "" })
	slices.Sort(unique)
	return slices.Compact(unique)
}
