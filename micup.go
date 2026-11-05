package micup

import (
	"context"

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
	EventReady               = gateway.EventReady
	EventResumed             = gateway.EventResumed
	EventGuildCreate         = gateway.EventGuildCreate
	EventGuildSync           = gateway.EventGuildSync
	EventGuildUpdate         = gateway.EventGuildUpdate
	EventGuildDelete         = gateway.EventGuildDelete
	EventGuildRoleCreate     = gateway.EventGuildRoleCreate
	EventGuildRoleUpdate     = gateway.EventGuildRoleUpdate
	EventGuildRoleUpdateBulk = gateway.EventGuildRoleUpdateBulk
	EventGuildRoleDelete     = gateway.EventGuildRoleDelete
	EventGuildMemberAdd      = gateway.EventGuildMemberAdd
	EventGuildMemberUpdate   = gateway.EventGuildMemberUpdate
	EventGuildMemberRemove   = gateway.EventGuildMemberRemove
	EventGuildMembersChunk   = gateway.EventGuildMembersChunk
	EventGuildBanAdd         = gateway.EventGuildBanAdd
	EventGuildBanRemove      = gateway.EventGuildBanRemove
	EventChannelUpdateBulk   = gateway.EventChannelUpdateBulk
	EventChannelPinsUpdate   = gateway.EventChannelPinsUpdate
	EventMessageUpdate       = gateway.EventMessageUpdate
	EventMessageDelete       = gateway.EventMessageDelete
	EventMessageDeleteBulk   = gateway.EventMessageDeleteBulk
	EventReactionAdd         = gateway.EventReactionAdd
	EventReactionRemove      = gateway.EventReactionRemove
	EventReactionRemoveAll   = gateway.EventReactionRemoveAll
	EventReactionRemoveEmoji = gateway.EventReactionRemoveEmoji
	EventTypingStart         = gateway.EventTypingStart
	EventPresenceUpdate      = gateway.EventPresenceUpdate
	EventUserUpdate          = gateway.EventUserUpdate
	EventWebhooksUpdate      = gateway.EventWebhooksUpdate
	EventInviteCreate        = gateway.EventInviteCreate
	EventInviteDelete        = gateway.EventInviteDelete
	EventRateLimited         = gateway.EventRateLimited
	EventChannelCreate       = gateway.EventChannelCreate
	EventChannelUpdate       = gateway.EventChannelUpdate
	EventChannelDelete       = gateway.EventChannelDelete
	EventMessageCreate       = gateway.EventMessageCreate
	EventInteractionCreate   = gateway.EventInteractionCreate
	EventVoiceStateUpdate    = gateway.EventVoiceStateUpdate
	EventVoiceServerUpdate   = gateway.EventVoiceServerUpdate
)

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
)

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

var (
	ErrShardNotOwned       = shard.ErrNotOwned
	ErrShardManagerRunning = shard.ErrRunning
)

type (
	MessageDelete     = model.MessageDelete
	MessageDeleteBulk = model.MessageDeleteBulk
	ReactionEvent     = model.ReactionEvent
	TypingStart       = model.TypingStart
	GuildMemberRemove = model.GuildMemberRemove
	RoleEvent         = model.RoleEvent
	RoleDelete        = model.RoleDelete
	Webhook           = model.Webhook
	WebhookMessage    = model.WebhookMessage
)

type (
	Permissions         = model.Permissions
	PermissionOverwrite = model.PermissionOverwrite
	Role                = model.Role
	GuildMember         = model.GuildMember
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
	Channel            = model.Channel
	Message            = model.Message
	MessageCreate      = model.MessageCreate
	MessageUpdate      = model.MessageUpdate
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
	ComponentData      = model.ComponentData
	ActionRow          = model.ActionRow
	Button             = model.Button
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
	InteractionTypePing               = model.InteractionTypePing
	InteractionTypeApplicationCommand = model.InteractionTypeApplicationCommand
	InteractionTypeMessageComponent   = model.InteractionTypeMessageComponent
	InteractionTypeAutocomplete       = model.InteractionTypeAutocomplete
	InteractionTypeModalSubmit        = model.InteractionTypeModalSubmit
)

func NewActionRow(buttons ...Button) ActionRow { return model.NewActionRow(buttons...) }

func PrimaryButton(customID, label string) Button { return model.PrimaryButton(customID, label) }

func SecondaryButton(customID, label string) Button { return model.SecondaryButton(customID, label) }

func SuccessButton(customID, label string) Button { return model.SuccessButton(customID, label) }

func DangerButton(customID, label string) Button { return model.DangerButton(customID, label) }

func LinkButton(url, label string) Button { return model.LinkButton(url, label) }

func DisableButtons(rows []ActionRow) []ActionRow { return model.DisableButtons(rows) }
