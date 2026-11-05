package model

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestPresenceSendsNullCustomStatus(t *testing.T) {
	data, err := json.Marshal(Presence{Status: StatusIdle})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"custom_status":null`) {
		t.Fatalf("a cleared custom status must be sent as null: %s", data)
	}
	data, _ = json.Marshal(Presence{Status: StatusOnline, CustomStatus: &CustomStatus{Text: "!help"}})
	if !strings.Contains(string(data), `"custom_status":{"text":"!help"}`) {
		t.Fatalf("custom status was not sent: %s", data)
	}
}

func TestCustomStatusValidate(t *testing.T) {
	now := time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC)
	until := CustomStatus{Text: "x"}.Until(now.Add(time.Hour).In(time.FixedZone("TRT", 3*3600)))
	if until.ExpiresAt != "2026-11-05T13:00:00Z" {
		t.Fatalf("Until did not write UTC RFC 3339: %s", until.ExpiresAt)
	}
	valid := []CustomStatus{
		{Text: "!yardım"},
		{EmojiName: "🎵"},
		{EmojiID: "1"},
		{Text: strings.Repeat("ş", MaxCustomStatusText)},
		until,
	}
	for _, c := range valid {
		if err := c.Validate(now); err != nil {
			t.Errorf("%+v: %v", c, err)
		}
	}
	invalid := []CustomStatus{
		{},
		{Text: strings.Repeat("a", MaxCustomStatusText+1)},
		// One emoji above U+FFFF is two UTF-16 units, as Micup counts.
		{Text: strings.Repeat("🎵", MaxCustomStatusText/2+1)},
		{EmojiName: strings.Repeat("🎵", MaxCustomStatusEmoji)},
		{Text: "x", ExpiresAt: "tomorrow"},
		{Text: "x", ExpiresAt: now.Format(time.RFC3339)},
	}
	for _, c := range invalid {
		if err := c.Validate(now); err == nil {
			t.Errorf("%+v was accepted", c)
		}
	}
	if err := (Presence{Status: "busy"}).Validate(now); err == nil {
		t.Error("invalid status accepted")
	}
	if err := (Presence{Status: StatusDND, CustomStatus: &CustomStatus{}}).Validate(now); err == nil {
		t.Error("empty custom status accepted")
	}
}
