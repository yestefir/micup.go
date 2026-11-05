package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type FieldError struct {
	Path    string `json:"path"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type APIError struct {
	StatusCode int
	Code       string
	Message    string
	Errors     []FieldError
	RequestID  string
	RetryAfter time.Duration
	Global     bool
}

func (e *APIError) Error() string {
	text := fmt.Sprintf("Micup API returned HTTP %d", e.StatusCode)
	if e.Code != "" {
		text += " (" + e.Code + ")"
	}
	if e.RequestID != "" {
		text += " request " + e.RequestID
	}
	return text
}

func readAPIError(resp *http.Response) *APIError {
	defer resp.Body.Close()
	var body struct {
		Code       string       `json:"code"`
		Message    string       `json:"message"`
		Errors     []FieldError `json:"errors"`
		RetryAfter float64      `json:"retry_after"`
		Global     bool         `json:"global"`
	}
	_ = json.NewDecoder(io.LimitReader(resp.Body, errorBodyLimit)).Decode(&body)
	apiErr := &APIError{
		StatusCode: resp.StatusCode,
		Code:       body.Code,
		Message:    body.Message,
		Errors:     body.Errors,
		RequestID:  resp.Header.Get("X-Request-ID"),
		Global:     body.Global || resp.Header.Get("X-RateLimit-Global") == "true",
	}
	switch {
	case body.RetryAfter > 0:
		apiErr.RetryAfter = max(time.Duration(body.RetryAfter*float64(time.Second)), minRetryAfter)
	default:
		apiErr.RetryAfter = headerSeconds(resp.Header.Get("Retry-After"))
	}
	return apiErr
}

func headerSeconds(value string) time.Duration {
	seconds, err := strconv.ParseFloat(value, 64)
	if err != nil || seconds <= 0 {
		return 0
	}
	return time.Duration(seconds * float64(time.Second))
}
