package model

type Member struct {
	User                       *User        `json:"user,omitempty"`
	Nick                       *string      `json:"nick,omitempty"`
	Avatar                     *string      `json:"avatar,omitempty"`
	Banner                     *string      `json:"banner,omitempty"`
	AccentColor                *int         `json:"accent_color,omitempty"`
	Roles                      []string     `json:"roles"`
	JoinedAt                   string       `json:"joined_at,omitempty"`
	Mute                       bool         `json:"mute,omitempty"`
	Deaf                       bool         `json:"deaf,omitempty"`
	Flags                      int          `json:"flags,omitempty"`
	CommunicationDisabledUntil *string      `json:"communication_disabled_until,omitempty"`
	Permissions                *Permissions `json:"permissions,omitempty"`
}

type MemberEdit struct {
	Nick                       *string
	Roles                      *[]string
	Mute                       *bool
	Deaf                       *bool
	CommunicationDisabledUntil *string
	ChannelID                  *string
	Disconnect                 bool
	ConnectionID               string
}

type MemberQuery struct {
	Limit int
	After string
}

type Ban struct {
	DeleteMessageSeconds int
	DurationSeconds      int
	Reason               string
}

type MembersRequest struct {
	GuildID   string
	Query     string
	Limit     int
	UserIDs   []string
	Presences bool
}
