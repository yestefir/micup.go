package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
	"time"

	"github.com/yestefir/micup.go/src/internal/version"
	"github.com/yestefir/micup.go/src/model"
)

const maxAuditReason = 512

var (
	errAroundIteration = errors.New("iteration takes Before or After, not Around")
	errAfterIteration  = errors.New("audit log iteration runs newest first and takes Before, not After")
)

type RequestOption func(*request)

func WithReason(reason string) RequestOption {
	return func(r *request) { r.reason = reason }
}

type File = model.File

type request struct {
	method     string
	path       string
	query      url.Values
	body       any
	files      []File
	reason     string
	dest       any
	idempotent bool
	anonymous  bool
}

type payload struct {
	data        []byte
	contentType string
}

func (c *Client) call(ctx context.Context, r *request) error {
	if !strings.HasPrefix(r.path, "/") || strings.HasPrefix(r.path, "//") {
		return errors.New("REST path must be absolute to the API")
	}
	body, err := r.encode()
	if err != nil {
		return err
	}
	route := routeKey(r.method, r.path)
	retryable := r.idempotent || r.method == http.MethodGet || r.method == http.MethodHead || r.method == http.MethodPut || r.method == http.MethodDelete
	var last error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if err = c.limits.wait(ctx, route); err != nil {
			return err
		}
		resp, err := c.send(ctx, r, body)
		if err != nil {
			if ctx.Err() != nil || !retryable || attempt+1 == maxAttempts {
				return err
			}
			last = err
			if err = sleep(ctx, serverBackoff<<attempt); err != nil {
				return err
			}
			continue
		}
		c.limits.observe(route, resp.Header)
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return decode(resp, r.dest)
		}
		apiErr := readAPIError(resp)
		last = apiErr
		var wait time.Duration
		switch {
		case resp.StatusCode == http.StatusTooManyRequests:
			wait = max(apiErr.RetryAfter, minRetryAfter)
			c.limits.block(route, apiErr.Global, wait)
		case resp.StatusCode == http.StatusLocked:
			wait = max(apiErr.RetryAfter, time.Second)
		case resp.StatusCode >= 500 && retryable:
			wait = max(apiErr.RetryAfter, serverBackoff<<attempt)
		default:
			return apiErr
		}
		if wait > maxRetryAfter || attempt+1 == maxAttempts {
			return apiErr
		}
		if err = sleep(ctx, wait); err != nil {
			return err
		}
	}
	return last
}

func (c *Client) send(ctx context.Context, r *request, body payload) (*http.Response, error) {
	target := c.baseURL + r.path
	if len(r.query) != 0 {
		target += "?" + r.query.Encode()
	}
	var reader io.Reader
	if body.data != nil {
		reader = bytes.NewReader(body.data)
	}
	req, err := http.NewRequestWithContext(ctx, r.method, target, reader)
	if err != nil {
		return nil, err
	}
	if !r.anonymous {
		req.Header.Set("Authorization", "Bot "+c.token)
	}
	req.Header.Set("User-Agent", version.UserAgent)
	if body.contentType != "" {
		req.Header.Set("Content-Type", body.contentType)
	}
	if reason := strings.TrimSpace(r.reason); reason != "" && len(reason) <= maxAuditReason && !strings.ContainsAny(reason, "\r\n") {
		req.Header.Set("X-Audit-Log-Reason", reason)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, errors.New("Micup API connection failed")
	}
	return resp, nil
}

func (r *request) encode() (payload, error) {
	if len(r.files) == 0 {
		if r.body == nil {
			return payload{}, nil
		}
		data, err := json.Marshal(r.body)
		return payload{data: data, contentType: "application/json"}, err
	}
	var buf bytes.Buffer
	form := multipart.NewWriter(&buf)
	if r.body != nil {
		data, err := json.Marshal(r.body)
		if err != nil {
			return payload{}, err
		}
		if err = form.WriteField("payload_json", string(data)); err != nil {
			return payload{}, err
		}
	}
	for i, file := range r.files {
		if file.Name == "" || file.Reader == nil {
			return payload{}, errors.New("a file needs a name and a reader")
		}
		header := textproto.MIMEHeader{}
		header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="files[%d]"; filename=%q`, i, file.Name))
		header.Set("Content-Type", contentTypeOr(file.ContentType))
		part, err := form.CreatePart(header)
		if err != nil {
			return payload{}, err
		}
		if _, err = io.Copy(part, file.Reader); err != nil {
			return payload{}, err
		}
	}
	if err := form.Close(); err != nil {
		return payload{}, err
	}
	return payload{data: buf.Bytes(), contentType: form.FormDataContentType()}, nil
}

func contentTypeOr(contentType string) string {
	if contentType == "" {
		return "application/octet-stream"
	}
	return contentType
}

func decode(resp *http.Response, dest any) error {
	defer resp.Body.Close()
	if dest != nil && resp.StatusCode != http.StatusNoContent {
		return json.NewDecoder(io.LimitReader(resp.Body, responseLimit)).Decode(dest)
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, discardLimit))
	return nil
}
