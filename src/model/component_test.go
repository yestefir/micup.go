package model

import (
	"encoding/json"
	"testing"
)

const buttonClick = `{
	"id": "1567726025931489280",
	"application_id": "1567701032728268800",
	"type": 3,
	"token": "interaction-token",
	"version": 1,
	"guild_id": "10",
	"channel_id": "20",
	"app_permissions": "18446744073709551615",
	"locale": "tr",
	"guild_locale": "en-US",
	"data": {"component_type": 2, "custom_id": "player:pause"},
	"member": {"roles": ["30"], "user": {"id": "40", "username": "npm"}},
	"message": {
		"id": "50", "channel_id": "20", "content": "Now playing", "author": {"id": "1567701032728268800", "bot": true},
		"components": [{"type": 1, "components": [
			{"type": 2, "style": 1, "label": "Pause", "custom_id": "player:pause", "disabled": false},
			{"type": 2, "style": 5, "label": "Link", "url": "https://example.com"}
		]}]
	}
}`

func TestButtonInteractionDecodes(t *testing.T) {
	var i Interaction
	if err := json.Unmarshal([]byte(buttonClick), &i); err != nil {
		t.Fatal(err)
	}
	if !i.IsButton() || i.CustomID() != "player:pause" {
		t.Fatalf("button click = %v %q", i.IsButton(), i.CustomID())
	}
	if u := i.Invoker(); u == nil || u.ID != "40" {
		t.Fatalf("invoker = %+v", u)
	}
	if i.AppPermissions == nil || !i.AppPermissions.Has(PermissionAdministrator) || i.Locale != "tr" {
		t.Fatalf("permissions or locale = %v %q", i.AppPermissions, i.Locale)
	}
	if i.Message == nil || len(i.Message.Components) != 1 || len(i.Message.Components[0].Components) != 2 {
		t.Fatalf("message components = %+v", i.Message)
	}
	link := i.Message.Components[0].Components[1]
	if link.Type != ComponentTypeButton || link.Style != ButtonStyleLink || link.URL != "https://example.com" {
		t.Fatalf("link button = %+v", link)
	}
}

func TestInteractionHelpersOnOtherTypes(t *testing.T) {
	dm := Interaction{Type: InteractionTypeMessageComponent, Data: json.RawMessage(`{"component_type":3,"custom_id":"menu"}`), User: &User{ID: "7"}}
	if dm.IsButton() || dm.CustomID() != "menu" || dm.Invoker().ID != "7" {
		t.Fatalf("select menu interaction = %v %q", dm.IsButton(), dm.CustomID())
	}
	command := Interaction{Type: InteractionTypeApplicationCommand, Data: json.RawMessage(`{"custom_id":"x"}`)}
	if _, ok := command.ComponentData(); ok || command.CustomID() != "" || command.Invoker() != nil {
		t.Fatal("command interaction read as a component")
	}
}

func TestComponentsMarshalWithTypes(t *testing.T) {
	rows := []ActionRow{{Components: []Button{{Style: ButtonStylePrimary, CustomID: "a", Label: "A"}}}, NewActionRow(LinkButton("https://example.com", "Docs").WithEmoji(Emoji{Name: "📘"}))}
	raw, err := json.Marshal(rows)
	if err != nil {
		t.Fatal(err)
	}
	want := `[{"type":1,"components":[{"type":2,"style":1,"label":"A","custom_id":"a"}]},{"type":1,"components":[{"type":2,"style":5,"label":"Docs","emoji":{"name":"📘"},"url":"https://example.com"}]}]`
	if string(raw) != want {
		t.Fatalf("json = %s", raw)
	}
	empty, _ := json.Marshal(ActionRow{})
	if string(empty) != `{"type":1,"components":[]}` {
		t.Fatalf("empty row = %s", empty)
	}
}

func TestButtonConstructors(t *testing.T) {
	for style, b := range map[int]Button{
		ButtonStylePrimary:   PrimaryButton("p", "P"),
		ButtonStyleSecondary: SecondaryButton("s", "S"),
		ButtonStyleSuccess:   SuccessButton("ok", "OK"),
		ButtonStyleDanger:    DangerButton("d", "D"),
	} {
		if b.Style != style || b.Type != ComponentTypeButton || b.CustomID == "" || b.URL != "" {
			t.Fatalf("style %d button = %+v", style, b)
		}
	}
	if b := PrimaryButton("p", "P").WithDisabled(true); !b.Disabled {
		t.Fatal("WithDisabled had no effect")
	}
}

func TestDisableButtonsKeepsLinksAndInput(t *testing.T) {
	rows := []ActionRow{NewActionRow(PrimaryButton("a", "A"), LinkButton("https://example.com", "L"))}
	out := DisableButtons(rows)
	if !out[0].Components[0].Disabled || out[0].Components[1].Disabled {
		t.Fatalf("disabled = %+v", out[0].Components)
	}
	if rows[0].Components[0].Disabled {
		t.Fatal("DisableButtons changed its input")
	}
}
