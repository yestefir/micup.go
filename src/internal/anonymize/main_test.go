package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func anonymize(t *testing.T, m *Mapping, raw string) map[string]any {
	t.Helper()
	out, err := m.Anonymize(strings.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	var v map[string]any
	dec := json.NewDecoder(strings.NewReader(string(out)))
	dec.UseNumber()
	if err := dec.Decode(&v); err != nil {
		t.Fatal(err)
	}
	return v
}

func TestAnonymizeReplacesIdentifyingValues(t *testing.T) {
	m := NewMapping([]string{"payload"})
	v := anonymize(t, m, `{
		"id": "1234567890123456789",
		"owner": {"id": "1234567890123456790", "username": "realname", "global_name": null, "avatar": "abc"},
		"bot": {"username": "payload"},
		"resolved": {"roles": {"1234567890123456791": {"id": "1234567890123456791", "permissions": "18014914009747009"}}},
		"changes": [{"key": "permissions", "new_value": "18014914009747009"}, {"key": "role_id", "new_value": "1234567890123456791"}],
		"token": "secret", "session_id": "s", "latitude": "52.1", "country_code": "TR", "email": "a@b.c",
		"code": "abcd", "attachment_size_limit": 52428800, "emoji_name": "myhandle"
	}`)
	if v["id"] != "100000000000000001" {
		t.Errorf("id: %v", v["id"])
	}
	owner := v["owner"].(map[string]any)
	if owner["id"] != "100000000000000002" || owner["username"] != "user1" || owner["global_name"] != nil || owner["avatar"] != "hash" {
		t.Errorf("owner: %v", owner)
	}
	if v["bot"].(map[string]any)["username"] != "payload" {
		t.Error("kept name was replaced")
	}
	roles := v["resolved"].(map[string]any)["roles"].(map[string]any)
	role, ok := roles["100000000000000003"].(map[string]any)
	if !ok || role["id"] != "100000000000000003" || role["permissions"] != "18014914009747009" {
		t.Errorf("resolved key or permissions: %v", roles)
	}
	changes := v["changes"].([]any)
	if changes[0].(map[string]any)["new_value"] != "18014914009747009" || changes[1].(map[string]any)["new_value"] != "100000000000000003" {
		t.Errorf("audit changes: %v", changes)
	}
	for key, want := range map[string]any{"token": "redacted", "session_id": "redacted", "latitude": "0", "country_code": "XX", "email": nil, "code": "invitecode", "emoji_name": "emoji"} {
		if v[key] != want {
			t.Errorf("%s: %v", key, v[key])
		}
	}
	if v["attachment_size_limit"] != json.Number("52428800") {
		t.Errorf("number changed: %v", v["attachment_size_limit"])
	}
}

func TestAnonymizeKeepsFieldOrderAndMappingAcrossRuns(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in")
	for name, body := range map[string]string{
		"A_EVENT/0.json": `{"z": "1234567890123456789", "a": "1234567890123456790"}`,
		"B_EVENT/0.json": `{"guild_id": "1234567890123456789"}`,
	} {
		path := filepath.Join(in, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mapFile := filepath.Join(dir, "map.json")
	out := filepath.Join(dir, "out")
	if err := run(in, out, mapFile, nil); err != nil {
		t.Fatal(err)
	}
	first, _ := os.ReadFile(filepath.Join(out, "A_EVENT.0.json"))
	if !strings.Contains(string(first), `"z": "100000000000000001"`) || !strings.Contains(string(first), `"a": "100000000000000002"`) {
		t.Fatalf("IDs not assigned in document order: %s", first)
	}
	second, _ := os.ReadFile(filepath.Join(out, "B_EVENT.0.json"))
	if !strings.Contains(string(second), `"100000000000000001"`) {
		t.Fatalf("same ID mapped differently: %s", second)
	}
	if err := os.WriteFile(filepath.Join(in, "B_EVENT", "1.json"), []byte(`{"guild_id": "1234567890123456789", "user_id": "9999999999999999999"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := run(in, out, mapFile, nil); err != nil {
		t.Fatal(err)
	}
	later, _ := os.ReadFile(filepath.Join(out, "B_EVENT.1.json"))
	if !strings.Contains(string(later), `"guild_id": "100000000000000001"`) || !strings.Contains(string(later), `"user_id": "100000000000000003"`) {
		t.Fatalf("mapping was not reused: %s", later)
	}
}
