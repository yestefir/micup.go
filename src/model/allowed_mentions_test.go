package model

import (
	"encoding/json"
	"testing"
)

func TestAllowedMentionsJSON(t *testing.T) {
	replied := false
	for _, tt := range []struct {
		name   string
		policy AllowedMentions
		want   string
	}{
		{"zero", AllowedMentions{}, `{"parse":[]}`},
		{"empty parse", AllowedMentions{Parse: []string{}}, `{"parse":[]}`},
		{"parse", AllowedMentions{Parse: []string{"users"}}, `{"parse":["users"]}`},
		{"explicit IDs", AllowedMentions{Users: []string{"1"}, Roles: []string{"2"}, RepliedUser: &replied}, `{"users":["1"],"roles":["2"],"replied_user":false,"parse":[]}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.policy)
			if err != nil || string(body) != tt.want {
				t.Fatalf("policy = %s, %v; want %s", body, err, tt.want)
			}
		})
	}
}
