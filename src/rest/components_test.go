package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/yestefir/micup.go/src/model"
)

func click() model.Interaction {
	return model.Interaction{ID: "9", ApplicationID: "app", Token: "tok/en", Type: model.InteractionTypeMessageComponent, Data: json.RawMessage(`{"component_type":2,"custom_id":"a"}`)}
}

func componentsJSON(t *testing.T, v any) string {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func TestSendMessageWithButtons(t *testing.T) {
	c, rec := recordingClient(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"id":"5","channel_id":"1","author":{"id":"b"},"components":[{"type":1,"components":[{"type":2,"style":1,"label":"A","custom_id":"a"}]}]}`)
	})
	msg, err := c.SendMessageWithOptions(context.Background(), "1", model.MessageOptions{Components: []model.ActionRow{model.NewActionRow(model.PrimaryButton("a", "A"))}})
	if err != nil {
		t.Fatal(err)
	}
	got := rec.last()
	if got.method != http.MethodPost || got.path != "/channels/1/messages" {
		t.Fatalf("request = %s %s", got.method, got.path)
	}
	if s := componentsJSON(t, got.body["components"]); s != `[{"components":[{"custom_id":"a","label":"A","style":1,"type":2}],"type":1}]` {
		t.Fatalf("components = %s", s)
	}
	if len(msg.Components) != 1 || msg.Components[0].Components[0].CustomID != "a" {
		t.Fatalf("message components = %+v", msg.Components)
	}
}

func TestEditMessageButtons(t *testing.T) {
	c, rec := recordingClient(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"id":"5","channel_id":"1","author":{"id":"b"}}`)
	})
	rows := []model.ActionRow{model.NewActionRow(model.DangerButton("stop", "Stop"))}
	if _, err := c.EditMessage(context.Background(), "1", "5", model.MessageEdit{Components: &rows}); err != nil {
		t.Fatal(err)
	}
	if got := rec.last(); got.method != http.MethodPatch || !strings.Contains(string(got.raw), `"custom_id":"stop"`) {
		t.Fatalf("edit = %s %s", got.method, got.raw)
	}
	var none []model.ActionRow
	if _, err := c.EditMessage(context.Background(), "1", "5", model.MessageEdit{Components: &none}); err != nil {
		t.Fatal(err)
	}
	if got := rec.last(); !strings.Contains(string(got.raw), `"components":[]`) {
		t.Fatalf("clearing buttons sent %s", got.raw)
	}
}

func TestInvalidButtonsAreRejected(t *testing.T) {
	c, rec := recordingClient(t, nil)
	row := func(b ...model.Button) []model.ActionRow { return []model.ActionRow{model.NewActionRow(b...)} }
	six := make([]model.ActionRow, model.MaxActionRows+1)
	for i := range six {
		six[i] = model.NewActionRow(model.PrimaryButton(fmt.Sprint(i), "x"))
	}
	cases := map[string][]model.ActionRow{
		"too many rows":     six,
		"empty row":         {{}},
		"too many buttons":  row(model.PrimaryButton("1", "x"), model.PrimaryButton("2", "x"), model.PrimaryButton("3", "x"), model.PrimaryButton("4", "x"), model.PrimaryButton("5", "x"), model.PrimaryButton("6", "x")),
		"no custom ID":      row(model.PrimaryButton("", "x")),
		"long custom ID":    row(model.PrimaryButton(strings.Repeat("a", model.MaxCustomIDLength+1), "x")),
		"duplicate ID":      {model.NewActionRow(model.PrimaryButton("a", "x")), model.NewActionRow(model.SecondaryButton("a", "y"))},
		"no label or emoji": row(model.PrimaryButton("a", "")),
		"long label":        row(model.PrimaryButton("a", strings.Repeat("é", model.MaxButtonLabelLength+1))),
		"link without URL":  row(model.LinkButton("", "x")),
		"link with ID":      row(model.Button{Style: model.ButtonStyleLink, URL: "https://example.com", CustomID: "a", Label: "x"}),
		"URL on non-link":   row(model.Button{Style: model.ButtonStylePrimary, URL: "https://example.com", CustomID: "a", Label: "x"}),
		"unknown style":     row(model.Button{Style: 9, CustomID: "a", Label: "x"}),
		"not a button":      row(model.Button{Type: 3, Style: 1, CustomID: "a", Label: "x"}),
		"not an action row": {{Type: 2, Components: []model.Button{model.PrimaryButton("a", "x")}}},
	}
	for name, rows := range cases {
		if _, err := c.SendMessageWithOptions(context.Background(), "1", model.MessageOptions{Components: rows}); err == nil {
			t.Errorf("%s: accepted", name)
		}
		if _, err := c.EditMessage(context.Background(), "1", "5", model.MessageEdit{Components: &rows}); err == nil {
			t.Errorf("%s: edit accepted", name)
		}
	}
	if _, err := c.ExecuteWebhook(context.Background(), "w", "t", model.WebhookMessage{MessageOptions: model.MessageOptions{Content: "x", Components: row(model.PrimaryButton("a", "x"))}}); err == nil {
		t.Error("webhook with buttons accepted")
	}
	if len(rec.requests) != 0 {
		t.Fatalf("invalid buttons reached Micup: %d requests", len(rec.requests))
	}
	emojiOnly := row(model.SecondaryButton("a", "").WithEmoji(model.Emoji{Name: "⏭️"}))
	if _, err := c.SendMessageWithOptions(context.Background(), "1", model.MessageOptions{Components: emojiOnly}); err != nil {
		t.Fatalf("emoji-only button rejected: %v", err)
	}
}

func TestInteractionUpdateCallbacks(t *testing.T) {
	c, rec := recordingClient(t, nil)
	ctx := context.Background()
	content := "Paused"
	rows := model.DisableButtons([]model.ActionRow{model.NewActionRow(model.PrimaryButton("a", "A"))})
	if err := c.UpdateInteractionMessage(ctx, click(), model.MessageEdit{Content: &content, Components: &rows, AllowedMentions: &model.AllowedMentions{}}); err != nil {
		t.Fatal(err)
	}
	got := rec.last()
	if got.method != http.MethodPost || got.path != "/interactions/9/tok%2Fen/callback" {
		t.Fatalf("update = %s %s", got.method, got.path)
	}
	if got.body["type"] != float64(7) || !strings.Contains(string(got.raw), `"content":"Paused"`) || !strings.Contains(string(got.raw), `"disabled":true`) {
		t.Fatalf("update body = %s", got.raw)
	}
	data := got.body["data"].(map[string]any)
	mentions := data["allowed_mentions"].(map[string]any)
	if parse, ok := mentions["parse"].([]any); !ok || len(parse) != 0 {
		t.Fatalf("update mention policy = %s, want an explicit empty parse array", got.raw)
	}
	if err := c.DeferInteractionUpdate(ctx, click()); err != nil {
		t.Fatal(err)
	}
	if got = rec.last(); string(got.raw) != `{"type":6}` {
		t.Fatalf("defer update body = %s", got.raw)
	}
	if err := c.RespondToInteraction(ctx, click(), model.MessageOptions{Content: "Pick", Components: rows}, true); err != nil {
		t.Fatal(err)
	}
	if got = rec.last(); got.body["type"] != float64(4) || !strings.Contains(string(got.raw), `"custom_id":"a"`) || !strings.Contains(string(got.raw), `"flags":64`) {
		t.Fatalf("respond body = %s", got.raw)
	}
}

func TestInteractionUpdateRejectsBadInput(t *testing.T) {
	c, rec := recordingClient(t, nil)
	ctx := context.Background()
	content := "x"
	command := click()
	command.Type = model.InteractionTypeApplicationCommand
	if err := c.UpdateInteractionMessage(ctx, command, model.MessageEdit{Content: &content}); err == nil {
		t.Error("command interaction updated a message")
	}
	if err := c.DeferInteractionUpdate(ctx, model.Interaction{Type: model.InteractionTypeMessageComponent}); err == nil {
		t.Error("interaction without ID and token accepted")
	}
	if err := c.UpdateInteractionMessage(ctx, click(), model.MessageEdit{}); err == nil {
		t.Error("empty update accepted")
	}
	if err := c.UpdateInteractionMessage(ctx, click(), model.MessageEdit{Files: []model.File{{Name: "a"}}}); err == nil {
		t.Error("update with files accepted")
	}
	if len(rec.requests) != 0 {
		t.Fatalf("bad updates reached Micup: %d requests", len(rec.requests))
	}
}

func TestEditInteractionResponse(t *testing.T) {
	c, rec := recordingClient(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"id":"50","channel_id":"20","author":{"id":"app"},"content":"Done"}`)
	})
	content := "Done"
	var none []model.ActionRow
	msg, err := c.EditInteractionResponse(context.Background(), click(), model.MessageEdit{Content: &content, Components: &none})
	if err != nil {
		t.Fatal(err)
	}
	got := rec.last()
	if got.method != http.MethodPatch || got.path != "/webhooks/app/tok%2Fen/messages/@original" {
		t.Fatalf("edit response = %s %s", got.method, got.path)
	}
	if !strings.Contains(string(got.raw), `"components":[]`) || msg.Content != "Done" {
		t.Fatalf("edit response body = %s, message %+v", got.raw, msg)
	}
}

func TestSendMessageWithSelectMenus(t *testing.T) {
	c, rec := recordingClient(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"id":"5","channel_id":"1","author":{"id":"b"},"components":[{"type":1,"components":[{"type":6,"custom_id":"roles"}]}]}`)
	})
	rows := []model.ActionRow{
		model.NewSelectRow(model.StringSelect("color", model.NewSelectOption("Red", "r"), model.NewSelectOption("Blue", "b")).WithValues(1, 2)),
		model.NewSelectRow(model.RoleSelect("roles")),
	}
	msg, err := c.SendMessageWithOptions(context.Background(), "1", model.MessageOptions{Components: rows})
	if err != nil {
		t.Fatal(err)
	}
	want := `[{"components":[{"custom_id":"color","max_values":2,"min_values":1,"options":[{"label":"Red","value":"r"},{"label":"Blue","value":"b"}],"type":3}],"type":1},{"components":[{"custom_id":"roles","type":6}],"type":1}]`
	if s := componentsJSON(t, rec.last().body["components"]); s != want {
		t.Fatalf("components = %s", s)
	}
	if len(msg.Components) != 1 || msg.Components[0].Select == nil || msg.Components[0].Select.CustomID != "roles" {
		t.Fatalf("message components = %+v", msg.Components)
	}
}

func TestInvalidSelectMenusAreRejected(t *testing.T) {
	c, rec := recordingClient(t, nil)
	sel := func(m model.SelectMenu) []model.ActionRow { return []model.ActionRow{model.NewSelectRow(m)} }
	opt := model.NewSelectOption("A", "a")
	many := make([]model.SelectOption, model.MaxSelectOptions+1)
	for i := range many {
		many[i] = model.NewSelectOption("x", fmt.Sprint(i))
	}
	mixed := model.NewSelectRow(model.UserSelect("u"))
	mixed.Components = []model.Button{model.PrimaryButton("b", "B")}
	cases := map[string][]model.ActionRow{
		"buttons and select":     {mixed},
		"no custom ID":           sel(model.UserSelect("")),
		"long custom ID":         sel(model.UserSelect(strings.Repeat("a", model.MaxCustomIDLength+1))),
		"duplicate ID":           {model.NewActionRow(model.PrimaryButton("a", "A")), model.NewSelectRow(model.UserSelect("a"))},
		"unknown type":           sel(model.SelectMenu{Type: 4, CustomID: "a"}),
		"long placeholder":       sel(model.UserSelect("a").WithPlaceholder(strings.Repeat("é", model.MaxSelectPlaceholder+1))),
		"min above max":          sel(model.UserSelect("a").WithValues(3, 2)),
		"max zero":               sel(model.UserSelect("a").WithValues(0, 0)),
		"max over limit":         sel(model.UserSelect("a").WithValues(1, model.MaxSelectOptions+1)),
		"string without options": sel(model.StringSelect("a")),
		"too many options":       sel(model.StringSelect("a", many...)),
		"max over options":       sel(model.StringSelect("a", opt).WithValues(1, 2)),
		"option without label":   sel(model.StringSelect("a", model.NewSelectOption("", "a"))),
		"option without value":   sel(model.StringSelect("a", model.NewSelectOption("A", ""))),
		"long description":       sel(model.StringSelect("a", opt.WithDescription(strings.Repeat("d", model.MaxSelectOptionDescription+1)))),
		"duplicate value":        sel(model.StringSelect("a", opt, model.NewSelectOption("B", "a"))),
		"too many defaults":      sel(model.StringSelect("a", opt.WithDefault(true), model.NewSelectOption("B", "b").WithDefault(true))),
		"string default_values":  sel(model.StringSelect("a", opt).WithDefaults(model.DefaultUser("1"))),
		"options on user select": sel(model.SelectMenu{Type: model.ComponentTypeUserSelect, CustomID: "a", Options: []model.SelectOption{opt}}),
		"channel types on role":  sel(model.SelectMenu{Type: model.ComponentTypeRoleSelect, CustomID: "a", ChannelTypes: []int{0}}),
		"wrong default type":     sel(model.UserSelect("a").WithDefaults(model.DefaultRole("1"))),
		"default without ID":     sel(model.ChannelSelect("a").WithDefaults(model.DefaultChannel(""))),
		"defaults over max":      sel(model.RoleSelect("a").WithDefaults(model.DefaultRole("1"), model.DefaultRole("2"))),
		"channel on mentionable": sel(model.MentionableSelect("a").WithDefaults(model.DefaultChannel("1"))),
	}
	for name, rows := range cases {
		if _, err := c.SendMessageWithOptions(context.Background(), "1", model.MessageOptions{Components: rows}); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
	if len(rec.requests) != 0 {
		t.Fatalf("invalid select menus reached Micup: %d requests", len(rec.requests))
	}
	valid := []model.ActionRow{
		model.NewSelectRow(model.MentionableSelect("m").WithValues(0, 2).WithDefaults(model.DefaultUser("1"), model.DefaultRole("2"))),
		model.NewSelectRow(model.ChannelSelect("c", model.ChannelTypeGuildText).WithDefaults(model.DefaultChannel("3"))),
		model.NewSelectRow(model.StringSelect("s", opt.WithDefault(true))),
	}
	if _, err := c.SendMessageWithOptions(context.Background(), "1", model.MessageOptions{Components: valid}); err != nil {
		t.Fatalf("valid select menus rejected: %v", err)
	}
}
