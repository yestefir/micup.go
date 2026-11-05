package model

import (
	"errors"
	"fmt"
	"time"
	"unicode/utf16"
)

const (
	StatusOnline    = "online"
	StatusIdle      = "idle"
	StatusDND       = "dnd"
	StatusInvisible = "invisible"
	// StatusOffline is published as invisible by Micup.
	StatusOffline = "offline"
)

// Micup limits from its custom status schema. Lengths are counted in UTF-16
// units, as Micup counts them.
const (
	MaxCustomStatusText  = 128
	MaxCustomStatusEmoji = 32
)

// Presence is the bot's own status. An update replaces the whole presence: a
// nil CustomStatus is sent as null, which clears the custom status on Micup.
type Presence struct {
	Status       string        `json:"status"`
	AFK          bool          `json:"afk"`
	Mobile       bool          `json:"mobile"`
	CustomStatus *CustomStatus `json:"custom_status"`
}

type CustomStatus struct {
	Text string `json:"text,omitempty"`
	// ExpiresAt is an RFC 3339 time; Micup clears the status then. Until sets it.
	ExpiresAt string `json:"expires_at,omitempty"`
	// EmojiID is a custom emoji; Micup ignores EmojiName when it is set.
	EmojiID string `json:"emoji_id,omitempty"`
	// EmojiName is a single Unicode emoji.
	EmojiName string `json:"emoji_name,omitempty"`
}

// Until returns a copy of the custom status that Micup clears at t.
func (c CustomStatus) Until(t time.Time) CustomStatus {
	c.ExpiresAt = t.UTC().Format(time.RFC3339)
	return c
}

// Validate reports values Micup would refuse. Micup drops a refused Gateway
// presence update without telling the bot, so the session checks first.
func (c CustomStatus) Validate(now time.Time) error {
	if c.Text == "" && c.EmojiID == "" && c.EmojiName == "" {
		return errors.New("custom status needs text or an emoji; clear it instead")
	}
	if n := utf16Len(c.Text); n > MaxCustomStatusText {
		return fmt.Errorf("custom status text is %d characters, Micup allows %d", n, MaxCustomStatusText)
	}
	if n := utf16Len(c.EmojiName); n > MaxCustomStatusEmoji {
		return fmt.Errorf("custom status emoji name is %d characters, Micup allows %d", n, MaxCustomStatusEmoji)
	}
	if c.ExpiresAt != "" {
		at, err := time.Parse(time.RFC3339, c.ExpiresAt)
		if err != nil {
			return fmt.Errorf("custom status expires_at %q is not an RFC 3339 time", c.ExpiresAt)
		}
		if !at.After(now) {
			return errors.New("custom status expires_at must be in the future")
		}
	}
	return nil
}

// Validate checks the status and the custom status.
func (p Presence) Validate(now time.Time) error {
	if !ValidStatus(p.Status) {
		return fmt.Errorf("invalid presence status %q", p.Status)
	}
	if p.CustomStatus != nil {
		return p.CustomStatus.Validate(now)
	}
	return nil
}

func ValidStatus(status string) bool {
	switch status {
	case StatusOnline, StatusIdle, StatusDND, StatusInvisible, StatusOffline:
		return true
	}
	return false
}

func utf16Len(s string) int { return len(utf16.Encode([]rune(s))) }
