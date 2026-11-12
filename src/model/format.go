package model

import (
	"strconv"
	"strings"
	"time"
)

const (
	TimestampShortTime     = "t"
	TimestampLongTime      = "T"
	TimestampShortDate     = "d"
	TimestampLongDate      = "D"
	TimestampShortDateTime = "f"
	TimestampLongDateTime  = "F"
	TimestampRelative      = "R"
)

func MentionUser(userID string) string { return "<@" + userID + ">" }

func MentionRole(roleID string) string { return "<@&" + roleID + ">" }

func MentionChannel(channelID string) string { return "<#" + channelID + ">" }

func FormatTimestamp(at time.Time, style string) string {
	text := "<t:" + strconv.FormatInt(at.Unix(), 10)
	if style != "" {
		text += ":" + style
	}
	return text + ">"
}

func (e Emoji) String() string {
	switch {
	case e.ID == "":
		return e.Name
	case e.Animated:
		return "<a:" + e.Name + ":" + e.ID + ">"
	}
	return "<:" + e.Name + ":" + e.ID + ">"
}

func (e Emoji) Reaction() string {
	if e.ID != "" {
		return e.Name + ":" + e.ID
	}
	return e.Name
}

var markdownEscaper = strings.NewReplacer(
	`\`, `\\`, `*`, `\*`, `_`, `\_`, `~`, "\\~", "`", "\\`", `|`, `\|`, `>`, `\>`, `<`, `\<`, `[`, `\[`, `]`, `\]`, `#`, `\#`,
)

func EscapeMarkdown(text string) string { return markdownEscaper.Replace(text) }

func (u User) Mention() string { return MentionUser(u.ID) }

func (u User) DisplayName() string {
	if u.GlobalName != nil && *u.GlobalName != "" {
		return *u.GlobalName
	}
	return u.Username
}

func (u User) Tag() string {
	if u.Discriminator == "" || u.Discriminator == "0" || u.Discriminator == "0000" {
		return u.Username
	}
	return u.Username + "#" + u.Discriminator
}

func (m Member) DisplayName() string {
	if m.Nick != nil && *m.Nick != "" {
		return *m.Nick
	}
	if m.User != nil {
		return m.User.DisplayName()
	}
	return ""
}

func (m Member) TimedOut(now time.Time) bool {
	if m.CommunicationDisabledUntil == nil {
		return false
	}
	until, err := time.Parse(time.RFC3339, *m.CommunicationDisabledUntil)
	return err == nil && until.After(now)
}

func (r Role) Mention() string { return MentionRole(r.ID) }

func (c Channel) Mention() string { return MentionChannel(c.ID) }

func (c Channel) IsThread() bool {
	return c.Type == ChannelTypeAnnouncementThread || c.Type == ChannelTypePublicThread || c.Type == ChannelTypePrivateThread
}

func (c Channel) IsVoice() bool { return c.Type == ChannelTypeGuildVoice }

func (c Channel) IsDM() bool { return c.Type == ChannelTypeDM || c.Type == ChannelTypeGroupDM }

func (c Channel) IsForum() bool {
	return c.Type == ChannelTypeGuildForum || c.Type == ChannelTypeGuildMedia
}

func (m Message) CreatedAt() (time.Time, error) { return SnowflakeTime(m.ID) }
