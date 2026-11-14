package bot

import (
	"context"
	"errors"
	"net/http"

	"github.com/yestefir/micup.go/src/discovery"
	"github.com/yestefir/micup.go/src/gateway"
	"github.com/yestefir/micup.go/src/internal/httpx"
	"github.com/yestefir/micup.go/src/internal/version"
	"github.com/yestefir/micup.go/src/model"
	"github.com/yestefir/micup.go/src/rest"
	"github.com/yestefir/micup.go/src/session"
	"github.com/yestefir/micup.go/src/shard"
)

const Version = version.Version

type Config struct {
	Token      string
	Origin     string
	HTTP       *http.Client
	ShardID    int
	ShardCount int
}

type Bot struct {
	REST                  *rest.Client
	ID                    string
	ApplicationID         string
	MessageContentEnabled *bool
	endpoints             discovery.Endpoints
	config                Config
}

func New(ctx context.Context, config Config) (*Bot, error) {
	if !rest.ValidToken(config.Token) {
		return nil, errors.New("a valid bot token is required")
	}
	if config.ShardCount == 0 {
		config.ShardCount = 1
	}
	if err := gateway.ValidateShard(config.ShardID, config.ShardCount); err != nil {
		return nil, err
	}
	origin, err := discovery.ParseOrigin(config.Origin)
	if err != nil {
		return nil, err
	}
	config.HTTP = httpx.NoRedirects(config.HTTP)
	endpoints, err := discovery.Discover(ctx, config.HTTP, origin)
	if err != nil {
		return nil, err
	}
	client, err := rest.New(rest.Config{BaseURL: endpoints.API, Token: config.Token, HTTP: config.HTTP})
	if err != nil {
		return nil, err
	}
	b := &Bot{REST: client, endpoints: endpoints, config: config}
	if err = b.identify(ctx); err != nil {
		return nil, err
	}
	return b, nil
}

func (b *Bot) identify(ctx context.Context) error {
	user, err := b.REST.CurrentUser(ctx)
	if err != nil {
		return err
	}
	if user.ID == "" || !user.Bot {
		return errors.New("credential must belong to a Micup bot")
	}
	b.ID = user.ID
	b.ApplicationID = user.ID
	if app, err := b.REST.CurrentApplication(ctx); err == nil {
		b.MessageContentEnabled = app.MessageContentEnabled()
		if app.ID != "" {
			b.ApplicationID = app.ID
		}
	}
	return nil
}

func (b *Bot) Endpoints() discovery.Endpoints { return b.endpoints }

func (b *Bot) Media() model.Media { return model.Media{Base: b.endpoints.Media} }

func (b *Bot) SessionConfig() session.Config {
	return session.Config{
		Endpoint:   b.endpoints.Gateway,
		Token:      b.config.Token,
		BotID:      b.ID,
		ShardID:    b.config.ShardID,
		ShardCount: b.config.ShardCount,
		Channels:   b.REST,
		Members:    b.REST,
	}
}

func (b *Bot) Gateway(ctx context.Context, handler gateway.Handler) error {
	return gateway.Run(ctx, gateway.Options{
		URL:        b.endpoints.Gateway,
		Token:      b.config.Token,
		ShardID:    b.config.ShardID,
		ShardCount: b.config.ShardCount,
	}, handler)
}

func (b *Bot) NewSession(options session.Options) (*session.Session, error) {
	return session.New(b.SessionConfig(), options)
}

func (b *Bot) NewShardManager(options shard.Options) (*shard.Manager, error) {
	return shard.NewManager(b.SessionConfig(), options)
}

const (
	MaxGuildsPerShard    = 2500
	targetGuildsPerShard = 2000
)

func (b *Bot) GuildCount(ctx context.Context) (int, error) {
	count := 0
	for _, err := range b.REST.EachGuild(ctx) {
		if err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func (b *Bot) RecommendedShards(ctx context.Context) (int, error) {
	guilds, err := b.GuildCount(ctx)
	if err != nil {
		return 0, err
	}
	shards := max((guilds+targetGuildsPerShard-1)/targetGuildsPerShard, 1)
	if info, err := b.REST.GatewayInfo(ctx); err == nil {
		shards = max(shards, info.Shards)
	}
	return shards, nil
}
