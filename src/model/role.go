package model

const (
	MaxRoleName  = 100
	MaxRoleColor = 0xFFFFFF
)

type RoleCreate struct {
	Name             string
	Color            int
	Permissions      *Permissions
	Icon             string
	UnicodeEmoji     string
	IconSourceRoleID string
}

type RoleEdit struct {
	Name          Optional[string]
	Color         Optional[int]
	Permissions   Optional[Permissions]
	Hoist         Optional[bool]
	HoistPosition Optional[int]
	Mentionable   Optional[bool]
	Icon          Optional[string]
	UnicodeEmoji  Optional[string]
}

type RolePosition struct {
	ID       string `json:"id"`
	Position int    `json:"position"`
}

type RoleHoistPosition struct {
	ID            string `json:"id"`
	HoistPosition int    `json:"hoist_position"`
}
