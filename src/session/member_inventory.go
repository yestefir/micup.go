package session

import (
	"context"
	"encoding/json"
	"errors"
	"maps"
	"sort"
	"sync"
	"time"

	"github.com/yestefir/micup.go/src/events"
	"github.com/yestefir/micup.go/src/gateway"
	"github.com/yestefir/micup.go/src/model"
)

type inventoryGuild struct {
	members                      map[string]bool
	delta                        map[string]*bool
	available, complete, syncing bool
	retry                        time.Time
}

type memberInventory struct {
	mu     sync.Mutex
	guilds map[string]*inventoryGuild
}

func newInventoryGuild() *inventoryGuild {
	return &inventoryGuild{members: map[string]bool{}, delta: map[string]*bool{}, available: true}
}

func (s *memberInventory) member(g *inventoryGuild, id string, bot *bool) {
	if id == "" {
		return
	}
	if bot == nil {
		delete(g.members, id)
	} else {
		g.members[id] = *bot
	}
	if g.syncing {
		g.delta[id] = bot
	}
}

func (s *memberInventory) observe(name string, data json.RawMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.guilds == nil {
		s.guilds = map[string]*inventoryGuild{}
	}
	switch name {
	case gateway.EventReady:
		s.guilds = map[string]*inventoryGuild{}
		ready, _ := events.Parse[events.Ready](data)
		for _, g := range ready.Guilds {
			s.guild(g)
		}
	case gateway.EventGuildCreate, gateway.EventGuildSync:
		if g, ok := events.Parse[events.GuildCreate](data); ok {
			s.guild(g)
		}
	case gateway.EventGuildDelete:
		e, ok := events.Parse[events.GuildDelete](data)
		if !ok {
			return
		}
		if !e.Unavailable {
			delete(s.guilds, e.ID)
		} else if g := s.guilds[e.ID]; g != nil {
			g.available = false
		}
	case gateway.EventGuildMemberAdd, gateway.EventGuildMemberUpdate:
		if e, ok := events.Parse[events.GuildMember](data); ok && e.User != nil {
			s.observeMember(e.GuildID, e.User)
		}
	case gateway.EventGuildMemberRemove:
		if e, ok := events.Parse[events.GuildMemberRemove](data); ok {
			if g := s.guilds[e.GuildID]; g != nil {
				s.member(g, e.User.ID, nil)
			}
		}
	case gateway.EventVoiceStateUpdate:
		if e, ok := events.Parse[model.VoiceState](data); ok && e.Member != nil && e.Member.User != nil {
			s.observeMember(e.GuildID, e.Member.User)
		}
	}
}

func (s *memberInventory) observeMember(guildID string, user *model.User) {
	if g := s.guilds[guildID]; g != nil {
		bot := user.Bot
		s.member(g, user.ID, &bot)
	}
}

func (s *memberInventory) guild(e events.GuildCreate) {
	if e.ID == "" {
		return
	}
	if e.Unavailable {
		if g := s.guilds[e.ID]; g != nil {
			g.available = false
		} else {
			g = newInventoryGuild()
			g.available = false
			s.guilds[e.ID] = g
		}
		return
	}
	g := newInventoryGuild()
	for _, m := range e.Members {
		if m.User != nil {
			g.members[m.User.ID] = m.User.Bot
		}
	}
	for _, v := range e.VoiceStates {
		if v.Member != nil && v.Member.User != nil {
			g.members[v.Member.User.ID] = v.Member.User.Bot
		}
	}
	s.guilds[e.ID] = g
}

type membersPage func(context.Context, string, string) ([]model.Member, error)

func (s *memberInventory) syncOne(ctx context.Context, page membersPage) bool {
	s.mu.Lock()
	var id string
	var target *inventoryGuild
	for k, g := range s.guilds {
		if g.available && !g.complete && !g.syncing && !time.Now().Before(g.retry) {
			id, target = k, g
			g.syncing = true
			g.delta = map[string]*bool{}
			break
		}
	}
	s.mu.Unlock()
	if target == nil {
		return false
	}
	result := map[string]bool{}
	after := ""
	var err error
	for {
		var members []model.Member
		members, err = page(ctx, id, after)
		if err != nil {
			break
		}
		for _, m := range members {
			if m.User == nil || m.User.ID == "" {
				err = errors.New("member enumeration returned invalid members or a non-advancing cursor")
				break
			}
			result[m.User.ID] = m.User.Bot
		}
		if err != nil {
			break
		}
		if len(members) < 1000 {
			break
		}
		next := ""
		if m := members[len(members)-1]; m.User != nil {
			next = m.User.ID
		}
		if next == "" || next == after {
			err = errors.New("member enumeration returned invalid members or a non-advancing cursor")
			break
		}
		after = next
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.guilds[id] != target {
		return true
	}
	target.syncing = false
	if err != nil {
		target.retry = time.Now().Add(10 * time.Second)
		return true
	}
	for id, bot := range target.delta {
		if bot == nil {
			delete(result, id)
		} else {
			result[id] = *bot
		}
	}
	target.members = result
	target.complete = true
	target.delta = nil
	return true
}

type MemberLookup interface {
	Members(context.Context, string, model.MemberQuery) ([]model.Member, error)
}

type GuildMembers struct {
	GuildID   string
	Members   map[string]bool
	Complete  bool
	Available bool
}

func (s *Session) MemberInventories() []GuildMembers {
	s.inventory.mu.Lock()
	defer s.inventory.mu.Unlock()
	result := make([]GuildMembers, 0, len(s.inventory.guilds))
	for id, g := range s.inventory.guilds {
		result = append(result, GuildMembers{GuildID: id, Members: maps.Clone(g.members), Complete: g.complete, Available: g.available})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].GuildID < result[j].GuildID })
	return result
}

func (s *Session) enumerateMembers(ctx context.Context) {
	page := func(ctx context.Context, guild, after string) ([]model.Member, error) {
		return s.config.Members.Members(ctx, guild, model.MemberQuery{Limit: 1000, After: after})
	}
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		for ctx.Err() == nil && s.Ready() {
			request, cancel := context.WithTimeout(ctx, 30*time.Second)
			worked := s.inventory.syncOne(request, page)
			cancel()
			if !worked {
				break
			}
		}
	}
}
