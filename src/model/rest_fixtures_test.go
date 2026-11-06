package model

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func decodeFixture(t *testing.T, name string, dest any) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "rest", name+".json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, dest); err != nil {
		t.Fatalf("%s: %v", name, err)
	}
}

func TestRESTFixturesDecode(t *testing.T) {
	var guild Guild
	decodeFixture(t, "get_guild", &guild)
	if guild.ID == "" || guild.OwnerID == "" || len(guild.Roles) == 0 || len(guild.Features) == 0 || guild.MemberCount == 0 {
		t.Errorf("guild fields missing: %+v", guild)
	}

	var roles []Role
	decodeFixture(t, "get_roles", &roles)
	for _, role := range roles {
		if role.ID == "" || role.Colors == nil {
			t.Errorf("role fields missing: %+v", role)
		}
	}
	var admin bool
	for _, role := range roles {
		admin = admin || role.Permissions.Has(PermissionAdministrator)
	}
	if !admin {
		t.Error("string permissions were not decoded")
	}

	for _, name := range []string{"create_role", "modify_role"} {
		var role Role
		decodeFixture(t, name, &role)
		if role.ID == "" || role.Name == "" {
			t.Errorf("%s: %+v", name, role)
		}
	}

	var member Member
	decodeFixture(t, "get_member", &member)
	if member.User == nil || member.User.ID == "" || member.JoinedAt == "" {
		t.Errorf("member fields missing: %+v", member)
	}
	var members []Member
	decodeFixture(t, "get_guild_members", &members)
	if len(members) == 0 {
		t.Error("no members decoded")
	}

	for _, name := range []string{"get_channel", "create_channel", "modify_channel"} {
		var channel Channel
		decodeFixture(t, name, &channel)
		if channel.ID == "" || channel.GuildID == "" {
			t.Errorf("%s: %+v", name, channel)
		}
	}
	var topic Channel
	decodeFixture(t, "modify_channel", &topic)
	if topic.Topic == nil || *topic.Topic == "" {
		t.Error("channel topic was not decoded")
	}
	var channels []Channel
	decodeFixture(t, "get_guild_channels", &channels)
	if len(channels) == 0 {
		t.Error("no channels decoded")
	}

	var message Message
	decodeFixture(t, "get_message", &message)
	if message.ID == "" || message.Author.ID == "" || message.Timestamp == "" {
		t.Errorf("message fields missing: %+v", message)
	}

	var pins PinsPage
	decodeFixture(t, "get_pins", &pins)

	var invite Invite
	decodeFixture(t, "create_invite", &invite)
	if invite.Code == "" || invite.Inviter == nil || invite.Guild == nil || invite.Channel == nil || invite.ExpiresAt == nil {
		t.Errorf("invite fields missing: %+v", invite)
	}

	var webhooks []Webhook
	decodeFixture(t, "get_channel_webhooks", &webhooks)
	if len(webhooks) != 1 || webhooks[0].User == nil {
		t.Errorf("webhooks: %+v", webhooks)
	}

	var app struct {
		ID  string `json:"id"`
		Bot User   `json:"bot"`
	}
	decodeFixture(t, "get_application_me", &app)
	if app.ID == "" || app.ID != app.Bot.ID {
		t.Errorf("application ID is not the bot user ID: %+v", app)
	}

	var command ApplicationCommand
	decodeFixture(t, "create_guild_command", &command)
	if command.ID == "" || command.Version == "" || len(command.Options) != 4 || command.Options[1].MaxValue == nil || len(command.Options[1].Choices) != 2 || !command.Options[3].Autocomplete {
		t.Errorf("guild command: %+v", command)
	}
	var global ApplicationCommand
	decodeFixture(t, "create_global_command", &global)
	if global.ID == "" || global.GuildID != "" || global.DMPermission == nil {
		t.Errorf("global command: %+v", global)
	}
	var commands []ApplicationCommand
	decodeFixture(t, "bulk_guild_commands", &commands)
	if len(commands) != 4 || commands[1].DefaultMemberPermissions == nil || !commands[1].DefaultMemberPermissions.Has(PermissionAdministrator) || commands[1].NameLocalizations["tr"] == "" || commands[2].Type != CommandTypeUser || commands[3].Type != CommandTypeMessage {
		t.Errorf("bulk commands: %+v", commands)
	}
	for _, c := range commands {
		if err := c.Validate(); err != nil {
			t.Errorf("Micup accepted %s but Validate refuses it: %v", c.Name, err)
		}
	}

	var user User
	decodeFixture(t, "get_user_me", &user)
	if user.ID == "" || !user.Bot || user.Discriminator == "" {
		t.Errorf("user fields missing: %+v", user)
	}
}

func TestEveryRESTFixtureIsChecked(t *testing.T) {
	source, err := os.ReadFile("rest_fixtures_test.go")
	if err != nil {
		t.Fatal(err)
	}
	files, _ := filepath.Glob(filepath.Join("testdata", "rest", "*.json"))
	for _, file := range files {
		name := strings.TrimSuffix(filepath.Base(file), ".json")
		if !strings.Contains(string(source), `"`+name+`"`) {
			t.Errorf("testdata/rest/%s.json is not decoded by any test", name)
		}
	}
}
