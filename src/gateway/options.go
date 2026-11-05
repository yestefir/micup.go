package gateway

import (
	"context"
	"errors"
	"net/url"

	"github.com/yestefir/micup.go/src/internal/version"
)

type Options struct {
	URL            string
	Token          string
	ShardID        int
	ShardCount     int
	UserAgent      string
	Browser        string
	Resume         *Resume
	OnConnect      func(*Conn)
	OnDisconnect   func()
	BeforeIdentify func(context.Context) error
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
