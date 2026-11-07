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
