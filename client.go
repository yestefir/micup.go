package micup

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	DefaultOrigin        = "https://micup.gg"
	defaultHTTPTimeout   = 15 * time.Second
	discoveryLimitBytes  = 1 << 20
	messageContentIntent = (1 << 18) | (1 << 19)
)

type Config struct {
	Token      string
	Origin     string
	HTTP       *http.Client
	ShardID    int
	ShardCount int
}

type Client struct {
	config                Config
	apiURL                string
	gatewayURL            string
	BotID                 string
	MessageContentEnabled *bool
}

func New(ctx context.Context, config Config) (*Client, error) {
	config, origin, err := normalizeConfig(config)
	if err != nil {
		return nil, err
	}
	apiURL, gatewayURL, err := discover(ctx, config.HTTP, origin)
	if err != nil {
		return nil, err
	}
	c := &Client{config: config, apiURL: apiURL, gatewayURL: gatewayURL}
	if err = c.identifyBot(ctx); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Client) Gateway(ctx context.Context, handler EventHandler) error {
	if handler == nil {
		return errors.New("Gateway event handler is required")
	}
	return runConnection(ctx, c, handler, connectionHooks{
		userAgent:  defaultUserAgent,
		browser:    defaultBrowser,
		shardID:    c.config.ShardID,
		shardCount: c.config.ShardCount,
		resume:     &resumeState{},
	})
}

func (c *Client) GatewayEndpoint() string { return c.gatewayURL }

func normalizeConfig(config Config) (Config, *url.URL, error) {
	if config.Token == "" || strings.ContainsAny(config.Token, "\r\n") {
		return config, nil, errors.New("a valid bot token is required")
	}
	if config.Origin == "" {
		config.Origin = DefaultOrigin
	}
	if config.ShardCount == 0 {
		config.ShardCount = 1
	}
	if err := validateShard(config.ShardID, config.ShardCount); err != nil {
		return config, nil, err
	}
	origin, err := url.Parse(config.Origin)
	if err != nil || origin.Host == "" || origin.User != nil || (origin.Scheme != "http" && origin.Scheme != "https") {
		return config, nil, errors.New("invalid Micup origin")
	}
	if config.HTTP == nil {
		config.HTTP = &http.Client{Timeout: defaultHTTPTimeout}
	}
	httpClient := *config.HTTP
	httpClient.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	config.HTTP = &httpClient
	return config, origin, nil
}

func discover(ctx context.Context, client *http.Client, origin *url.URL) (string, string, error) {
	base := strings.TrimRight(origin.String(), "/")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/.well-known/fluxer", nil)
	if err != nil {
		return "", "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", errors.New("Micup discovery is unavailable")
	}
	defer resp.Body.Close()
	var discovery struct {
		Endpoints struct {
			API     string `json:"api_public"`
			Gateway string `json:"gateway"`
		} `json:"endpoints"`
	}
	if resp.StatusCode != http.StatusOK || json.NewDecoder(io.LimitReader(resp.Body, discoveryLimitBytes)).Decode(&discovery) != nil {
		return "", "", errors.New("invalid Micup discovery response")
	}
	api, err := url.Parse(discovery.Endpoints.API)
	if err != nil || api.Host != origin.Host || api.Scheme != origin.Scheme {
		return "", "", errors.New("Micup API origin mismatch")
	}
	gw, err := url.Parse(discovery.Endpoints.Gateway)
	if err != nil || gw.Host != origin.Host || (gw.Scheme != "ws" && gw.Scheme != "wss") || (origin.Scheme == "https" && gw.Scheme != "wss") {
		return "", "", errors.New("invalid Micup Gateway endpoint")
	}
	q := gw.Query()
	q.Set("v", "1")
	q.Set("encoding", "json")
	gw.RawQuery = q.Encode()
	return strings.TrimRight(api.String(), "/") + "/v1", gw.String(), nil
}

func (c *Client) identifyBot(ctx context.Context) error {
	var user struct {
		ID  string `json:"id"`
		Bot bool   `json:"bot"`
	}
	if err := c.REST(ctx, http.MethodGet, "/users/@me", nil, &user); err != nil {
		return err
	}
	if user.ID == "" || !user.Bot {
		return errors.New("credential must belong to a Micup bot")
	}
	c.BotID = user.ID
	var app struct {
		Flags *uint64 `json:"flags"`
	}
	if c.REST(ctx, http.MethodGet, "/applications/@me", nil, &app) == nil && app.Flags != nil {
		enabled := *app.Flags&messageContentIntent != 0
		c.MessageContentEnabled = &enabled
	}
	return nil
}
