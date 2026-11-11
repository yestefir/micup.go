package model

const (
	ChannelTypeGuildText          = 0
	ChannelTypeDM                 = 1
	ChannelTypeGuildVoice         = 2
	ChannelTypeGroupDM            = 3
	ChannelTypeGuildCategory      = 4
	ChannelTypeGuildAnnouncement  = 5
	ChannelTypeAnnouncementThread = 10
	ChannelTypePublicThread       = 11
	ChannelTypePrivateThread      = 12
	ChannelTypeGuildForum         = 15
	ChannelTypeGuildMedia         = 16
	ChannelTypeGuildLink          = 998
	ChannelTypePersonalNotes      = 999
)

const (
	ChannelFlagPinned                   = 1 << 1
	ChannelFlagRequireTag               = 1 << 4
	ChannelFlagHideMediaDownloadOptions = 1 << 15
)

const (
	AutoArchiveHour      = 60
	AutoArchiveDay       = 1440
	AutoArchiveThreeDays = 4320
	AutoArchiveWeek      = 10080
)

const (
	ForumSortLatestActivity = 0
	ForumSortCreationTime   = 1
	ForumLayoutDefault      = 0
	ForumLayoutList         = 1
	ForumLayoutGrid         = 2
	ForumTagMatchSome       = "match_some"
	ForumTagMatchAll        = "match_all"
)

const (
	ThreadMemberHasInteracted = 1 << 0
	ThreadMemberAllMessages   = 1 << 1
	ThreadMemberOnlyMentions  = 1 << 2
	ThreadMemberNoMessages    = 1 << 3
)

const (
	MaxChannelName      = 100
	MaxChannelTopic     = 1024
	MaxForumTopic       = 4096
	MaxSlowmodeSeconds  = 21600
	MaxForumTags        = 20
	MaxAppliedTags      = 5
	MaxForumTagName     = 50
	MaxChannelPositions = 500
)

type Channel struct {
	ID                            string                `json:"id"`
	GuildID                       string                `json:"guild_id"`
	Name                          string                `json:"name"`
	Type                          int                   `json:"type"`
	Topic                         *string               `json:"topic,omitempty"`
	URL                           *string               `json:"url,omitempty"`
	Icon                          *string               `json:"icon,omitempty"`
	OwnerID                       string                `json:"owner_id,omitempty"`
	Position                      int                   `json:"position,omitempty"`
	ParentID                      string                `json:"parent_id,omitempty"`
	Bitrate                       int                   `json:"bitrate,omitempty"`
	UserLimit                     int                   `json:"user_limit,omitempty"`
	VoiceConnectionLimit          int                   `json:"voice_connection_limit,omitempty"`
	RTCRegion                     *string               `json:"rtc_region,omitempty"`
	RTCP2P                        bool                  `json:"rtc_p2p,omitempty"`
	NSFW                          bool                  `json:"nsfw,omitempty"`
	NSFWOverride                  *bool                 `json:"nsfw_override,omitempty"`
	ContentWarningLevel           int                   `json:"content_warning_level,omitempty"`
	ContentWarningText            *string               `json:"content_warning_text,omitempty"`
	RateLimitPerUser              int                   `json:"rate_limit_per_user,omitempty"`
	LastMessageID                 string                `json:"last_message_id,omitempty"`
	LastPinTimestamp              *string               `json:"last_pin_timestamp,omitempty"`
	PermissionOverwrites          []PermissionOverwrite `json:"permission_overwrites,omitempty"`
	Permissions                   *Permissions          `json:"permissions,omitempty"`
	Recipients                    []User                `json:"recipients,omitempty"`
	Nicks                         map[string]string     `json:"nicks,omitempty"`
	Flags                         int                   `json:"flags,omitempty"`
	ThreadMetadata                *ThreadMetadata       `json:"thread_metadata,omitempty"`
	MessageCount                  int                   `json:"message_count,omitempty"`
	TotalMessageSent              int                   `json:"total_message_sent,omitempty"`
	MemberCount                   int                   `json:"member_count,omitempty"`
	AppliedTags                   []string              `json:"applied_tags,omitempty"`
	MemberIDsPreview              []string              `json:"member_ids_preview,omitempty"`
	Member                        *ThreadMember         `json:"member,omitempty"`
	DefaultAutoArchiveDuration    int                   `json:"default_auto_archive_duration,omitempty"`
	DefaultThreadRateLimitPerUser int                   `json:"default_thread_rate_limit_per_user,omitempty"`
	AvailableTags                 []ForumTag            `json:"available_tags,omitempty"`
	DefaultReactionEmoji          *DefaultReaction      `json:"default_reaction_emoji,omitempty"`
	DefaultSortOrder              *int                  `json:"default_sort_order,omitempty"`
	DefaultForumLayout            int                   `json:"default_forum_layout,omitempty"`
	DefaultTagSetting             string                `json:"default_tag_setting,omitempty"`
}

type ThreadMetadata struct {
	Archived            bool   `json:"archived"`
	AutoArchiveDuration int    `json:"auto_archive_duration"`
	ArchiveTimestamp    string `json:"archive_timestamp,omitempty"`
	Locked              bool   `json:"locked"`
	Invitable           *bool  `json:"invitable,omitempty"`
	CreateTimestamp     string `json:"create_timestamp,omitempty"`
}

type ThreadMember struct {
	ID            string  `json:"id,omitempty"`
	UserID        string  `json:"user_id,omitempty"`
	GuildID       string  `json:"guild_id,omitempty"`
	JoinTimestamp string  `json:"join_timestamp,omitempty"`
	Flags         int     `json:"flags"`
	Muted         bool    `json:"muted,omitempty"`
	Member        *Member `json:"member,omitempty"`
}

type ForumTag struct {
	ID        string  `json:"id,omitempty"`
	Name      string  `json:"name"`
	Moderated bool    `json:"moderated"`
	EmojiID   *string `json:"emoji_id"`
	EmojiName *string `json:"emoji_name"`
}

type DefaultReaction struct {
	EmojiID   *string `json:"emoji_id"`
	EmojiName *string `json:"emoji_name"`
}

type ChannelEdit struct {
	Name                          Optional[string]
	Type                          Optional[int]
	Topic                         Optional[string]
	URL                           Optional[string]
	ParentID                      Optional[string]
	Bitrate                       Optional[int]
	UserLimit                     Optional[int]
	VoiceConnectionLimit          Optional[int]
	PermissionOverwrites          Optional[[]PermissionOverwrite]
	NSFW                          Optional[bool]
	ContentWarningLevel           Optional[int]
	ContentWarningText            Optional[string]
	RateLimitPerUser              Optional[int]
	RTCRegion                     Optional[string]
	RTCP2P                        Optional[bool]
	DefaultAutoArchiveDuration    Optional[int]
	DefaultThreadRateLimitPerUser Optional[int]
	AvailableTags                 Optional[[]ForumTag]
	DefaultReactionEmoji          Optional[DefaultReaction]
	DefaultSortOrder              Optional[int]
	DefaultForumLayout            Optional[int]
	DefaultTagSetting             Optional[string]
	Flags                         Optional[int]
}

type ChannelCreate struct {
	Type    int
	Name    string
	Options ChannelEdit
}

type ChannelPosition struct {
	ID                 string
	Position           Optional[int]
	ParentID           Optional[string]
	PrecedingSiblingID Optional[string]
	LockPermissions    bool
}

type FollowedChannel struct {
	ChannelID string `json:"channel_id"`
	WebhookID string `json:"webhook_id"`
}

type ChannelFollowerStats struct {
	ChannelCount int `json:"channel_count"`
	GuildCount   int `json:"guild_count"`
}

type ThreadStart struct {
	Name                string
	Type                int
	AutoArchiveDuration int
	RateLimitPerUser    Optional[int]
	Invitable           Optional[bool]
}

type ForumPostStart struct {
	Name                string
	Message             MessageOptions
	AppliedTags         []string
	AutoArchiveDuration int
	RateLimitPerUser    Optional[int]
}

type ForumPost struct {
	Channel
	Message *Message `json:"message,omitempty"`
}

type ThreadEdit struct {
	Name                Optional[string]
	Archived            Optional[bool]
	AutoArchiveDuration Optional[int]
	Locked              Optional[bool]
	RateLimitPerUser    Optional[int]
	Flags               Optional[int]
	AppliedTags         Optional[[]string]
	Invitable           Optional[bool]
}

type ThreadList struct {
	Threads       []Channel      `json:"threads"`
	Members       []ThreadMember `json:"members,omitempty"`
	HasMore       bool           `json:"has_more,omitempty"`
	TotalResults  int            `json:"total_results,omitempty"`
	FirstMessages []Message      `json:"first_messages,omitempty"`
}

type ArchivedThreadQuery struct {
	Before string
	Limit  int
}

type ThreadSearch struct {
	Name       string
	Tags       []string
	TagSetting string
	Archived   *bool
	SortBy     string
	SortOrder  string
	Limit      int
	Offset     int
	MaxID      string
	MinID      string
}

type ThreadMemberQuery struct {
	WithMember bool
	After      string
	Limit      int
}

type ForumPostData struct {
	Owner        *Member  `json:"owner"`
	FirstMessage *Message `json:"first_message"`
}
