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
	AvatarColor   *int    `json:"avatar_color,omitempty"`
	MentionFlags  int     `json:"mention_flags,omitempty"`
	Banner        *string `json:"banner,omitempty"`
	BannerColor   *int    `json:"banner_color,omitempty"`
	AccentColor   *int    `json:"accent_color,omitempty"`
	Bio           *string `json:"bio,omitempty"`
	Pronouns      *string `json:"pronouns,omitempty"`
}

const (
	UserFlagStaff                     = 1 << 0
	UserFlagPartner                   = 1 << 2
	UserFlagBugHunter                 = 1 << 3
	UserFlagFriendlyBot               = 1 << 4
	UserFlagFriendlyBotManualApproval = 1 << 5
	UserFlagSpammer                   = 1 << 6
	UserFlagProfileHidden             = 1 << 7
)

type UserProfileQuery struct {
	GuildID           string
	WithMutualFriends bool
	WithMutualGuilds  bool
}

type UserProfile struct {
	User               User            `json:"user"`
	Profile            ProfileDetails  `json:"user_profile"`
	GuildMember        *Member         `json:"guild_member,omitempty"`
	GuildMemberProfile *ProfileDetails `json:"guild_member_profile,omitempty"`
	PremiumType        *int            `json:"premium_type,omitempty"`
	PremiumSince       *string         `json:"premium_since,omitempty"`
	Badges             []ProfileBadge  `json:"profile_badges,omitempty"`
	MutualFriends      []User          `json:"mutual_friends,omitempty"`
	MutualGuilds       []MutualGuild   `json:"mutual_guilds,omitempty"`
	TimezoneOffset     *int            `json:"timezone_offset,omitempty"`
	Limited            bool            `json:"profile_limited,omitempty"`
}

type ProfileDetails struct {
	Bio         *string `json:"bio"`
	Pronouns    *string `json:"pronouns"`
	Banner      *string `json:"banner"`
	BannerColor *int    `json:"banner_color,omitempty"`
	AccentColor *int    `json:"accent_color"`
}

type ProfileBadge struct {
	Key         string   `json:"key"`
	Kind        string   `json:"kind"`
	Label       string   `json:"label"`
	Asset       *string  `json:"asset"`
	AssetURL    *string  `json:"asset_url"`
	AspectRatio *float64 `json:"aspect_ratio"`
	Caption     *string  `json:"caption"`
	Href        *string  `json:"href"`
	Count       *int     `json:"count,omitempty"`
}

type MutualGuild struct {
	ID   string  `json:"id"`
	Nick *string `json:"nick"`
}
