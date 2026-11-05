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
	RESTClient  = rest.Client
	RESTConfig  = rest.Config
	APIError    = rest.APIError
	Application = rest.Application
)

func NewREST(config RESTConfig) (*RESTClient, error) { return rest.New(config) }

type (
	Event          = gateway.Event
	EventHandler   = gateway.Handler
	GatewayOptions = gateway.Options
	GatewayConn    = gateway.Conn
	Resume         = gateway.Resume
	CloseError     = gateway.CloseError
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
	OpDispatch         = gateway.OpDispatch
	OpHeartbeat        = gateway.OpHeartbeat
	OpIdentify         = gateway.OpIdentify
	OpVoiceStateUpdate = gateway.OpVoiceStateUpdate
	OpResume           = gateway.OpResume
	OpReconnect        = gateway.OpReconnect
	OpInvalidSession   = gateway.OpInvalidSession
	OpHello            = gateway.OpHello
	OpHeartbeatAck     = gateway.OpHeartbeatAck
)

var (
	ErrReconnectRequested = gateway.ErrReconnectRequested
	ErrSessionInvalidated = gateway.ErrSessionInvalidated
	ErrGatewayClosed      = gateway.ErrClosed
)

type (
	Session        = session.Session
	SessionConfig  = session.Config
	SessionOptions = session.Options
	VoiceJoin      = session.VoiceJoin
	ChannelLookup  = session.ChannelLookup
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
	User              = model.User
	Guild             = model.Guild
	Channel           = model.Channel
	Message           = model.Message
	MessageCreate     = model.MessageCreate
	MessageOptions    = model.MessageOptions
	Embed             = model.Embed
	EmbedFooter       = model.EmbedFooter
	EmbedMedia        = model.EmbedMedia
	EmbedAuthor       = model.EmbedAuthor
	EmbedField        = model.EmbedField
	Interaction       = model.Interaction
	VoiceState        = model.VoiceState
	VoiceServerUpdate = model.VoiceServerUpdate
)

const (
	MessageTypeDefault   = model.MessageTypeDefault
	MessageTypeReply     = model.MessageTypeReply
	MessageFlagEphemeral = model.MessageFlagEphemeral
	MaxEmbeds            = model.MaxEmbeds
)
