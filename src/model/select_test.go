package model

import (
	"encoding/json"
	"testing"
)

const roleSelectPick = `{
	"id": "1", "application_id": "2", "type": 3, "token": "t", "guild_id": "10", "channel_id": "20",
	"data": {
		"component_type": 6, "custom_id": "roles:pick", "values": ["31", "32"],
		"resolved": {"roles": {"31": {"id": "31", "name": "Red"}, "32": {"id": "32", "name": "Blue"}}}
	},
	"member": {"user": {"id": "40", "username": "npm"}},
	"message": {
		"id": "50", "channel_id": "20", "author": {"id": "2", "bot": true},
		"components": [
			{"type": 1, "components": [{"type": 6, "custom_id": "roles:pick", "placeholder": "Pick roles", "min_values": 0, "max_values": 2}]},
			{"type": 1, "components": [{"type": 3, "custom_id": "color", "options": [{"label": "Red", "value": "r", "default": true}]}]},
			{"type": 1, "components": [{"type": 2, "style": 1, "label": "OK", "custom_id": "ok"}]}
		]
	}
}`

func TestSelectInteractionDecodes(t *testing.T) {
	var i Interaction
	if err := json.Unmarshal([]byte(roleSelectPick), &i); err != nil {
		t.Fatal(err)
	}
	if !i.IsSelectMenu() || i.IsButton() || i.CustomID() != "roles:pick" {
		t.Fatalf("select = %v %v %q", i.IsSelectMenu(), i.IsButton(), i.CustomID())
	}
	if values := i.SelectedValues(); len(values) != 2 || values[0] != "31" {
		t.Fatalf("values = %v", values)
	}
	data, _ := i.ComponentData()
	if roles := data.Roles(); len(roles) != 2 || roles[1].Name != "Blue" {
		t.Fatalf("roles = %+v", roles)
	}
	if data.Users() != nil || data.Members() != nil || data.Channels() != nil {
		t.Fatal("resolved returned entities that were not sent")
	}
	rows := i.Message.Components
	if len(rows) != 3 || rows[0].Select == nil || len(rows[0].Components) != 0 {
		t.Fatalf("rows = %+v", rows)
	}
	if menu := rows[0].Select; menu.Type != ComponentTypeRoleSelect || menu.MinValues == nil || *menu.MinValues != 0 || *menu.MaxValues != 2 {
		t.Fatalf("role select = %+v", menu)
	}
	if menu := rows[1].Select; menu == nil || len(menu.Options) != 1 || !menu.Options[0].Default {
		t.Fatalf("string select = %+v", menu)
	}
	if rows[2].Select != nil || rows[2].Components[0].CustomID != "ok" {
		t.Fatalf("button row = %+v", rows[2])
	}
}

func TestSelectMembersFillUser(t *testing.T) {
	data := ComponentData{Values: []string{"7", "8"}, Resolved: &Resolved{
		Users:   map[string]User{"7": {ID: "7"}, "8": {ID: "8"}},
		Members: map[string]Member{"7": {}},
	}}
	members := data.Members()
	if len(members) != 1 || members[0].User == nil || members[0].User.ID != "7" {
		t.Fatalf("members = %+v", members)
	}
	if users := data.Users(); len(users) != 2 {
		t.Fatalf("users = %+v", users)
	}
}

func TestSelectRowsMarshal(t *testing.T) {
	rows := []ActionRow{
		NewSelectRow(StringSelect("color",
			NewSelectOption("Red", "r").WithDescription("Warm").WithEmoji(Emoji{Name: "🔴"}).WithDefault(true),
			NewSelectOption("Blue", "b"),
		).WithPlaceholder("Color").WithValues(1, 2)),
		NewSelectRow(ChannelSelect("log", ChannelTypeGuildText).WithDefaults(DefaultChannel("9")).WithDisabled(true)),
	}
	raw, err := json.Marshal(rows)
	if err != nil {
		t.Fatal(err)
	}
	want := `[{"type":1,"components":[{"type":3,"custom_id":"color","options":[{"label":"Red","value":"r","description":"Warm","emoji":{"name":"🔴"},"default":true},{"label":"Blue","value":"b"}],"placeholder":"Color","min_values":1,"max_values":2}]},` +
		`{"type":1,"components":[{"type":8,"custom_id":"log","channel_types":[0],"default_values":[{"id":"9","type":"channel"}],"disabled":true}]}]`
	if string(raw) != want {
		t.Fatalf("json = %s", raw)
	}
	var back []ActionRow
	if err := json.Unmarshal(raw, &back); err != nil || back[1].Select == nil || back[1].Select.ChannelTypes[0] != ChannelTypeGuildText {
		t.Fatalf("round trip = %+v %v", back, err)
	}
}

func TestSelectConstructors(t *testing.T) {
	for want, menu := range map[int]SelectMenu{
		ComponentTypeUserSelect:        UserSelect("u"),
		ComponentTypeRoleSelect:        RoleSelect("r"),
		ComponentTypeMentionableSelect: MentionableSelect("m"),
	} {
		if menu.Type != want || menu.CustomID == "" || !IsSelectType(menu.Type) {
			t.Fatalf("select %d = %+v", want, menu)
		}
	}
	if IsSelectType(ComponentTypeButton) || IsSelectType(ComponentTypeTextInput) {
		t.Fatal("non-select type reported as select")
	}
	if DefaultUser("1").Type != SelectDefaultUser || DefaultRole("1").Type != SelectDefaultRole {
		t.Fatal("default value types")
	}
}

func TestDisableComponents(t *testing.T) {
	rows := []ActionRow{NewActionRow(PrimaryButton("a", "A")), NewSelectRow(UserSelect("u"))}
	out := DisableComponents(rows)
	if !out[0].Components[0].Disabled || !out[1].Select.Disabled {
		t.Fatalf("disabled = %+v", out)
	}
	if rows[1].Select.Disabled || rows[0].Components[0].Disabled {
		t.Fatal("DisableComponents changed its input")
	}
	if kept := DisableButtons(rows); kept[1].Select == nil || kept[1].Select.Disabled {
		t.Fatalf("DisableButtons dropped or disabled the select: %+v", kept[1])
	}
}
