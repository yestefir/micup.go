package model

const (
	StatusOnline    = "online"
	StatusIdle      = "idle"
	StatusDND       = "dnd"
	StatusInvisible = "invisible"
	StatusOffline   = "offline"
)

type Presence struct {
	Status       string        `json:"status"`
	AFK          bool          `json:"afk"`
	Mobile       bool          `json:"mobile"`
	CustomStatus *CustomStatus `json:"custom_status,omitempty"`
}

type CustomStatus struct {
	Text      string `json:"text,omitempty"`
	ExpiresAt string `json:"expires_at,omitempty"`
	EmojiID   string `json:"emoji_id,omitempty"`
	EmojiName string `json:"emoji_name,omitempty"`
}

func ValidStatus(status string) bool {
	switch status {
	case StatusOnline, StatusIdle, StatusDND, StatusInvisible, StatusOffline:
		return true
	}
	return false
}
