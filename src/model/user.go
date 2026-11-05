package model

type User struct {
	ID         string  `json:"id"`
	Username   string  `json:"username"`
	GlobalName *string `json:"global_name,omitempty"`
	Avatar     *string `json:"avatar,omitempty"`
	Bot        bool    `json:"bot"`
}

type Guild struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	OwnerID string `json:"owner_id,omitempty"`
}
