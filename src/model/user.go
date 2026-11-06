package model

type User struct {
	ID            string  `json:"id"`
	Username      string  `json:"username"`
	Discriminator string  `json:"discriminator,omitempty"`
	GlobalName    *string `json:"global_name,omitempty"`
	Avatar        *string `json:"avatar,omitempty"`
	Bot           bool    `json:"bot"`
	System        bool    `json:"system,omitempty"`
	Flags         int     `json:"flags,omitempty"`
}
