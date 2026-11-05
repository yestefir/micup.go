package rest

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

	"github.com/yestefir/micup.go/src/internal/httpx"
	"github.com/yestefir/micup.go/src/internal/version"
)

const (
	APIVersion    = "v1"
	maxAttempts   = 4
	maxRetryAfter = 30 * time.Second
	minRetryAfter = time.Second
	responseLimit = 2 << 20
	discardLimit  = 8192
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
}

type APIError struct{ StatusCode int }

func (e *APIError) Error() string { return fmt.Sprintf("Micup API returned HTTP %d", e.StatusCode) }

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
	}, nil
}

func (c *Client) Do(ctx context.Context, method, path string, body, dest any) error {
	if !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") {
		return errors.New("REST path must be absolute to the API")
	}
	var raw []byte
	if body != nil {
		var err error
		if raw, err = json.Marshal(body); err != nil {
			return err
		}
	}
	for attempt := 0; attempt < maxAttempts; attempt++ {
		resp, err := c.send(ctx, method, path, raw)
		if err != nil {
			return err
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			wait := retryAfter(resp)
			if wait > maxRetryAfter {
				return &APIError{StatusCode: http.StatusTooManyRequests}
			}
			if err = sleep(ctx, wait); err != nil {
				return err
			}
			continue
		}
		return decode(resp, dest)
	}
	return &APIError{StatusCode: http.StatusTooManyRequests}
}

func (c *Client) send(ctx context.Context, method, path string, raw []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bot "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", version.UserAgent)
	resp, err := c.http.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, errors.New("Micup API connection failed")
	}
	return resp, nil
}

func decode(resp *http.Response, dest any) error {
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{StatusCode: resp.StatusCode}
	}
	if dest != nil && resp.StatusCode != http.StatusNoContent {
		return json.NewDecoder(io.LimitReader(resp.Body, responseLimit)).Decode(dest)
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, discardLimit))
	return nil
}

func retryAfter(resp *http.Response) time.Duration {
	var rate struct {
		RetryAfter float64 `json:"retry_after"`
	}
	_ = json.NewDecoder(io.LimitReader(resp.Body, discardLimit)).Decode(&rate)
	_ = resp.Body.Close()
	return max(time.Duration(rate.RetryAfter*float64(time.Second)), minRetryAfter)
}

func sleep(ctx context.Context, wait time.Duration) error {
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
