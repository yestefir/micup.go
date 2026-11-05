package rest

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/yestefir/micup.go/src/internal/httpx"
)

const (
	APIVersion     = "v1"
	maxAttempts    = 5
	maxRetryAfter  = 30 * time.Second
	minRetryAfter  = time.Millisecond
	serverBackoff  = 500 * time.Millisecond
	responseLimit  = 2 << 20
	errorBodyLimit = 64 << 10
	discardLimit   = 8192
)

type Config struct {
	BaseURL string
	Token   string
	HTTP    *http.Client
}

type Client struct {
	baseURL string
	token   string
	http    *http.Client
	limits  *limiter
}

func ValidToken(token string) bool { return token != "" && !strings.ContainsAny(token, "\r\n") }

func New(config Config) (*Client, error) {
	if !ValidToken(config.Token) {
		return nil, errors.New("a valid bot token is required")
	}
	base, err := url.Parse(config.BaseURL)
	if err != nil || base.Host == "" || base.User != nil || (base.Scheme != "http" && base.Scheme != "https") {
		return nil, errors.New("invalid Micup API URL")
	}
	return &Client{
		baseURL: strings.TrimRight(base.String(), "/") + "/" + APIVersion,
		token:   config.Token,
		http:    httpx.NoRedirects(config.HTTP),
		limits:  newLimiter(),
	}, nil
}

func (c *Client) Do(ctx context.Context, method, path string, body, dest any) error {
	return c.call(ctx, &request{method: method, path: path, body: body, dest: dest})
}

func sleep(ctx context.Context, wait time.Duration) error {
	if wait <= 0 {
		return ctx.Err()
	}
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func path(parts ...string) string {
	var b strings.Builder
	for _, part := range parts {
		b.WriteByte('/')
		b.WriteString(part)
	}
	return b.String()
}
