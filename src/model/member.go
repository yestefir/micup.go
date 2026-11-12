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
	ProfileFlags               int          `json:"profile_flags,omitempty"`
	MentionFlags               int          `json:"mention_flags,omitempty"`
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
	TimeoutReason              string
}

type CurrentMemberEdit struct {
	Nick         Optional[string]
	Avatar       Optional[string]
	Banner       Optional[string]
	Bio          Optional[string]
	Pronouns     Optional[string]
	AccentColor  Optional[int]
	ProfileFlags Optional[int]
	MentionFlags Optional[int]
}

const (
	MemberProfileAvatarUnset = 1 << 0
	MemberProfileBannerUnset = 1 << 1

	MentionNoPreference    = 0
	MentionPreferMention   = 1
	MentionPreferNoMention = 2
)

const (
	JoinSourceCreator       = 0
	JoinSourceInstantInvite = 1
	JoinSourceVanityURL     = 2
	JoinSourceBotInvite     = 3
	JoinSourceAdminForceAdd = 4
	JoinSourceDiscovery     = 6
)

type MemberSearch struct {
	Query             string
	Limit             int
	Offset            int
	RoleIDs           []string
	JoinedAtGTE       int64
	JoinedAtLTE       int64
	JoinSourceTypes   []int
	SourceInviteCodes []string
	IsBot             *bool
	UserCreatedAtGTE  int64
	UserCreatedAtLTE  int64
	SortBy            string
	SortOrder         string
}

type MemberSearchResult struct {
	ID            string                   `json:"id"`
	GuildID       string                   `json:"guild_id"`
	UserID        string                   `json:"user_id"`
	Username      string                   `json:"username"`
	Discriminator string                   `json:"discriminator"`
	GlobalName    *string                  `json:"global_name"`
	Nickname      *string                  `json:"nickname"`
	RoleIDs       []string                 `json:"role_ids"`
	JoinedAt      int64                    `json:"joined_at"`
	Supplemental  MemberSearchSupplemental `json:"supplemental"`
	IsBot         bool                     `json:"is_bot"`
}

type MemberSearchSupplemental struct {
	JoinSourceType   *int    `json:"join_source_type"`
	SourceInviteCode *string `json:"source_invite_code"`
	InviterID        *string `json:"inviter_id"`
}

type MemberSearchPage struct {
	GuildID          string               `json:"guild_id"`
	Members          []MemberSearchResult `json:"members"`
	PageResultCount  int                  `json:"page_result_count"`
	TotalResultCount int                  `json:"total_result_count"`
	Indexing         bool                 `json:"indexing"`
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
