package model

type Member struct {
	User                       *User    `json:"user,omitempty"`
	Nick                       *string  `json:"nick,omitempty"`
	Roles                      []string `json:"roles"`
	JoinedAt                   string   `json:"joined_at,omitempty"`
	Mute                       bool     `json:"mute,omitempty"`
	Deaf                       bool     `json:"deaf,omitempty"`
	CommunicationDisabledUntil *string  `json:"communication_disabled_until,omitempty"`
}
