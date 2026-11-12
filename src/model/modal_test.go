package model

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestModalMarshalsOneInputPerRow(t *testing.T) {
	modal := Modal{CustomID: "song", Title: "Şarkı iste", Inputs: []TextInput{
		ShortInput("query", "Şarkı").WithLength(2, 100).WithRequired(true).WithPlaceholder("Tarkan Kuzu Kuzu"),
		ParagraphInput("note", "Not").WithValue("hi"),
	}}
	if err := modal.Validate(); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(modal)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"components":[{"components":[{"custom_id":"query","label":"Şarkı","max_length":100,"min_length":2,"placeholder":"Tarkan Kuzu Kuzu","required":true,"style":1,"type":4}],"type":1},{"components":[{"custom_id":"note","label":"Not","style":2,"type":4,"value":"hi"}],"type":1}],"custom_id":"song","title":"Şarkı iste"}`
	if string(data) != want {
		t.Fatalf("modal JSON:\n%s\nwant\n%s", data, want)
	}
}

func TestModalValidate(t *testing.T) {
	ok := ShortInput("a", "A")
	bad := map[string]Modal{
		"no title":       {CustomID: "m", Inputs: []TextInput{ok}},
		"long title":     {CustomID: "m", Title: strings.Repeat("a", MaxModalTitle+1), Inputs: []TextInput{ok}},
		"no custom id":   {Title: "t", Inputs: []TextInput{ok}},
		"no inputs":      {CustomID: "m", Title: "t"},
		"six inputs":     {CustomID: "m", Title: "t", Inputs: []TextInput{ShortInput("1", "1"), ShortInput("2", "2"), ShortInput("3", "3"), ShortInput("4", "4"), ShortInput("5", "5"), ShortInput("6", "6")}},
		"duplicate":      {CustomID: "m", Title: "t", Inputs: []TextInput{ok, ok}},
		"no label":       {CustomID: "m", Title: "t", Inputs: []TextInput{{CustomID: "a", Style: TextInputStyleShort}}},
		"bad style":      {CustomID: "m", Title: "t", Inputs: []TextInput{{CustomID: "a", Label: "A", Style: 3}}},
		"min over max":   {CustomID: "m", Title: "t", Inputs: []TextInput{ok.WithLength(10, 5)}},
		"max too large":  {CustomID: "m", Title: "t", Inputs: []TextInput{ok.WithLength(0, MaxTextInputLength+1)}},
		"value too long": {CustomID: "m", Title: "t", Inputs: []TextInput{ok.WithLength(0, 2).WithValue("abc")}},
	}
	for name, modal := range bad {
		if err := modal.Validate(); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
}

func TestModalDataReadsSubmittedValues(t *testing.T) {
	i := Interaction{Type: InteractionTypeModalSubmit, Data: json.RawMessage(`{"custom_id":"song","components":[{"type":1,"components":[{"type":4,"custom_id":"query","value":"kuzu"}]},{"type":1,"components":[{"type":4,"custom_id":"note","value":""}]}]}`)}
	data, ok := i.ModalData()
	if !ok || !i.IsModalSubmit() || data.CustomID != "song" {
		t.Fatalf("modal data: %+v %v", data, ok)
	}
	if v, ok := data.Value("query"); !ok || v != "kuzu" {
		t.Fatalf("query = %q %v", v, ok)
	}
	if v, ok := data.Value("note"); !ok || v != "" {
		t.Fatalf("empty note = %q %v", v, ok)
	}
	if _, ok := data.Value("missing"); ok {
		t.Fatal("missing input found")
	}
	if _, ok := (Interaction{Type: InteractionTypeApplicationCommand, Data: i.Data}).ModalData(); ok {
		t.Fatal("command read as modal submit")
	}
}
