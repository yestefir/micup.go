package model

type Guild struct {
	ID                          string       `json:"id"`
	Name                        string       `json:"name"`
	OwnerID                     string       `json:"owner_id,omitempty"`
	Icon                        *string      `json:"icon,omitempty"`
	Banner                      *string      `json:"banner,omitempty"`
	Splash                      *string      `json:"splash,omitempty"`
	Features                    []string     `json:"features,omitempty"`
	AFKChannelID                *string      `json:"afk_channel_id,omitempty"`
	AFKTimeout                  int          `json:"afk_timeout,omitempty"`
	SystemChannelID             *string      `json:"system_channel_id,omitempty"`
	SystemChannelFlags          int          `json:"system_channel_flags,omitempty"`
	RulesChannelID              *string      `json:"rules_channel_id,omitempty"`
	VanityURLCode               *string      `json:"vanity_url_code,omitempty"`
	VerificationLevel           int          `json:"verification_level,omitempty"`
	MFALevel                    int          `json:"mfa_level,omitempty"`
	NSFW                        bool         `json:"nsfw,omitempty"`
	NSFWLevel                   int          `json:"nsfw_level,omitempty"`
	ExplicitContentFilter       int          `json:"explicit_content_filter,omitempty"`
	DefaultMessageNotifications int          `json:"default_message_notifications,omitempty"`
	PremiumTier                 int          `json:"premium_tier,omitempty"`
	PremiumSubscriptionCount    int          `json:"premium_subscription_count,omitempty"`
	MemberCount                 int          `json:"member_count,omitempty"`
	OnlineCount                 int          `json:"online_count,omitempty"`
	Roles                       []Role       `json:"roles,omitempty"`
	Emojis                      []Emoji      `json:"emojis,omitempty"`
	Stickers                    []Sticker    `json:"stickers,omitempty"`
	EmbedSplash                 *string      `json:"embed_splash,omitempty"`
	SplashCardAlignment         int          `json:"splash_card_alignment,omitempty"`
	BadgeIDs                    []string     `json:"badge_ids,omitempty"`
	ContentWarningLevel         int          `json:"content_warning_level,omitempty"`
	ContentWarningText          *string      `json:"content_warning_text,omitempty"`
	DisabledOperations          int          `json:"disabled_operations,omitempty"`
	MessageHistoryCutoff        *string      `json:"message_history_cutoff,omitempty"`
	Permissions                 *Permissions `json:"permissions,omitempty"`
	ApproximateMemberCount      int          `json:"approximate_member_count,omitempty"`
	ApproximatePresenceCount    int          `json:"approximate_presence_count,omitempty"`
}

func (g Guild) HasFeature(feature string) bool {
	for _, have := range g.Features {
		if have == feature {
			return true
		}
	}
	return false
}

const (
	VerificationNone   = 0
	VerificationLow    = 1
	VerificationMedium = 2
	VerificationHigh   = 3

	MFALevelNone     = 0
	MFALevelElevated = 1

	ExplicitFilterDisabled            = 0
	ExplicitFilterMembersWithoutRoles = 1
	ExplicitFilterAllMembers          = 2

	NotifyAllMessages  = 0
	NotifyOnlyMentions = 1

	SystemChannelSuppressJoinNotifications = 1 << 0
)

const (
	FeatureAnimatedBanner           = "ANIMATED_BANNER"
	FeatureAnimatedIcon             = "ANIMATED_ICON"
	FeatureBanner                   = "BANNER"
	FeatureCloneEmojiEnabled        = "CLONE_EMOJI_ENABLED"
	FeatureCloneStickerEnabled      = "CLONE_STICKER_ENABLED"
	FeatureDetachedBanner           = "DETACHED_BANNER"
	FeatureDiscoverable             = "DISCOVERABLE"
	FeatureExpressionPurgeAllowed   = "EXPRESSION_PURGE_ALLOWED"
	FeatureHideOwnerCrown           = "HIDE_OWNER_CROWN"
	FeatureInviteSplash             = "INVITE_SPLASH"
	FeatureInvitesDisabled          = "INVITES_DISABLED"
	FeaturePartnered                = "PARTNERED"
	FeatureTextChannelFlexibleNames = "TEXT_CHANNEL_FLEXIBLE_NAMES"
	FeatureVanityURL                = "VANITY_URL"
	FeatureVerified                 = "VERIFIED"
	FeatureVisitsEnabled            = "VISITS_ENABLED"
	FeatureVoiceE2EE                = "VOICE_E2EE"
)

type GuildEdit struct {
	Name                        Optional[string]
	Icon                        Optional[string]
	Banner                      Optional[string]
	Splash                      Optional[string]
	EmbedSplash                 Optional[string]
	SplashCardAlignment         Optional[int]
	SystemChannelID             Optional[string]
	SystemChannelFlags          Optional[int]
	AFKChannelID                Optional[string]
	AFKTimeout                  Optional[int]
	DefaultMessageNotifications Optional[int]
	VerificationLevel           Optional[int]
	NSFW                        Optional[bool]
	ContentWarningLevel         Optional[int]
	ContentWarningText          Optional[string]
	ExplicitContentFilter       Optional[int]
	Features                    Optional[[]string]
	MessageHistoryCutoff        Optional[string]
}

type GuildQuery struct {
	Before     string
	After      string
	Limit      int
	WithCounts bool
}

type VanityURL struct {
	Code *string `json:"code"`
	Uses int     `json:"uses,omitempty"`
}

type BanEntry struct {
	User        User    `json:"user"`
	Reason      *string `json:"reason"`
	ModeratorID string  `json:"moderator_id,omitempty"`
	BannedAt    string  `json:"banned_at,omitempty"`
	ExpiresAt   *string `json:"expires_at"`
}

const (
	InviteTypeGuild   = 0
	InviteTypeGroupDM = 1
	MaxInviteUses     = 100
	MaxInviteAge      = 604800
)

type InviteCreate struct {
	MaxUses   int
	MaxAge    int
	Unique    bool
	Temporary bool
}

type Sticker struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Animated    bool     `json:"animated,omitempty"`
	User        *User    `json:"user,omitempty"`
}

type EmojiCreate struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

type StickerCreate struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Image       string   `json:"image"`
}

type StickerEdit struct {
	Name        string
	Description string
	Tags        []string
}

type ExpressionFailure struct {
	Name  string `json:"name"`
	Error string `json:"error"`
}

type EmojiBulkResult struct {
	Success []Emoji             `json:"success"`
	Failed  []ExpressionFailure `json:"failed"`
}

type StickerBulkResult struct {
	Success []Sticker           `json:"success"`
	Failed  []ExpressionFailure `json:"failed"`
}

type ExpressionMetadata struct {
	ID           string `json:"id"`
	GuildID      string `json:"guild_id"`
	Name         string `json:"name"`
	Animated     bool   `json:"animated"`
	AllowCloning bool   `json:"allow_cloning"`
}

const (
	MaxBulkExpressions = 50
	MaxEmojiName       = 32
	MaxStickerName     = 30
)

type GatewayInfo struct {
	URL               string            `json:"url"`
	Shards            int               `json:"shards"`
	SessionStartLimit SessionStartLimit `json:"session_start_limit"`
}

type SessionStartLimit struct {
	Total          int `json:"total"`
	Remaining      int `json:"remaining"`
	ResetAfter     int `json:"reset_after"`
	MaxConcurrency int `json:"max_concurrency"`
}

type Invite struct {
	Code          string   `json:"code"`
	Type          int      `json:"type"`
	GuildID       string   `json:"guild_id,omitempty"`
	ChannelID     string   `json:"channel_id,omitempty"`
	Guild         *Guild   `json:"guild,omitempty"`
	Channel       *Channel `json:"channel,omitempty"`
	Inviter       *User    `json:"inviter,omitempty"`
	Uses          int      `json:"uses"`
	MaxUses       int      `json:"max_uses"`
	MaxAge        int      `json:"max_age"`
	Temporary     bool     `json:"temporary"`
	CreatedAt     string   `json:"created_at,omitempty"`
	ExpiresAt     *string  `json:"expires_at,omitempty"`
	MemberCount   int      `json:"member_count,omitempty"`
	PresenceCount int      `json:"presence_count,omitempty"`
}
