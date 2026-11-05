package micup

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	restMaxAttempts   = 4
	restMaxRetryAfter = 30 * time.Second
	restResponseLimit = 2 << 20
	restDiscardLimit  = 8192
)

type APIError struct{ StatusCode int }

func (e *APIError) Error() string { return fmt.Sprintf("Micup API returned HTTP %d", e.StatusCode) }

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
	for attempt := 0; attempt < restMaxAttempts; attempt++ {
		req, requestErr := http.NewRequestWithContext(ctx, method, c.apiURL+path, bytes.NewReader(raw))
		if requestErr != nil {
			return requestErr
		}
		req.Header.Set("Authorization", "Bot "+c.config.Token)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", defaultUserAgent)
		resp, requestErr := c.config.HTTP.Do(req)
		if requestErr != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return errors.New("Micup API connection failed")
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			wait := retryAfter(resp)
			if wait > restMaxRetryAfter {
				return &APIError{StatusCode: http.StatusTooManyRequests}
			}
			if err = sleep(ctx, wait); err != nil {
				return err
			}
			continue
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			_ = resp.Body.Close()
			return &APIError{StatusCode: resp.StatusCode}
		}
		if dest != nil && resp.StatusCode != http.StatusNoContent {
			err = json.NewDecoder(io.LimitReader(resp.Body, restResponseLimit)).Decode(dest)
		} else {
			_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, restDiscardLimit))
		}
		_ = resp.Body.Close()
		return err
	}
	return &APIError{StatusCode: http.StatusTooManyRequests}
}

func retryAfter(resp *http.Response) time.Duration {
	var rate struct {
		RetryAfter float64 `json:"retry_after"`
	}
	_ = json.NewDecoder(io.LimitReader(resp.Body, restDiscardLimit)).Decode(&rate)
	_ = resp.Body.Close()
	return max(time.Duration(rate.RetryAfter*float64(time.Second)), time.Second)
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
