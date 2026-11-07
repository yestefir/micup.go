package model

import "encoding/json"

func (m AllowedMentions) MarshalJSON() ([]byte, error) {
	type policy AllowedMentions
	parse := m.Parse
	if parse == nil {
		parse = []string{}
	}
	return json.Marshal(struct {
		policy
		Parse []string `json:"parse"`
	}{policy: policy(m), Parse: parse})
}
