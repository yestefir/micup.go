package state

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yestefir/micup.go/src/gateway"
	"github.com/yestefir/micup.go/src/model"
)

const (
	capturedGuild   = "100000000000000001"
	capturedOwner   = "100000000000000019"
	capturedBot     = "100000000000000011"
	capturedText    = "100000000000000004"
	capturedVoice   = "100000000000000008"
	capturedChannel = "100000000000000002"
	capturedRole    = "100000000000000015"
)

func applyCaptured(t *testing.T, c *Cache, name string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "events", "testdata", name+".0.json"))
	if err != nil {
		t.Fatal(err)
	}
	c.Apply(name, data)
}

func TestCacheFollowsCapturedEvents(t *testing.T) {
	c := New(capturedBot)
	for _, name := range []string{gateway.EventReady, gateway.EventGuildCreate, gateway.EventChannelCreate, gateway.EventGuildRoleCreate, gateway.EventVoiceStateUpdate} {
		applyCaptured(t, c, name)
	}
	if c.Guilds() != 1 || c.Unavailable(capturedGuild) || c.OwnerID(capturedGuild) != capturedOwner {
		t.Fatalf("guild: %d %v %q", c.Guilds(), c.Unavailable(capturedGuild), c.OwnerID(capturedGuild))
	}
	if _, ok := c.Channel(capturedChannel); !ok {
		t.Fatal("created channel is missing")
	}
	if _, ok := c.Role(capturedGuild, capturedRole); !ok {
		t.Fatal("created role is missing")
	}
	if c.VoiceChannel(capturedGuild, capturedOwner) != capturedVoice {
		t.Fatal("voice state was not applied")
	}
	if perms, ok := c.BotPermissions(capturedGuild, capturedText); !ok || !perms.Has(model.PermissionAdministrator) {
		t.Fatalf("bot permissions: %v %v", perms, ok)
	}
	applyCaptured(t, c, gateway.EventChannelDelete)
	applyCaptured(t, c, gateway.EventGuildRoleDelete)
	if _, ok := c.Channel(capturedChannel); ok {
		t.Fatal("deleted channel is still cached")
	}
	if _, ok := c.Role(capturedGuild, capturedRole); ok {
		t.Fatal("deleted role is still cached")
	}
}
