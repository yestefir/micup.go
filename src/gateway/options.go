package gateway

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/yestefir/micup.go/src/internal/version"
	"github.com/yestefir/micup.go/src/model"
)

const (
	FlagDebounceMessageReactions = 1 << 1
	FlagChannelThreads           = 1 << 2
	FlagGuildVisits              = 1 << 3
	maxIgnoredEvents             = 256
)

type Options struct {
	URL                  string
	Token                string
	ShardID              int
	ShardCount           int
	UserAgent            string
	Browser              string
	Resume               *Resume
	IgnoredEvents        []string
	Presence             *model.Presence
	PresenceFunc         func() *model.Presence
	Flags                int
	Compress             bool
	OnConnect            func(*Conn)
	OnDisconnect         func()
	OnSessionInvalidated func()
	OnGatewayError       func(GatewayError)
	BeforeIdentify       func(context.Context) error
}

func ValidateShard(shardID, shardCount int) error {
	if shardCount < 1 || shardID < 0 || shardID >= shardCount {
		return errors.New("invalid shard configuration")
	}
	return nil
}

func (o Options) normalize() (Options, error) {
	if o.Token == "" {
		return o, errors.New("Gateway token is required")
	}
	if err := ValidateIgnoredEvents(o.IgnoredEvents); err != nil {
		return o, err
	}
	if o.Flags < 0 {
		return o, errors.New("Gateway session flags must not be negative")
	}
	if o.Presence != nil && !model.ValidStatus(o.Presence.Status) {
		return o, fmt.Errorf("invalid presence status %q", o.Presence.Status)
	}
	if o.ShardCount == 0 {
		o.ShardCount = 1
	}
	if err := ValidateShard(o.ShardID, o.ShardCount); err != nil {
		return o, err
	}
	endpoint, err := url.Parse(o.URL)
	if err != nil || endpoint.Host == "" || (endpoint.Scheme != "ws" && endpoint.Scheme != "wss") {
		return o, errors.New("invalid Micup Gateway endpoint")
	}
	q := endpoint.Query()
	q.Set("v", ProtocolVersion)
	q.Set("encoding", Encoding)
	if o.Compress {
		q.Set("compress", CompressionZstdStream)
		q.Set("stream", "1")
	} else {
		q.Del("compress")
		q.Del("stream")
	}
	endpoint.RawQuery = q.Encode()
	o.URL = endpoint.String()
	if o.UserAgent == "" {
		o.UserAgent = version.UserAgent
	}
	if o.Browser == "" {
		o.Browser = version.Browser
	}
	if o.Resume == nil {
		o.Resume = &Resume{}
	}
	return o, nil
}

func ValidateIgnoredEvents(events []string) error {
	if len(events) > maxIgnoredEvents {
		return fmt.Errorf("at most %d ignored events are allowed", maxIgnoredEvents)
	}
	for _, name := range events {
		if name == "" || name != strings.ToUpper(name) || strings.ContainsAny(name, " \t\r\n") {
			return fmt.Errorf("invalid ignored event name %q", name)
		}
	}
	return nil
}

func (o Options) presence() *model.Presence {
	if o.PresenceFunc != nil {
		if presence := o.PresenceFunc(); presence != nil {
			return presence
		}
	}
	return o.Presence
}
