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
	EventReady             = gateway.EventReady
	EventResumed           = gateway.EventResumed
	EventGuildCreate       = gateway.EventGuildCreate
	EventGuildDelete       = gateway.EventGuildDelete
	EventChannelCreate     = gateway.EventChannelCreate
	EventChannelUpdate     = gateway.EventChannelUpdate
	EventChannelDelete     = gateway.EventChannelDelete
	EventMessageCreate     = gateway.EventMessageCreate
	EventInteractionCreate = gateway.EventInteractionCreate
	EventVoiceStateUpdate  = gateway.EventVoiceStateUpdate
	EventVoiceServerUpdate = gateway.EventVoiceServerUpdate
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
	VoiceChannelFull        = gateway.VoiceChannelFull
	VoiceP2PConsentRequired = gateway.VoiceP2PConsentRequired
	VoiceP2PUnavailable     = gateway.VoiceP2PUnavailable
)

type (
	Session           = session.Session
	SessionConfig     = session.Config
	SessionOptions    = session.Options
	VoiceJoin         = session.VoiceJoin
	VoiceRefusedError = session.VoiceRefusedError
	ChannelLookup     = session.ChannelLookup
)

func NewSession(config SessionConfig, options SessionOptions) (*Session, error) {
	return session.New(config, options)
}

const (
	DefaultMinBackoff = session.DefaultMinBackoff
	DefaultMaxBackoff = session.DefaultMaxBackoff
)

var (
	ErrNotReady         = session.ErrNotReady
	ErrVoiceJoinPending = session.ErrVoiceJoinPending
	ErrVoiceGrant       = session.ErrVoiceGrant
	ErrVoiceMoved       = session.ErrVoiceMoved
	ErrVoiceRefused     = session.ErrVoiceRefused
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
