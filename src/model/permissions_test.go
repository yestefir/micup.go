package model

import (
	"encoding/json"
	"testing"
)

func TestComputePermissions(t *testing.T) {
	const guild = "g"
	roles := map[string]Role{
		guild: {ID: guild, Permissions: PermissionViewChannel | PermissionConnect | PermissionSendMessages},
		"dj":  {ID: "dj", Permissions: PermissionSpeak},
		"mod": {ID: "mod", Permissions: PermissionAdministrator},
	}
	if got := ComputePermissions(guild, "owner", "owner", roles, nil, nil); got != PermissionAll {
		t.Fatal("owner did not get every permission")
	}
	if got := ComputePermissions(guild, "owner", "u", roles, []string{"mod"}, []PermissionOverwrite{{ID: guild, Type: OverwriteRole, Deny: PermissionViewChannel}}); got != PermissionAll {
		t.Fatal("administrator was limited by an overwrite")
	}
	overwrites := []PermissionOverwrite{
		{ID: guild, Type: OverwriteRole, Deny: PermissionConnect | PermissionSendMessages},
		{ID: "dj", Type: OverwriteRole, Allow: PermissionConnect},
		{ID: "other", Type: OverwriteRole, Deny: PermissionConnect},
		{ID: "u", Type: OverwriteMember, Deny: PermissionSpeak},
	}
	got := ComputePermissions(guild, "owner", "u", roles, []string{"dj", "other"}, overwrites)
	if !got.Has(PermissionViewChannel|PermissionConnect) || got.Has(PermissionSendMessages) || got.Has(PermissionSpeak) {
		t.Fatalf("overwrite order is wrong: %b", got)
	}
}

func TestPermissionsJSON(t *testing.T) {
	var role Role
	if err := json.Unmarshal([]byte(`{"id":"r","permissions":"2251799813685248"}`), &role); err != nil || !role.Permissions.Has(PermissionPinMessages) {
		t.Fatalf("decimal string not decoded: %v %d", err, role.Permissions)
	}
	if err := json.Unmarshal([]byte(`{"id":"r","permissions":1024}`), &role); err != nil || role.Permissions != PermissionViewChannel {
		t.Fatalf("number not decoded: %v", err)
	}
	data, _ := json.Marshal(PermissionConnect)
	if string(data) != `"1048576"` {
		t.Fatalf("permissions not encoded as a string: %s", data)
	}
}
