package rest

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yestefir/micup.go/src/model"
)

func capturedResponse(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "model", "testdata", "rest", name+".json"))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestCommandEndpoints(t *testing.T) {
	single := capturedResponse(t, "create_guild_command")
	bulk := capturedResponse(t, "bulk_guild_commands")
	c, rec := recordingClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodPut || r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/commands"):
			w.Write(bulk)
		default:
			w.Write(single)
		}
	})
	ctx := context.Background()
	echo := model.ApplicationCommand{Name: "echo", Description: "Repeat text", Options: []model.CommandOption{{Type: model.OptionTypeString, Name: "text", Description: "What", Required: true}}}

	created, err := c.CreateGuildCommand(ctx, "app", "g", echo)
	if err != nil || created.ID == "" || rec.last().method != http.MethodPost || rec.last().path != "/applications/app/guilds/g/commands" || rec.last().body["type"] != float64(1) {
		t.Fatalf("create guild command: %v %+v", err, rec.last())
	}
	if _, err = c.CreateGlobalCommand(ctx, "app", echo); err != nil || rec.last().path != "/applications/app/commands" {
		t.Fatalf("create global command: %v %+v", err, rec.last())
	}
	if _, err = c.GuildCommand(ctx, "app", "g", "c1"); err != nil || rec.last().path != "/applications/app/guilds/g/commands/c1" {
		t.Fatalf("get guild command: %v", err)
	}
	if _, err = c.GlobalCommand(ctx, "app", "c1"); err != nil || rec.last().path != "/applications/app/commands/c1" {
		t.Fatalf("get global command: %v", err)
	}
	list, err := c.GuildCommands(ctx, "app", "g")
	if err != nil || len(list) != 4 {
		t.Fatalf("list guild commands: %v %d", err, len(list))
	}
	if _, err = c.GlobalCommands(ctx, "app"); err != nil || rec.last().method != http.MethodGet {
		t.Fatalf("list global commands: %v", err)
	}
	dm := false
	if _, err = c.EditGuildCommand(ctx, "app", "g", "c1", model.ApplicationCommand{Description: "New", DMPermission: &dm}); err != nil || rec.last().method != http.MethodPatch {
		t.Fatalf("edit guild command: %v", err)
	}
	if body := rec.last().body; len(body) != 2 || body["description"] != "New" || body["dm_permission"] != false {
		t.Fatalf("edit sent %v", body)
	}
	if _, err = c.EditGlobalCommand(ctx, "app", "c1", model.ApplicationCommand{Name: "renamed"}); err != nil || rec.last().path != "/applications/app/commands/c1" {
		t.Fatalf("edit global command: %v", err)
	}
	set, err := c.SetGuildCommands(ctx, "app", "g", []model.ApplicationCommand{echo, {Type: model.CommandTypeUser, Name: "Inspect user"}})
	if err != nil || len(set) != 4 || rec.last().method != http.MethodPut {
		t.Fatalf("set guild commands: %v", err)
	}
	if _, err = c.SetGlobalCommands(ctx, "app", nil); err != nil || string(rec.last().raw) != "[]" {
		t.Fatalf("clearing global commands sent %s: %v", rec.last().raw, err)
	}
	if err = c.DeleteGuildCommand(ctx, "app", "g", "c1"); err != nil || rec.last().method != http.MethodDelete {
		t.Fatalf("delete guild command: %v", err)
	}
	if err = c.DeleteGlobalCommand(ctx, "app", "c1"); err != nil || rec.last().path != "/applications/app/commands/c1" {
		t.Fatalf("delete global command: %v", err)
	}
}

func TestCommandEndpointsRejectBadInput(t *testing.T) {
	c, rec := recordingClient(t, nil)
	ctx := context.Background()
	valid := model.ApplicationCommand{Name: "ping", Description: "Ping"}
	if _, err := c.CreateGlobalCommand(ctx, "", valid); err == nil {
		t.Error("missing application ID accepted")
	}
	if _, err := c.CreateGuildCommand(ctx, "app", "", valid); err == nil {
		t.Error("missing guild ID accepted")
	}
	if _, err := c.CreateGlobalCommand(ctx, "app", model.ApplicationCommand{Name: "Ping"}); err == nil {
		t.Error("invalid command accepted")
	}
	if err := c.DeleteGlobalCommand(ctx, "app", ""); err == nil {
		t.Error("missing command ID accepted")
	}
	if _, err := c.SetGlobalCommands(ctx, "app", make([]model.ApplicationCommand, maxBulkCommands+1)); err == nil {
		t.Error("too many commands accepted")
	}
	if _, err := c.SetGuildCommands(ctx, "app", "g", []model.ApplicationCommand{valid, {Name: "bad name"}}); err == nil {
		t.Error("invalid bulk command accepted")
	}
	if len(rec.requests) != 0 {
		t.Fatalf("invalid input reached Micup: %+v", rec.requests)
	}
}

func TestRespondAutocomplete(t *testing.T) {
	c, rec := recordingClient(t, nil)
	ctx := context.Background()
	i := model.Interaction{ID: "i", Token: "t", Type: model.InteractionTypeAutocomplete}
	if err := c.RespondAutocomplete(ctx, i, []model.CommandChoice{{Name: "Apple", Value: "apple"}}); err != nil {
		t.Fatal(err)
	}
	last := rec.last()
	data, _ := last.body["data"].(map[string]any)
	choices, _ := data["choices"].([]any)
	if last.path != "/interactions/i/t/callback" || last.body["type"] != float64(8) || len(choices) != 1 {
		t.Fatalf("autocomplete sent %+v", last)
	}
	if err := c.RespondAutocomplete(ctx, i, nil); err != nil || !strings.Contains(string(rec.last().raw), `"choices":[]`) {
		t.Fatalf("empty choices sent %s: %v", rec.last().raw, err)
	}
	if err := c.RespondAutocomplete(ctx, model.Interaction{ID: "i", Token: "t", Type: model.InteractionTypeApplicationCommand}, nil); err == nil {
		t.Fatal("command interaction accepted autocomplete choices")
	}
	if err := c.RespondAutocomplete(ctx, i, make([]model.CommandChoice, model.MaxCommandChoices+1)); err == nil {
		t.Fatal("too many choices accepted")
	}
}

func TestDeferInteractionOmitsEmptyData(t *testing.T) {
	c, rec := recordingClient(t, nil)
	ctx := context.Background()
	i := model.Interaction{ID: "i", Token: "t", Type: model.InteractionTypeApplicationCommand}
	if err := c.DeferInteraction(ctx, i, false); err != nil {
		t.Fatal(err)
	}
	if got := string(rec.last().raw); got != `{"type":5}` {
		t.Fatalf("visible defer sent %s; Micup rejects an empty data object", got)
	}
	if err := c.DeferInteraction(ctx, i, true); err != nil {
		t.Fatal(err)
	}
	if got := string(rec.last().raw); got != `{"data":{"flags":64},"type":5}` {
		t.Fatalf("ephemeral defer sent %s", got)
	}
}
