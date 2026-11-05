package model

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Bot      bool   `json:"bot"`
}

type Guild struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
