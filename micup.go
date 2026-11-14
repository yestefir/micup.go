package micup

import (
	"context"
	"time"

	"github.com/yestefir/micup.go/src/bot"
	"github.com/yestefir/micup.go/src/discovery"
	"github.com/yestefir/micup.go/src/gateway"
	"github.com/yestefir/micup.go/src/model"
	"github.com/yestefir/micup.go/src/rest"
	"github.com/yestefir/micup.go/src/session"
	"github.com/yestefir/micup.go/src/shard"
	"github.com/yestefir/micup.go/src/state"
)

const Version = bot.Version

type (
	Config = bot.Config
	Bot    = bot.Bot
)

func New(ctx context.Context, config Config) (*Bot, error) { return bot.New(ctx, config) }

const DefaultOrigin = discovery.DefaultOrigin

type Endpoints = discovery.Endpoints

type (
	RESTClient    = rest.Client
	RESTConfig    = rest.Config
	APIError      = rest.APIError
	FieldError    = rest.FieldError
	Application   = rest.Application
	RequestOption = rest.RequestOption
	File          = rest.File
	Upload        = rest.Upload
)

func EmojiString(emoji Emoji) string { return rest.EmojiString(emoji) }

const (
	CodeInvalidFormBody       = rest.CodeInvalidFormBody
	CodeRateLimited           = rest.CodeRateLimited
	CodeSlowmodeRateLimited   = rest.CodeSlowmodeRateLimited
	CodeResourceLocked        = rest.CodeResourceLocked
	CodeUnauthorized          = rest.CodeUnauthorized
	CodeMissingAccess         = rest.CodeMissingAccess
	CodeMissingPermissions    = rest.CodeMissingPermissions
	CodeTwoFactorRequired     = rest.CodeTwoFactorRequired
	CodeCommunicationDisabled = rest.CodeCommunicationDisabled
	CodeContentBlocked        = rest.CodeContentBlocked
	CodeCannotSendEmpty       = rest.CodeCannotSendEmpty
	CodeUnknownChannel        = rest.CodeUnknownChannel
	CodeUnknownGuild          = rest.CodeUnknownGuild
	CodeUnknownMember         = rest.CodeUnknownMember
	CodeUnknownMessage        = rest.CodeUnknownMessage
	CodeUnknownRole           = rest.CodeUnknownRole
	CodeUnknownUser           = rest.CodeUnknownUser
	CodeUnknownInvite         = rest.CodeUnknownInvite
	CodeUnknownEmoji          = rest.CodeUnknownEmoji
	CodeUnknownSticker        = rest.CodeUnknownSticker
	CodeUnknownWebhook        = rest.CodeUnknownWebhook
	CodeThreadArchived        = rest.CodeThreadArchived
	CodeThreadLocked          = rest.CodeThreadLocked
	CodeMaxActiveThreads      = rest.CodeMaxActiveThreads
	CodeNSFWAgeRestricted     = rest.CodeNSFWAgeRestricted
	CodeFeatureDisabled       = rest.CodeFeatureDisabled
)

func ErrorCode(err error) string { return rest.ErrorCode(err) }

func IsErrorCode(err error, codes ...string) bool { return rest.IsErrorCode(err, codes...) }

func IsNotFound(err error) bool { return rest.IsNotFound(err) }

func IsForbidden(err error) bool { return rest.IsForbidden(err) }

type Optional[T any] = model.Optional[T]

func Set[T any](value T) Optional[T] { return model.Set(value) }

func Clear[T any]() Optional[T] { return model.Clear[T]() }

func ImageData(contentType string, data []byte) (string, error) {
	return model.ImageData(contentType, data)
}

func AttachmentURL(filename string) string { return model.AttachmentURL(filename) }

const SnowflakeEpoch = model.SnowflakeEpoch

func ParseSnowflake(id string) (uint64, error) { return model.ParseSnowflake(id) }

func SnowflakeTime(id string) (time.Time, error) { return model.SnowflakeTime(id) }

func SnowflakeAt(at time.Time) string { return model.SnowflakeAt(at) }

func CompareSnowflakes(a, b string) int { return model.CompareSnowflakes(a, b) }

func MentionUser(userID string) string { return model.MentionUser(userID) }

func MentionRole(roleID string) string { return model.MentionRole(roleID) }

func MentionChannel(channelID string) string { return model.MentionChannel(channelID) }

func FormatTimestamp(at time.Time, style string) string { return model.FormatTimestamp(at, style) }

func EscapeMarkdown(text string) string { return model.EscapeMarkdown(text) }

func NewREST(config RESTConfig) (*RESTClient, error) { return rest.New(config) }

func WithReason(reason string) RequestOption { return rest.WithReason(reason) }

type (
	Event          = gateway.Event
	EventHandler   = gateway.Handler
	GatewayOptions = gateway.Options
	GatewayConn    = gateway.Conn
	Resume         = gateway.Resume
	CloseError     = gateway.CloseError
	GatewayError   = gateway.GatewayError
)

func RunGateway(ctx context.Context, options GatewayOptions, handler EventHandler) error {
	return gateway.Run(ctx, options, handler)
}

const (
	OpDispatch                   = gateway.OpDispatch
	OpHeartbeat                  = gateway.OpHeartbeat
	OpIdentify                   = gateway.OpIdentify
	OpPresenceUpdate             = gateway.OpPresenceUpdate
	OpVoiceStateUpdate           = gateway.OpVoiceStateUpdate
	OpRequestGuildMembers        = gateway.OpRequestGuildMembers
	OpGatewayError               = gateway.OpGatewayError
	OpLazyRequest                = gateway.OpLazyRequest
	OpRequestGuildCounts         = gateway.OpRequestGuildCounts
	OpRequestChannelMemberCounts = gateway.OpRequestChannelMemberCounts
	OpResume                     = gateway.OpResume
	OpReconnect                  = gateway.OpReconnect
	OpInvalidSession             = gateway.OpInvalidSession
	OpHello                      = gateway.OpHello
	OpHeartbeatAck               = gateway.OpHeartbeatAck
	OpVoiceSignal                = gateway.OpVoiceSignal
	OpRequestForumUnreads        = gateway.OpRequestForumUnreads
)

const (
	CloseUnknownError         = gateway.CloseUnknownError
	CloseUnknownOpcode        = gateway.CloseUnknownOpcode
	CloseDecodeError          = gateway.CloseDecodeError
	CloseNotAuthenticated     = gateway.CloseNotAuthenticated
	CloseAuthenticationFailed = gateway.CloseAuthenticationFailed
	CloseAlreadyAuthenticated = gateway.CloseAlreadyAuthenticated
	CloseInvalidSequence      = gateway.CloseInvalidSequence
	CloseRateLimited          = gateway.CloseRateLimited
	CloseSessionTimeout       = gateway.CloseSessionTimeout
	CloseInvalidShard         = gateway.CloseInvalidShard
	CloseShardingRequired     = gateway.CloseShardingRequired
	CloseInvalidAPIVersion    = gateway.CloseInvalidAPIVersion
)

func ValidateSubscriptions(subscriptions map[string]Subscription) error {
	return gateway.ValidateSubscriptions(subscriptions)
}

var (
	ErrReconnectRequested = gateway.ErrReconnectRequested
	ErrGatewayClosed      = gateway.ErrClosed
)

const (
	FlagDebounceMessageReactions = gateway.FlagDebounceMessageReactions
	FlagChannelThreads           = gateway.FlagChannelThreads
	FlagGuildVisits              = gateway.FlagGuildVisits
)

type (
	Presence     = model.Presence
	CustomStatus = model.CustomStatus
)

const (
	StatusOnline    = model.StatusOnline
	StatusIdle      = model.StatusIdle
	StatusDND       = model.StatusDND
	StatusInvisible = model.StatusInvisible
	StatusOffline   = model.StatusOffline

	MaxCustomStatusText  = model.MaxCustomStatusText
	MaxCustomStatusEmoji = model.MaxCustomStatusEmoji
)

const (
	VoiceChannelFull        = gateway.VoiceChannelFull
	VoiceP2PConsentRequired = gateway.VoiceP2PConsentRequired
	VoiceP2PUnavailable     = gateway.VoiceP2PUnavailable
)

type (
	Session                 = session.Session
	SessionConfig           = session.Config
	SessionOptions          = session.Options
	VoiceJoin               = session.VoiceJoin
	VoiceRefusedError       = session.VoiceRefusedError
	MembersRateLimitedError = session.MembersRateLimitedError
	ChannelLookup           = session.ChannelLookup
	MemberLookup            = session.MemberLookup
	GuildMembers            = session.GuildMembers
)

func NewSession(config SessionConfig, options SessionOptions) (*Session, error) {
	return session.New(config, options)
}

const (
	DefaultMinBackoff = session.DefaultMinBackoff
	DefaultMaxBackoff = session.DefaultMaxBackoff
)

var (
	ErrNotReady           = session.ErrNotReady
	ErrVoiceJoinPending   = session.ErrVoiceJoinPending
	ErrVoiceGrant         = session.ErrVoiceGrant
	ErrVoiceMoved         = session.ErrVoiceMoved
	ErrVoiceRefused       = session.ErrVoiceRefused
	ErrVoicePermission    = session.ErrVoicePermission
	ErrMembersRateLimited = session.ErrMembersRateLimited
	ErrNotInVoice         = session.ErrNotInVoice
	ErrNoResponse         = session.ErrNoResponse
)

type (
	State     = state.Cache
	VoiceMove = state.VoiceMove
)

func NewState(botID string) *State { return state.New(botID) }

type (
	ShardManager = shard.Manager
	ShardOptions = shard.Options
)

func NewShardManager(config SessionConfig, options ShardOptions) (*ShardManager, error) {
	return shard.NewManager(config, options)
}

func ShardForGuild(guildID string, shardCount int) (int, error) {
	return shard.ForGuild(guildID, shardCount)
}

const DefaultIdentifyInterval = shard.DefaultIdentifyInterval

const MaxGuildsPerShard = bot.MaxGuildsPerShard

var (
	ErrShardNotOwned       = shard.ErrNotOwned
	ErrShardManagerRunning = shard.ErrRunning
)

type (
	Webhook        = model.Webhook
	WebhookMessage = model.WebhookMessage
)

type (
	Permissions         = model.Permissions
	PermissionOverwrite = model.PermissionOverwrite
	Role                = model.Role
	MemberEdit          = model.MemberEdit
	MemberQuery         = model.MemberQuery
	Ban                 = model.Ban
	MembersRequest      = model.MembersRequest
)

func ComputePermissions(guildID, ownerID, userID string, roles map[string]Role, memberRoles []string, overwrites []PermissionOverwrite) Permissions {
	return model.ComputePermissions(guildID, ownerID, userID, roles, memberRoles, overwrites)
}

const (
	PermissionCreateInstantInvite   = model.PermissionCreateInstantInvite
	PermissionKickMembers           = model.PermissionKickMembers
	PermissionBanMembers            = model.PermissionBanMembers
	PermissionAdministrator         = model.PermissionAdministrator
	PermissionManageChannels        = model.PermissionManageChannels
	PermissionManageGuild           = model.PermissionManageGuild
	PermissionAddReactions          = model.PermissionAddReactions
	PermissionViewAuditLog          = model.PermissionViewAuditLog
	PermissionPrioritySpeaker       = model.PermissionPrioritySpeaker
	PermissionStream                = model.PermissionStream
	PermissionViewChannel           = model.PermissionViewChannel
	PermissionSendMessages          = model.PermissionSendMessages
	PermissionSendTTSMessages       = model.PermissionSendTTSMessages
	PermissionManageMessages        = model.PermissionManageMessages
	PermissionEmbedLinks            = model.PermissionEmbedLinks
	PermissionAttachFiles           = model.PermissionAttachFiles
	PermissionReadMessageHistory    = model.PermissionReadMessageHistory
	PermissionMentionEveryone       = model.PermissionMentionEveryone
	PermissionUseExternalEmojis     = model.PermissionUseExternalEmojis
	PermissionConnect               = model.PermissionConnect
	PermissionSpeak                 = model.PermissionSpeak
	PermissionMuteMembers           = model.PermissionMuteMembers
	PermissionDeafenMembers         = model.PermissionDeafenMembers
	PermissionMoveMembers           = model.PermissionMoveMembers
	PermissionUseVAD                = model.PermissionUseVAD
	PermissionChangeNickname        = model.PermissionChangeNickname
	PermissionManageNicknames       = model.PermissionManageNicknames
	PermissionManageRoles           = model.PermissionManageRoles
	PermissionManageWebhooks        = model.PermissionManageWebhooks
	PermissionManageExpressions     = model.PermissionManageExpressions
	PermissionManageThreads         = model.PermissionManageThreads
	PermissionCreatePublicThreads   = model.PermissionCreatePublicThreads
	PermissionCreatePrivateThreads  = model.PermissionCreatePrivateThreads
	PermissionUseExternalStickers   = model.PermissionUseExternalStickers
	PermissionSendMessagesInThreads = model.PermissionSendMessagesInThreads
	PermissionModerateMembers       = model.PermissionModerateMembers
	PermissionCreateExpressions     = model.PermissionCreateExpressions
	PermissionPinMessages           = model.PermissionPinMessages
	PermissionBypassSlowmode        = model.PermissionBypassSlowmode
	PermissionUpdateRTCRegion       = model.PermissionUpdateRTCRegion
	PermissionViewChannelMembers    = model.PermissionViewChannelMembers
	PermissionAll                   = model.PermissionAll
	OverwriteRole                   = model.OverwriteRole
	OverwriteMember                 = model.OverwriteMember
)

const (
	ChannelTypeGuildText          = model.ChannelTypeGuildText
	ChannelTypeDM                 = model.ChannelTypeDM
	ChannelTypeGuildVoice         = model.ChannelTypeGuildVoice
	ChannelTypeGroupDM            = model.ChannelTypeGroupDM
	ChannelTypeGuildCategory      = model.ChannelTypeGuildCategory
	ChannelTypeGuildAnnouncement  = model.ChannelTypeGuildAnnouncement
	ChannelTypeAnnouncementThread = model.ChannelTypeAnnouncementThread
	ChannelTypePublicThread       = model.ChannelTypePublicThread
	ChannelTypePrivateThread      = model.ChannelTypePrivateThread
	ChannelTypeGuildForum         = model.ChannelTypeGuildForum
	ChannelTypeGuildMedia         = model.ChannelTypeGuildMedia
	ChannelTypeGuildLink          = model.ChannelTypeGuildLink
	ChannelTypePersonalNotes      = model.ChannelTypePersonalNotes
)

type (
	User               = model.User
	Guild              = model.Guild
	Invite             = model.Invite
	RoleColors         = model.RoleColors
	Channel            = model.Channel
	Message            = model.Message
	MessageOptions     = model.MessageOptions
	MessageEdit        = model.MessageEdit
	MessageQuery       = model.MessageQuery
	MessageReference   = model.MessageReference
	Attachment         = model.Attachment
	UploadedAttachment = model.UploadedAttachment
	AllowedMentions    = model.AllowedMentions
	Emoji              = model.Emoji
	Reaction           = model.Reaction
	Member             = model.Member
	Pin                = model.Pin
	PinsPage           = model.PinsPage
	PinQuery           = model.PinQuery
	Embed              = model.Embed
	EmbedFooter        = model.EmbedFooter
	EmbedMedia         = model.EmbedMedia
	EmbedAuthor        = model.EmbedAuthor
	EmbedField         = model.EmbedField
	Interaction        = model.Interaction
	InteractionGuild   = model.InteractionGuild
	Modal              = model.Modal
	TextInput          = model.TextInput
	ModalData          = model.ModalData
	ApplicationCommand = model.ApplicationCommand
	CommandOption      = model.CommandOption
	CommandChoice      = model.CommandChoice
	CommandData        = model.CommandData
	CommandDataOption  = model.CommandDataOption
	Resolved           = model.Resolved
	ComponentData      = model.ComponentData
	ActionRow          = model.ActionRow
	Button             = model.Button
	SelectMenu         = model.SelectMenu
	SelectOption       = model.SelectOption
	SelectDefault      = model.SelectDefault
	VoiceState         = model.VoiceState
	VoiceServerUpdate  = model.VoiceServerUpdate
)

const (
	MessageTypeDefault              = model.MessageTypeDefault
	MessageTypeRecipientAdd         = model.MessageTypeRecipientAdd
	MessageTypeRecipientRemove      = model.MessageTypeRecipientRemove
	MessageTypeCall                 = model.MessageTypeCall
	MessageTypeChannelNameChange    = model.MessageTypeChannelNameChange
	MessageTypeChannelIconChange    = model.MessageTypeChannelIconChange
	MessageTypeChannelPinnedMessage = model.MessageTypeChannelPinnedMessage
	MessageTypeUserJoin             = model.MessageTypeUserJoin
	MessageTypeChannelFollowAdd     = model.MessageTypeChannelFollowAdd
	MessageTypeThreadCreated        = model.MessageTypeThreadCreated
	MessageTypeReply                = model.MessageTypeReply
	MessageTypeThreadStarterMessage = model.MessageTypeThreadStarterMessage
)

const (
	MessageFlagCrossposted           = model.MessageFlagCrossposted
	MessageFlagIsCrosspost           = model.MessageFlagIsCrosspost
	MessageFlagSuppressEmbeds        = model.MessageFlagSuppressEmbeds
	MessageFlagSourceMessageDeleted  = model.MessageFlagSourceMessageDeleted
	MessageFlagHasThread             = model.MessageFlagHasThread
	MessageFlagEphemeral             = model.MessageFlagEphemeral
	MessageFlagSuppressNotifications = model.MessageFlagSuppressNotifications
	MessageFlagVoiceMessage          = model.MessageFlagVoiceMessage
)

const (
	MaxEmbeds          = model.MaxEmbeds
	MaxAttachments     = model.MaxAttachments
	MaxBulkDelete      = model.MaxBulkDelete
	MaxMessagesPerPage = model.MaxMessagesPerPage
	MentionUsers       = model.MentionUsers
	MentionRoles       = model.MentionRoles
	MentionEveryone    = model.MentionEveryone
)

const (
	ComponentTypeActionRow = model.ComponentTypeActionRow
	ComponentTypeButton    = model.ComponentTypeButton
	ButtonStylePrimary     = model.ButtonStylePrimary
	ButtonStyleSecondary   = model.ButtonStyleSecondary
	ButtonStyleSuccess     = model.ButtonStyleSuccess
	ButtonStyleDanger      = model.ButtonStyleDanger
	ButtonStyleLink        = model.ButtonStyleLink
	MaxActionRows          = model.MaxActionRows
	MaxButtonsPerRow       = model.MaxButtonsPerRow
	MaxCustomIDLength      = model.MaxCustomIDLength
	MaxButtonLabelLength   = model.MaxButtonLabelLength
)

const (
	ComponentTypeStringSelect      = model.ComponentTypeStringSelect
	ComponentTypeUserSelect        = model.ComponentTypeUserSelect
	ComponentTypeRoleSelect        = model.ComponentTypeRoleSelect
	ComponentTypeMentionableSelect = model.ComponentTypeMentionableSelect
	ComponentTypeChannelSelect     = model.ComponentTypeChannelSelect
	SelectDefaultUser              = model.SelectDefaultUser
	SelectDefaultRole              = model.SelectDefaultRole
	SelectDefaultChannel           = model.SelectDefaultChannel
	MaxSelectOptions               = model.MaxSelectOptions
	MaxSelectPlaceholder           = model.MaxSelectPlaceholder
	MaxSelectOptionLabel           = model.MaxSelectOptionLabel
	MaxSelectOptionValue           = model.MaxSelectOptionValue
	MaxSelectOptionDescription     = model.MaxSelectOptionDescription
)

const (
	InteractionTypePing               = model.InteractionTypePing
	InteractionTypeApplicationCommand = model.InteractionTypeApplicationCommand
	InteractionTypeMessageComponent   = model.InteractionTypeMessageComponent
	InteractionTypeAutocomplete       = model.InteractionTypeAutocomplete
	InteractionTypeModalSubmit        = model.InteractionTypeModalSubmit
	InteractionContextGuild           = model.InteractionContextGuild
	InteractionContextBotDM           = model.InteractionContextBotDM
	InteractionContextPrivateChannel  = model.InteractionContextPrivateChannel
	CommandTypeChatInput              = model.CommandTypeChatInput
	CommandTypeUser                   = model.CommandTypeUser
	CommandTypeMessage                = model.CommandTypeMessage
	OptionTypeSubcommand              = model.OptionTypeSubcommand
	OptionTypeSubcommandGroup         = model.OptionTypeSubcommandGroup
	OptionTypeString                  = model.OptionTypeString
	OptionTypeInteger                 = model.OptionTypeInteger
	OptionTypeBoolean                 = model.OptionTypeBoolean
	OptionTypeUser                    = model.OptionTypeUser
	OptionTypeChannel                 = model.OptionTypeChannel
	OptionTypeRole                    = model.OptionTypeRole
	OptionTypeMentionable             = model.OptionTypeMentionable
	OptionTypeNumber                  = model.OptionTypeNumber
	OptionTypeAttachment              = model.OptionTypeAttachment
	MaxCommandOptions                 = model.MaxCommandOptions
	MaxCommandChoices                 = model.MaxCommandChoices
	MaxCommandName                    = model.MaxCommandName
	MaxCommandDescription             = model.MaxCommandDescription
	ComponentTypeTextInput            = model.ComponentTypeTextInput
	TextInputStyleShort               = model.TextInputStyleShort
	TextInputStyleParagraph           = model.TextInputStyleParagraph
	MaxModalInputs                    = model.MaxModalInputs
	MaxModalTitle                     = model.MaxModalTitle
	MaxTextInputLabel                 = model.MaxTextInputLabel
	MaxTextInputLength                = model.MaxTextInputLength
)

func NewActionRow(buttons ...Button) ActionRow { return model.NewActionRow(buttons...) }

func PrimaryButton(customID, label string) Button { return model.PrimaryButton(customID, label) }

func SecondaryButton(customID, label string) Button { return model.SecondaryButton(customID, label) }

func SuccessButton(customID, label string) Button { return model.SuccessButton(customID, label) }

func DangerButton(customID, label string) Button { return model.DangerButton(customID, label) }

func LinkButton(url, label string) Button { return model.LinkButton(url, label) }

func DisableButtons(rows []ActionRow) []ActionRow { return model.DisableButtons(rows) }

func DisableComponents(rows []ActionRow) []ActionRow { return model.DisableComponents(rows) }

func NewSelectRow(menu SelectMenu) ActionRow { return model.NewSelectRow(menu) }

func StringSelect(customID string, options ...SelectOption) SelectMenu {
	return model.StringSelect(customID, options...)
}

func UserSelect(customID string) SelectMenu { return model.UserSelect(customID) }

func RoleSelect(customID string) SelectMenu { return model.RoleSelect(customID) }

func MentionableSelect(customID string) SelectMenu { return model.MentionableSelect(customID) }

func ChannelSelect(customID string, channelTypes ...int) SelectMenu {
	return model.ChannelSelect(customID, channelTypes...)
}

func NewSelectOption(label, value string) SelectOption { return model.NewSelectOption(label, value) }

func DefaultUser(id string) SelectDefault { return model.DefaultUser(id) }

func DefaultRole(id string) SelectDefault { return model.DefaultRole(id) }

func DefaultChannel(id string) SelectDefault { return model.DefaultChannel(id) }

func ShortInput(customID, label string) TextInput { return model.ShortInput(customID, label) }

func ParagraphInput(customID, label string) TextInput { return model.ParagraphInput(customID, label) }

type (
	ThreadMetadata           = model.ThreadMetadata
	ThreadMember             = model.ThreadMember
	ForumTag                 = model.ForumTag
	DefaultReaction          = model.DefaultReaction
	ChannelEdit              = model.ChannelEdit
	ChannelCreate            = model.ChannelCreate
	ChannelPosition          = model.ChannelPosition
	FollowedChannel          = model.FollowedChannel
	ChannelFollowerStats     = model.ChannelFollowerStats
	ThreadStart              = model.ThreadStart
	ForumPostStart           = model.ForumPostStart
	ForumPost                = model.ForumPost
	ThreadEdit               = model.ThreadEdit
	ThreadList               = model.ThreadList
	ArchivedThreadQuery      = model.ArchivedThreadQuery
	ThreadSearch             = model.ThreadSearch
	ThreadMemberQuery        = model.ThreadMemberQuery
	ForumPostData            = model.ForumPostData
	GuildEdit                = model.GuildEdit
	GuildQuery               = model.GuildQuery
	VanityURL                = model.VanityURL
	BanEntry                 = model.BanEntry
	InviteCreate             = model.InviteCreate
	Sticker                  = model.Sticker
	EmojiCreate              = model.EmojiCreate
	StickerCreate            = model.StickerCreate
	StickerEdit              = model.StickerEdit
	ExpressionFailure        = model.ExpressionFailure
	EmojiBulkResult          = model.EmojiBulkResult
	StickerBulkResult        = model.StickerBulkResult
	ExpressionMetadata       = model.ExpressionMetadata
	GatewayInfo              = model.GatewayInfo
	SessionStartLimit        = model.SessionStartLimit
	CurrentMemberEdit        = model.CurrentMemberEdit
	MemberSearch             = model.MemberSearch
	MemberSearchResult       = model.MemberSearchResult
	MemberSearchSupplemental = model.MemberSearchSupplemental
	MemberSearchPage         = model.MemberSearchPage
	UserProfileQuery         = model.UserProfileQuery
	UserProfile              = model.UserProfile
	ProfileDetails           = model.ProfileDetails
	ProfileBadge             = model.ProfileBadge
	MutualGuild              = model.MutualGuild
	WebhookSource            = model.WebhookSource
	WebhookEdit              = model.WebhookEdit
	RoleCreate               = model.RoleCreate
	RoleEdit                 = model.RoleEdit
	RolePosition             = model.RolePosition
	RoleHoistPosition        = model.RoleHoistPosition
	ChannelMention           = model.ChannelMention
	StickerItem              = model.StickerItem
	MessageSnapshot          = model.MessageSnapshot
	MessageCall              = model.MessageCall
	MessageForward           = model.MessageForward
	KeptAttachment           = model.KeptAttachment
	ReactionUsersPage        = model.ReactionUsersPage
	ReactionUsersQuery       = model.ReactionUsersQuery
	RefreshedURL             = model.RefreshedURL
	CrosspostSource          = model.CrosspostSource
	CrosspostSourceGuild     = model.CrosspostSourceGuild
	MessageSearch            = model.MessageSearch
	MessageSearchResult      = model.MessageSearchResult
	EmbedProvider            = model.EmbedProvider
	AuditLogEntry            = model.AuditLogEntry
	AuditLogWebhook          = model.AuditLogWebhook
	AuditLog                 = model.AuditLog
	AuditLogQuery            = model.AuditLogQuery
)

const (
	ChannelFlagPinned                           = model.ChannelFlagPinned
	ChannelFlagRequireTag                       = model.ChannelFlagRequireTag
	ChannelFlagHideMediaDownloadOptions         = model.ChannelFlagHideMediaDownloadOptions
	AutoArchiveHour                             = model.AutoArchiveHour
	AutoArchiveDay                              = model.AutoArchiveDay
	AutoArchiveThreeDays                        = model.AutoArchiveThreeDays
	AutoArchiveWeek                             = model.AutoArchiveWeek
	ForumSortLatestActivity                     = model.ForumSortLatestActivity
	ForumSortCreationTime                       = model.ForumSortCreationTime
	ForumLayoutDefault                          = model.ForumLayoutDefault
	ForumLayoutList                             = model.ForumLayoutList
	ForumLayoutGrid                             = model.ForumLayoutGrid
	ForumTagMatchSome                           = model.ForumTagMatchSome
	ForumTagMatchAll                            = model.ForumTagMatchAll
	ThreadMemberHasInteracted                   = model.ThreadMemberHasInteracted
	ThreadMemberAllMessages                     = model.ThreadMemberAllMessages
	ThreadMemberOnlyMentions                    = model.ThreadMemberOnlyMentions
	ThreadMemberNoMessages                      = model.ThreadMemberNoMessages
	MaxChannelName                              = model.MaxChannelName
	MaxChannelTopic                             = model.MaxChannelTopic
	MaxForumTopic                               = model.MaxForumTopic
	MaxSlowmodeSeconds                          = model.MaxSlowmodeSeconds
	MaxForumTags                                = model.MaxForumTags
	MaxAppliedTags                              = model.MaxAppliedTags
	MaxForumTagName                             = model.MaxForumTagName
	MaxChannelPositions                         = model.MaxChannelPositions
	EmbedTypeRich                               = model.EmbedTypeRich
	MaxEmbedTitle                               = model.MaxEmbedTitle
	MaxEmbedDescription                         = model.MaxEmbedDescription
	MaxEmbedFields                              = model.MaxEmbedFields
	MaxEmbedFieldName                           = model.MaxEmbedFieldName
	MaxEmbedFieldValue                          = model.MaxEmbedFieldValue
	MaxEmbedFooter                              = model.MaxEmbedFooter
	MaxEmbedAuthor                              = model.MaxEmbedAuthor
	MaxEmbedURL                                 = model.MaxEmbedURL
	VerificationNone                            = model.VerificationNone
	VerificationLow                             = model.VerificationLow
	VerificationMedium                          = model.VerificationMedium
	VerificationHigh                            = model.VerificationHigh
	MFALevelNone                                = model.MFALevelNone
	MFALevelElevated                            = model.MFALevelElevated
	ExplicitFilterDisabled                      = model.ExplicitFilterDisabled
	ExplicitFilterMembersWithoutRoles           = model.ExplicitFilterMembersWithoutRoles
	ExplicitFilterAllMembers                    = model.ExplicitFilterAllMembers
	NotifyAllMessages                           = model.NotifyAllMessages
	NotifyOnlyMentions                          = model.NotifyOnlyMentions
	SystemChannelSuppressJoinNotifications      = model.SystemChannelSuppressJoinNotifications
	FeatureAnimatedBanner                       = model.FeatureAnimatedBanner
	FeatureAnimatedIcon                         = model.FeatureAnimatedIcon
	FeatureBanner                               = model.FeatureBanner
	FeatureCloneEmojiEnabled                    = model.FeatureCloneEmojiEnabled
	FeatureCloneStickerEnabled                  = model.FeatureCloneStickerEnabled
	FeatureDetachedBanner                       = model.FeatureDetachedBanner
	FeatureDiscoverable                         = model.FeatureDiscoverable
	FeatureExpressionPurgeAllowed               = model.FeatureExpressionPurgeAllowed
	FeatureHideOwnerCrown                       = model.FeatureHideOwnerCrown
	FeatureInviteSplash                         = model.FeatureInviteSplash
	FeatureInvitesDisabled                      = model.FeatureInvitesDisabled
	FeaturePartnered                            = model.FeaturePartnered
	FeatureTextChannelFlexibleNames             = model.FeatureTextChannelFlexibleNames
	FeatureVanityURL                            = model.FeatureVanityURL
	FeatureVerified                             = model.FeatureVerified
	FeatureVisitsEnabled                        = model.FeatureVisitsEnabled
	FeatureVoiceE2EE                            = model.FeatureVoiceE2EE
	InviteTypeGuild                             = model.InviteTypeGuild
	InviteTypeGroupDM                           = model.InviteTypeGroupDM
	MaxInviteUses                               = model.MaxInviteUses
	MaxInviteAge                                = model.MaxInviteAge
	MaxBulkExpressions                          = model.MaxBulkExpressions
	MaxEmojiName                                = model.MaxEmojiName
	MaxStickerName                              = model.MaxStickerName
	MemberProfileAvatarUnset                    = model.MemberProfileAvatarUnset
	MemberProfileBannerUnset                    = model.MemberProfileBannerUnset
	MentionNoPreference                         = model.MentionNoPreference
	MentionPreferMention                        = model.MentionPreferMention
	MentionPreferNoMention                      = model.MentionPreferNoMention
	JoinSourceCreator                           = model.JoinSourceCreator
	JoinSourceInstantInvite                     = model.JoinSourceInstantInvite
	JoinSourceVanityURL                         = model.JoinSourceVanityURL
	JoinSourceBotInvite                         = model.JoinSourceBotInvite
	JoinSourceAdminForceAdd                     = model.JoinSourceAdminForceAdd
	JoinSourceDiscovery                         = model.JoinSourceDiscovery
	MessageFlagFailedToMentionSomeRolesInThread = model.MessageFlagFailedToMentionSomeRolesInThread
	AttachmentFlagSpoiler                       = model.AttachmentFlagSpoiler
	AttachmentFlagExplicitMedia                 = model.AttachmentFlagExplicitMedia
	AttachmentFlagAnimated                      = model.AttachmentFlagAnimated
	MessageReferenceDefault                     = model.MessageReferenceDefault
	MessageReferenceForward                     = model.MessageReferenceForward
	MaxStickersPerMessage                       = model.MaxStickersPerMessage
	MaxReactionUsersPerPage                     = model.MaxReactionUsersPerPage
	MaxPinsPerPage                              = model.MaxPinsPerPage
	MaxRefreshURLs                              = model.MaxRefreshURLs
	SearchAuthorUser                            = model.SearchAuthorUser
	SearchAuthorBot                             = model.SearchAuthorBot
	SearchAuthorWebhook                         = model.SearchAuthorWebhook
	SearchHasImage                              = model.SearchHasImage
	SearchHasSound                              = model.SearchHasSound
	SearchHasVideo                              = model.SearchHasVideo
	SearchHasFile                               = model.SearchHasFile
	SearchHasSticker                            = model.SearchHasSticker
	SearchHasEmbed                              = model.SearchHasEmbed
	SearchHasLink                               = model.SearchHasLink
	SearchHasSnapshot                           = model.SearchHasSnapshot
	SearchSortTimestamp                         = model.SearchSortTimestamp
	SearchSortRelevance                         = model.SearchSortRelevance
	SortAscending                               = model.SortAscending
	SortDescending                              = model.SortDescending
	MaxSearchHitsPerPage                        = model.MaxSearchHitsPerPage
	MaxSearchPage                               = model.MaxSearchPage
	UserFlagStaff                               = model.UserFlagStaff
	UserFlagPartner                             = model.UserFlagPartner
	UserFlagBugHunter                           = model.UserFlagBugHunter
	UserFlagFriendlyBot                         = model.UserFlagFriendlyBot
	UserFlagFriendlyBotManualApproval           = model.UserFlagFriendlyBotManualApproval
	UserFlagSpammer                             = model.UserFlagSpammer
	UserFlagProfileHidden                       = model.UserFlagProfileHidden
	WebhookTypeIncoming                         = model.WebhookTypeIncoming
	WebhookTypeChannelFollower                  = model.WebhookTypeChannelFollower
	MaxWebhookName                              = model.MaxWebhookName
	AuditGuildUpdate                            = model.AuditGuildUpdate
	AuditChannelCreate                          = model.AuditChannelCreate
	AuditChannelUpdate                          = model.AuditChannelUpdate
	AuditChannelDelete                          = model.AuditChannelDelete
	AuditChannelOverwriteCreate                 = model.AuditChannelOverwriteCreate
	AuditChannelOverwriteUpdate                 = model.AuditChannelOverwriteUpdate
	AuditChannelOverwriteDelete                 = model.AuditChannelOverwriteDelete
	AuditMemberKick                             = model.AuditMemberKick
	AuditMemberPrune                            = model.AuditMemberPrune
	AuditMemberBanAdd                           = model.AuditMemberBanAdd
	AuditMemberBanRemove                        = model.AuditMemberBanRemove
	AuditMemberUpdate                           = model.AuditMemberUpdate
	AuditMemberRoleUpdate                       = model.AuditMemberRoleUpdate
	AuditMemberMove                             = model.AuditMemberMove
	AuditMemberDisconnect                       = model.AuditMemberDisconnect
	AuditBotAdd                                 = model.AuditBotAdd
	AuditRoleCreate                             = model.AuditRoleCreate
	AuditRoleUpdate                             = model.AuditRoleUpdate
	AuditRoleDelete                             = model.AuditRoleDelete
	AuditInviteCreate                           = model.AuditInviteCreate
	AuditInviteUpdate                           = model.AuditInviteUpdate
	AuditInviteDelete                           = model.AuditInviteDelete
	AuditWebhookCreate                          = model.AuditWebhookCreate
	AuditWebhookUpdate                          = model.AuditWebhookUpdate
	AuditWebhookDelete                          = model.AuditWebhookDelete
	AuditEmojiCreate                            = model.AuditEmojiCreate
	AuditEmojiUpdate                            = model.AuditEmojiUpdate
	AuditEmojiDelete                            = model.AuditEmojiDelete
	AuditMessageDelete                          = model.AuditMessageDelete
	AuditMessageBulkDelete                      = model.AuditMessageBulkDelete
	AuditMessagePin                             = model.AuditMessagePin
	AuditMessageUnpin                           = model.AuditMessageUnpin
	AuditStickerCreate                          = model.AuditStickerCreate
	AuditStickerUpdate                          = model.AuditStickerUpdate
	AuditStickerDelete                          = model.AuditStickerDelete
	AuditThreadCreate                           = model.AuditThreadCreate
	AuditThreadUpdate                           = model.AuditThreadUpdate
	AuditThreadDelete                           = model.AuditThreadDelete
	MaxAuditLogEntries                          = model.MaxAuditLogEntries
	MaxRoleName                                 = model.MaxRoleName
	MaxRoleColor                                = model.MaxRoleColor
	TimestampShortTime                          = model.TimestampShortTime
	TimestampLongTime                           = model.TimestampLongTime
	TimestampShortDate                          = model.TimestampShortDate
	TimestampLongDate                           = model.TimestampLongDate
	TimestampShortDateTime                      = model.TimestampShortDateTime
	TimestampLongDateTime                       = model.TimestampLongDateTime
	TimestampRelative                           = model.TimestampRelative
)

type (
	Media        = model.Media
	ImageOptions = model.ImageOptions
)

const (
	ImagePNG  = model.ImagePNG
	ImageJPEG = model.ImageJPEG
	ImageWebP = model.ImageWebP
	ImageGIF  = model.ImageGIF
	ImageAVIF = model.ImageAVIF

	ElevatedPermissions = model.ElevatedPermissions
)

func InviteURL(base, code string) (string, error) { return model.InviteURL(base, code) }

func NewEmbed() Embed { return model.NewEmbed() }

func ParsePermission(name string) (Permissions, bool) { return model.ParsePermission(name) }
