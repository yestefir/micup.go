package model

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCommandValidate(t *testing.T) {
	ok := []ApplicationCommand{
		{Name: "ping", Description: "Ping"},
		{Name: "çal", Description: "Şarkı çal", Options: []CommandOption{{Type: OptionTypeString, Name: "şarkı", Description: "Ad", Required: true}, {Type: OptionTypeInteger, Name: "ses", Description: "Seviye"}}},
		{Name: "admin", Description: "Admin", Options: []CommandOption{{Type: OptionTypeSubcommandGroup, Name: "role", Description: "Roles", Options: []CommandOption{{Type: OptionTypeSubcommand, Name: "add", Description: "Add"}}}}},
		{Type: CommandTypeUser, Name: "Inspect user"},
		{Type: CommandTypeMessage, Name: "Quote message"},
	}
	for _, c := range ok {
		if err := c.Validate(); err != nil {
			t.Errorf("%s: %v", c.Name, err)
		}
	}
	bad := map[string]ApplicationCommand{
		"uppercase":          {Name: "Ping", Description: "Ping"},
		"space":              {Name: "two words", Description: "x"},
		"long name":          {Name: strings.Repeat("a", 33), Description: "x"},
		"no description":     {Name: "ping"},
		"long description":   {Name: "ping", Description: strings.Repeat("a", 101)},
		"context options":    {Type: CommandTypeUser, Name: "Inspect", Options: []CommandOption{{Type: OptionTypeString, Name: "x", Description: "x"}}},
		"context desc":       {Type: CommandTypeMessage, Name: "Quote", Description: "x"},
		"unknown type":       {Type: 9, Name: "x"},
		"required after":     {Name: "x", Description: "x", Options: []CommandOption{{Type: OptionTypeString, Name: "a", Description: "a"}, {Type: OptionTypeString, Name: "b", Description: "b", Required: true}}},
		"duplicate":          {Name: "x", Description: "x", Options: []CommandOption{{Type: OptionTypeString, Name: "a", Description: "a"}, {Type: OptionTypeString, Name: "a", Description: "a"}}},
		"choices and auto":   {Name: "x", Description: "x", Options: []CommandOption{{Type: OptionTypeString, Name: "a", Description: "a", Autocomplete: true, Choices: []CommandChoice{{Name: "n", Value: "v"}}}}},
		"group in group":     {Name: "x", Description: "x", Options: []CommandOption{{Type: OptionTypeSubcommandGroup, Name: "a", Description: "a", Options: []CommandOption{{Type: OptionTypeSubcommandGroup, Name: "b", Description: "b"}}}}},
		"unknown option":     {Name: "x", Description: "x", Options: []CommandOption{{Type: 99, Name: "a", Description: "a"}}},
		"too many options":   {Name: "x", Description: "x", Options: make([]CommandOption, MaxCommandOptions+1)},
		"option description": {Name: "x", Description: "x", Options: []CommandOption{{Type: OptionTypeString, Name: "a"}}},
	}
	for name, c := range bad {
		if err := c.Validate(); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
}

func TestCommandDataReadsOptions(t *testing.T) {
	raw := `{"id":"c","name":"admin","type":1,"options":[{"name":"role","type":2,"options":[{"name":"add","type":1,"options":[
		{"name":"role","type":8,"value":"r"},
		{"name":"user","type":6,"value":"u"},
		{"name":"count","type":4,"value":3},
		{"name":"ratio","type":10,"value":1.5},
		{"name":"note","type":3,"value":"hi","focused":true},
		{"name":"loud","type":5,"value":true},
		{"name":"where","type":7,"value":"ch"}]}]}],
		"resolved":{"roles":{"r":{"id":"r","name":"Mod","permissions":"8"}},"users":{"u":{"id":"u","username":"ada"}},"members":{"u":{"roles":["r"]}},"channels":{"ch":{"id":"ch","name":"genel","permissions":"8"}}}}`
	i := Interaction{Type: InteractionTypeApplicationCommand, Data: json.RawMessage(raw)}
	data, ok := i.CommandData()
	if !ok || !i.IsCommand() || i.IsAutocomplete() {
		t.Fatal("command data was not read")
	}
	if got := strings.Join(data.Path(), " "); got != "admin role add" {
		t.Fatalf("path %q", got)
	}
	if role, ok := data.Role("role"); !ok || role.Name != "Mod" || !role.Permissions.Has(PermissionAdministrator) {
		t.Fatalf("role: %+v", role)
	}
	if user, ok := data.User("user"); !ok || user.Username != "ada" {
		t.Fatalf("user: %+v", user)
	}
	if member, ok := data.Member("user"); !ok || member.User == nil || member.User.ID != "u" || member.Roles[0] != "r" {
		t.Fatalf("member: %+v", member)
	}
	if n, ok := data.Int("count"); !ok || n != 3 {
		t.Fatalf("int: %d", n)
	}
	if f, ok := data.Float("ratio"); !ok || f != 1.5 {
		t.Fatalf("float: %v", f)
	}
	if s, ok := data.String("note"); !ok || s != "hi" {
		t.Fatalf("string: %q", s)
	}
	if b, ok := data.Bool("loud"); !ok || !b {
		t.Fatal("bool")
	}
	if c, ok := data.Channel("where"); !ok || c.Name != "genel" || c.Permissions == nil {
		t.Fatalf("channel: %+v", c)
	}
	if f, ok := data.Focused(); !ok || f.Name != "note" {
		t.Fatalf("focused: %+v", f)
	}
	if _, ok := data.String("missing"); ok {
		t.Fatal("missing option found")
	}
	if _, ok := (Interaction{Type: InteractionTypeMessageComponent, Data: json.RawMessage(raw)}).CommandData(); ok {
		t.Fatal("component read as command")
	}
}

func TestCommandDataTargets(t *testing.T) {
	user := Interaction{Type: InteractionTypeApplicationCommand, Data: json.RawMessage(`{"name":"Inspect user","type":2,"target_id":"u","resolved":{"users":{"u":{"id":"u","username":"ada"}}}}`)}
	data, _ := user.CommandData()
	if target, ok := data.TargetUser(); !ok || target.Username != "ada" {
		t.Fatalf("target user: %+v", target)
	}
	if _, ok := data.TargetMessage(); ok {
		t.Fatal("user command returned a message")
	}
	message := Interaction{Type: InteractionTypeApplicationCommand, Data: json.RawMessage(`{"name":"Quote","type":3,"target_id":"m","resolved":{"messages":{"m":{"id":"m","content":"hi"}}}}`)}
	data, _ = message.CommandData()
	if target, ok := data.TargetMessage(); !ok || target.Content != "hi" {
		t.Fatalf("target message: %+v", target)
	}
}
