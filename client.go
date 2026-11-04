package micup

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const DefaultOrigin = "https://micup.gg"

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

type APIError struct{ StatusCode int }

type Message struct {
	ID        string  `json:"id"`
	ChannelID string  `json:"channel_id"`
	GuildID   string  `json:"guild_id,omitempty"`
	Content   string  `json:"content"`
	Embeds    []Embed `json:"embeds,omitempty"`
	Flags     int     `json:"flags,omitempty"`
}

func (e *APIError) Error() string { return fmt.Sprintf("Micup API returned HTTP %d", e.StatusCode) }

func New(ctx context.Context, config Config) (*Client, error) {
	if config.Token == "" || strings.ContainsAny(config.Token, "\r\n") {
		return nil, errors.New("a valid bot token is required")
	}
	if config.Origin == "" {
		config.Origin = DefaultOrigin
	}
	if config.ShardCount == 0 {
		config.ShardCount = 1
	}
	if config.ShardCount < 1 || config.ShardID < 0 || config.ShardID >= config.ShardCount {
		return nil, errors.New("invalid shard configuration")
	}
	origin, err := url.Parse(config.Origin)
	if err != nil || origin.Host == "" || origin.User != nil || (origin.Scheme != "http" && origin.Scheme != "https") {
		return nil, errors.New("invalid Micup origin")
	}
	if config.HTTP == nil {
		config.HTTP = &http.Client{Timeout: 15 * time.Second}
	}
	httpClient := *config.HTTP
	httpClient.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	config.HTTP = &httpClient
	base := strings.TrimRight(config.Origin, "/")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/.well-known/fluxer", nil)
	if err != nil {
		return nil, err
	}
	resp, err := config.HTTP.Do(req)
	if err != nil {
		return nil, errors.New("Micup discovery is unavailable")
	}
	defer resp.Body.Close()
	var discovery struct {
		Endpoints struct {
			API     string `json:"api_public"`
			Gateway string `json:"gateway"`
		} `json:"endpoints"`
	}
	if resp.StatusCode != http.StatusOK || json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&discovery) != nil {
		return nil, errors.New("invalid Micup discovery response")
	}
	api, err := url.Parse(discovery.Endpoints.API)
	if err != nil || api.Host != origin.Host || api.Scheme != origin.Scheme {
		return nil, errors.New("Micup API origin mismatch")
	}
	gw, err := url.Parse(discovery.Endpoints.Gateway)
	if err != nil || gw.Host != origin.Host || (gw.Scheme != "ws" && gw.Scheme != "wss") || (origin.Scheme == "https" && gw.Scheme != "wss") {
		return nil, errors.New("invalid Micup Gateway endpoint")
	}
	q := gw.Query()
	q.Set("v", "1")
	q.Set("encoding", "json")
	gw.RawQuery = q.Encode()
	c := &Client{config: config, apiURL: strings.TrimRight(api.String(), "/") + "/v1", gatewayURL: gw.String()}
	var user struct {
		ID  string `json:"id"`
		Bot bool   `json:"bot"`
	}
	if err = c.REST(ctx, http.MethodGet, "/users/@me", nil, &user); err != nil {
		return nil, err
	}
	if user.ID == "" || !user.Bot {
		return nil, errors.New("credential must belong to a Micup bot")
	}
	c.BotID = user.ID
	var app struct {
		Flags *uint64 `json:"flags"`
	}
	if c.REST(ctx, http.MethodGet, "/applications/@me", nil, &app) == nil && app.Flags != nil {
		enabled := *app.Flags&((1<<18)|(1<<19)) != 0
		c.MessageContentEnabled = &enabled
	}
	return c, nil
}

func (c *Client) REST(ctx context.Context, method, path string, body, dest any) error {
	if !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") {
		return errors.New("REST path must be absolute to the API")
	}
	var raw []byte
	var err error
	if body != nil {
		raw, err = json.Marshal(body)
		if err != nil {
			return err
		}
	}
	for attempt := 0; attempt < 4; attempt++ {
		req, requestErr := http.NewRequestWithContext(ctx, method, c.apiURL+path, bytes.NewReader(raw))
		if requestErr != nil {
			return requestErr
		}
		req.Header.Set("Authorization", "Bot "+c.config.Token)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "micup.go/0.2")
		resp, requestErr := c.config.HTTP.Do(req)
		if requestErr != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return errors.New("Micup API connection failed")
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			var rate struct {
				RetryAfter float64 `json:"retry_after"`
			}
			_ = json.NewDecoder(io.LimitReader(resp.Body, 8192)).Decode(&rate)
			_ = resp.Body.Close()
			wait := time.Duration(rate.RetryAfter * float64(time.Second))
			if wait < time.Second {
				wait = time.Second
			}
			if wait > 30*time.Second {
				return &APIError{StatusCode: 429}
			}
			timer := time.NewTimer(wait)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
			continue
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			_ = resp.Body.Close()
			return &APIError{StatusCode: resp.StatusCode}
		}
		if dest != nil && resp.StatusCode != http.StatusNoContent {
			err = json.NewDecoder(io.LimitReader(resp.Body, 2<<20)).Decode(dest)
		} else {
			_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 8192))
		}
		_ = resp.Body.Close()
		return err
	}
	return &APIError{StatusCode: 429}
}

func (c *Client) Gateway(ctx context.Context, handler EventHandler) error {
	return runGateway(ctx, c, handler)
}

func (c *Client) GatewayEndpoint() string { return c.gatewayURL }

func (c *Client) SendMessage(ctx context.Context, channelID, content, replyTo string) (Message, error) {
	if content == "" {
		return Message{}, errors.New("message content is required")
	}
	return c.SendMessageWithOptions(ctx, channelID, MessageOptions{Content: content, ReplyTo: replyTo})
}

func (c *Client) SendMessageWithOptions(ctx context.Context, channelID string, options MessageOptions) (Message, error) {
	if channelID == "" {
		return Message{}, errors.New("channel ID is required")
	}
	if options.Content == "" && len(options.Embeds) == 0 {
		return Message{}, errors.New("message content or at least one embed is required")
	}
	if len(options.Embeds) > 10 {
		return Message{}, errors.New("a message can contain at most 10 embeds")
	}
	body := map[string]any{"allowed_mentions": map[string]any{"parse": []string{}}}
	if options.Content != "" {
		body["content"] = options.Content
	}
	if len(options.Embeds) != 0 {
		body["embeds"] = options.Embeds
	}
	if options.ReplyTo != "" {
		body["message_reference"] = map[string]string{"message_id": options.ReplyTo, "channel_id": channelID}
	}
	if options.Nonce != "" {
		body["nonce"] = options.Nonce
	}
	if options.EnforceNonce {
		body["enforce_nonce"] = true
	}
	var message Message
	err := c.REST(ctx, http.MethodPost, "/channels/"+url.PathEscape(channelID)+"/messages", body, &message)
	return message, err
}

func (c *Client) RespondToInteraction(ctx context.Context, interaction Interaction, options MessageOptions, ephemeral bool) error {
	if err := validateInteraction(interaction); err != nil {
		return err
	}
	data, err := interactionMessageData(options, ephemeral)
	if err != nil {
		return err
	}
	body := map[string]any{"type": 4, "data": data}
	path := "/interactions/" + url.PathEscape(interaction.ID) + "/" + url.PathEscape(interaction.Token) + "/callback"
	return c.REST(ctx, http.MethodPost, path, body, nil)
}

func (c *Client) DeferInteraction(ctx context.Context, interaction Interaction, ephemeral bool) error {
	if err := validateInteraction(interaction); err != nil {
		return err
	}
	data := map[string]any{}
	if ephemeral {
		data["flags"] = 64
	}
	path := "/interactions/" + url.PathEscape(interaction.ID) + "/" + url.PathEscape(interaction.Token) + "/callback"
	return c.REST(ctx, http.MethodPost, path, map[string]any{"type": 5, "data": data}, nil)
}

func (c *Client) SendInteractionFollowup(ctx context.Context, interaction Interaction, options MessageOptions, ephemeral bool) (Message, error) {
	if interaction.ApplicationID == "" || interaction.Token == "" {
		return Message{}, errors.New("interaction application ID and token are required")
	}
	data, err := interactionMessageData(options, ephemeral)
	if err != nil {
		return Message{}, err
	}
	var message Message
	path := "/webhooks/" + url.PathEscape(interaction.ApplicationID) + "/" + url.PathEscape(interaction.Token)
	err = c.REST(ctx, http.MethodPost, path, data, &message)
	return message, err
}

func validateInteraction(interaction Interaction) error {
	if interaction.ID == "" || interaction.Token == "" {
		return errors.New("interaction ID and token are required")
	}
	return nil
}

func interactionMessageData(options MessageOptions, ephemeral bool) (map[string]any, error) {
	if options.Content == "" && len(options.Embeds) == 0 {
		return nil, errors.New("interaction response content or at least one embed is required")
	}
	if len(options.Embeds) > 10 {
		return nil, errors.New("an interaction response can contain at most 10 embeds")
	}
	data := map[string]any{"allowed_mentions": map[string]any{"parse": []string{}}}
	if options.Content != "" {
		data["content"] = options.Content
	}
	if len(options.Embeds) != 0 {
		data["embeds"] = options.Embeds
	}
	if ephemeral {
		data["flags"] = 64
	}
	return data, nil
}
