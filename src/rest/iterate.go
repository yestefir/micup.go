package rest

import (
	"context"
	"iter"
	"slices"

	"github.com/yestefir/micup.go/src/model"
)

func (c *Client) EachMessage(ctx context.Context, channelID string, initial model.MessageQuery) iter.Seq2[model.Message, error] {
	return func(yield func(model.Message, error) bool) {
		query := initial
		ascending := query.After != "" && query.Before == ""
		if query.Around != "" {
			yield(model.Message{}, errAroundIteration)
			return
		}
		if query.Limit == 0 {
			query.Limit = model.MaxMessagesPerPage
		}
		for {
			page, err := c.Messages(ctx, channelID, query)
			if err != nil {
				yield(model.Message{}, err)
				return
			}
			slices.SortFunc(page, func(a, b model.Message) int {
				if ascending {
					return model.CompareSnowflakes(a.ID, b.ID)
				}
				return model.CompareSnowflakes(b.ID, a.ID)
			})
			for _, message := range page {
				if !yield(message, nil) {
					return
				}
			}
			if len(page) < query.Limit {
				return
			}
			last := page[len(page)-1].ID
			if last == "" || ascending && last == query.After || !ascending && last == query.Before {
				return
			}
			if ascending {
				query.After = last
			} else {
				query.Before = last
			}
		}
	}
}

func (c *Client) EachMember(ctx context.Context, guildID string) iter.Seq2[model.Member, error] {
	return func(yield func(model.Member, error) bool) {
		query := model.MemberQuery{Limit: maxMembersPerPage}
		for {
			page, err := c.Members(ctx, guildID, query)
			if err != nil {
				yield(model.Member{}, err)
				return
			}
			previous := query.After
			for _, member := range page {
				if !yield(member, nil) {
					return
				}
				if member.User != nil {
					query.After = member.User.ID
				}
			}
			if len(page) < query.Limit || query.After == previous {
				return
			}
		}
	}
}

func (c *Client) EachReactionUser(ctx context.Context, channelID, messageID, emoji string) iter.Seq2[model.User, error] {
	return func(yield func(model.User, error) bool) {
		query := model.ReactionUsersQuery{Limit: model.MaxReactionUsersPerPage}
		for {
			page, err := c.ReactionUsers(ctx, channelID, messageID, emoji, query)
			if err != nil {
				yield(model.User{}, err)
				return
			}
			for _, user := range page.Items {
				if !yield(user, nil) {
					return
				}
			}
			if !page.HasMore || page.NextAfter == nil || *page.NextAfter == "" || *page.NextAfter == query.After {
				return
			}
			query.After = *page.NextAfter
		}
	}
}

func (c *Client) EachAuditLogEntry(ctx context.Context, guildID string, initial model.AuditLogQuery) iter.Seq2[model.AuditLogEntry, error] {
	return func(yield func(model.AuditLogEntry, error) bool) {
		query := initial
		if query.After != "" {
			yield(model.AuditLogEntry{}, errAfterIteration)
			return
		}
		if query.Limit == 0 {
			query.Limit = model.MaxAuditLogEntries
		}
		for {
			page, err := c.AuditLog(ctx, guildID, query)
			if err != nil {
				yield(model.AuditLogEntry{}, err)
				return
			}
			for _, entry := range page.Entries {
				if !yield(entry, nil) {
					return
				}
			}
			if len(page.Entries) < query.Limit {
				return
			}
			last := page.Entries[len(page.Entries)-1].ID
			if last == "" || last == query.Before {
				return
			}
			query.Before = last
		}
	}
}

func (c *Client) EachGuild(ctx context.Context) iter.Seq2[model.Guild, error] {
	return func(yield func(model.Guild, error) bool) {
		query := model.GuildQuery{Limit: maxGuildsPerPage}
		for {
			page, err := c.CurrentUserGuilds(ctx, query)
			if err != nil {
				yield(model.Guild{}, err)
				return
			}
			previous := query.After
			for _, guild := range page {
				if !yield(guild, nil) {
					return
				}
				query.After = guild.ID
			}
			if len(page) < query.Limit || query.After == "" || query.After == previous {
				return
			}
		}
	}
}

func (c *Client) EachPin(ctx context.Context, channelID string) iter.Seq2[model.Pin, error] {
	return func(yield func(model.Pin, error) bool) {
		query := model.PinQuery{Limit: model.MaxPinsPerPage}
		seen := map[string]bool{}
		for {
			page, err := c.PinnedMessages(ctx, channelID, query)
			if err != nil {
				yield(model.Pin{}, err)
				return
			}
			fresh := 0
			for _, pin := range page.Items {
				if seen[pin.Message.ID] {
					continue
				}
				seen[pin.Message.ID] = true
				fresh++
				if !yield(pin, nil) {
					return
				}
			}
			if !page.HasMore || fresh == 0 {
				return
			}
			query.Before = page.Items[len(page.Items)-1].PinnedAt
		}
	}
}

func (c *Client) EachThreadMember(ctx context.Context, threadID string, withMember bool) iter.Seq2[model.ThreadMember, error] {
	return func(yield func(model.ThreadMember, error) bool) {
		query := model.ThreadMemberQuery{Limit: maxThreadMembersPerPage, WithMember: withMember}
		for {
			page, err := c.ThreadMembers(ctx, threadID, query)
			if err != nil {
				yield(model.ThreadMember{}, err)
				return
			}
			previous := query.After
			for _, member := range page {
				if !yield(member, nil) {
					return
				}
				if member.UserID != "" {
					query.After = member.UserID
				}
			}
			if len(page) < query.Limit || query.After == previous {
				return
			}
		}
	}
}
