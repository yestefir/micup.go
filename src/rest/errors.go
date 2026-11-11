package rest

import (
	"encoding/json"
	"errors"
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
	if e.Message != "" {
		text += ": " + e.Message
	}
	for _, field := range e.Errors {
		text += "; " + field.Path + ": " + field.Message
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

const (
	CodeInvalidFormBody       = "INVALID_FORM_BODY"
	CodeRateLimited           = "RATE_LIMITED"
	CodeSlowmodeRateLimited   = "SLOWMODE_RATE_LIMITED"
	CodeResourceLocked        = "RESOURCE_LOCKED"
	CodeUnauthorized          = "UNAUTHORIZED"
	CodeMissingAccess         = "MISSING_ACCESS"
	CodeMissingPermissions    = "MISSING_PERMISSIONS"
	CodeTwoFactorRequired     = "TWO_FACTOR_REQUIRED"
	CodeCommunicationDisabled = "COMMUNICATION_DISABLED"
	CodeContentBlocked        = "CONTENT_BLOCKED"
	CodeCannotSendEmpty       = "CANNOT_SEND_EMPTY_MESSAGE"
	CodeUnknownChannel        = "UNKNOWN_CHANNEL"
	CodeUnknownGuild          = "UNKNOWN_GUILD"
	CodeUnknownMember         = "UNKNOWN_MEMBER"
	CodeUnknownMessage        = "UNKNOWN_MESSAGE"
	CodeUnknownRole           = "UNKNOWN_ROLE"
	CodeUnknownUser           = "UNKNOWN_USER"
	CodeUnknownInvite         = "UNKNOWN_INVITE"
	CodeUnknownEmoji          = "UNKNOWN_EMOJI"
	CodeUnknownSticker        = "UNKNOWN_STICKER"
	CodeUnknownWebhook        = "UNKNOWN_WEBHOOK"
	CodeThreadArchived        = "THREAD_ARCHIVED"
	CodeThreadLocked          = "THREAD_LOCKED"
	CodeMaxActiveThreads      = "MAX_ACTIVE_THREADS"
	CodeNSFWAgeRestricted     = "NSFW_CONTENT_AGE_RESTRICTED"
	CodeFeatureDisabled       = "FEATURE_TEMPORARILY_DISABLED"
)

func (e *APIError) HasFieldCode(code string) bool {
	for _, field := range e.Errors {
		if field.Code == code {
			return true
		}
	}
	return false
}

func (e *APIError) NotFound() bool { return e.StatusCode == http.StatusNotFound }

func (e *APIError) Forbidden() bool { return e.StatusCode == http.StatusForbidden }

func (e *APIError) RateLimited() bool { return e.StatusCode == http.StatusTooManyRequests }

func ErrorCode(err error) string {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code
	}
	return ""
}

func IsErrorCode(err error, codes ...string) bool {
	code := ErrorCode(err)
	if code == "" {
		return false
	}
	for _, want := range codes {
		if code == want {
			return true
		}
	}
	return false
}

func IsNotFound(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.NotFound()
}

func IsForbidden(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.Forbidden()
}
