package model

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func testCatalog(t *testing.T) *Catalog {
	t.Helper()
	c, err := NewCatalog("en-US", Messages{
		"en-US": {"play": "play", "playing": "Now playing %s", "join": "Join a voice channel first."},
		"tr":    {"play": "oynat", "playing": "%s çalınıyor"},
		"pt-BR": {"play": "tocar"},
	})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestCatalogPicksTheInteractionLocale(t *testing.T) {
	c := testCatalog(t)
	for _, tc := range []struct {
		interaction Interaction
		key         string
		want        string
	}{
		{Interaction{Locale: "tr"}, "playing", "x çalınıyor"},
		{Interaction{Locale: "en-US"}, "playing", "Now playing x"},
		{Interaction{Locale: "tr"}, "join", "Join a voice channel first."},
		{Interaction{Locale: "de"}, "playing", "Now playing x"},
		{Interaction{}, "playing", "Now playing x"},
		{Interaction{GuildLocale: "tr"}, "playing", "x çalınıyor"},
		{Interaction{Locale: "TR"}, "playing", "x çalınıyor"},
		{Interaction{Locale: "pt"}, "play", "tocar"},
		{Interaction{Locale: "en-GB"}, "play", "play"},
		{Interaction{Locale: "tr"}, "missing", "missing"},
	} {
		args := []any{}
		if tc.key == "playing" {
			args = append(args, "x")
		}
		if got := c.Text(tc.interaction, tc.key, args...); got != tc.want {
			t.Errorf("Text(%+v, %q) = %q, want %q", tc.interaction.Locale, tc.key, got, tc.want)
		}
	}
	if got := c.TextIn("tr_TR", "play"); got != "oynat" {
		t.Fatalf("TextIn(tr_TR) = %q", got)
	}
}

func TestCatalogLocalizationsSkipTheFallback(t *testing.T) {
	c := testCatalog(t)
	want := map[string]string{"tr": "oynat", "pt-BR": "tocar"}
	if got := c.Localizations("play"); !reflect.DeepEqual(got, want) {
		t.Fatalf("Localizations = %v, want %v", got, want)
	}
	if got := c.Localizations("join"); got != nil {
		t.Fatalf("Localizations of an untranslated key = %v, want nil", got)
	}
	if c.Fallback() != "en-US" || !reflect.DeepEqual(c.Locales(), []string{"en-US", "pt-BR", "tr"}) {
		t.Fatalf("Fallback %q, Locales %v", c.Fallback(), c.Locales())
	}
}

func TestNewCatalogRejectsMistakes(t *testing.T) {
	for name, tc := range map[string]struct {
		fallback string
		messages Messages
		want     string
	}{
		"missing fallback": {"en", Messages{"tr": {"a": "b"}}, "fallback"},
		"bad locale":       {"en", Messages{"en": {}, "turkish!": {}}, "language tag"},
		"duplicate":        {"en", Messages{"en-US": {}, "en_us": {}, "en": {}}, "twice"},
		"stray key":        {"en", Messages{"en": {"a": "b"}, "tr": {"c": "d"}}, "not in the fallback"},
	} {
		if _, err := NewCatalog(tc.fallback, tc.messages); err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: error = %v, want %q", name, err, tc.want)
		}
	}
}

func TestCommandLocalizationsAreValidated(t *testing.T) {
	valid := ApplicationCommand{
		Name:                     "play",
		Description:              "Play a song",
		NameLocalizations:        map[string]string{"tr": "oynat"},
		DescriptionLocalizations: map[string]string{"tr": "Bir şarkı çal"},
		Options: []CommandOption{{
			Type: OptionTypeString, Name: "query", Description: "Song name or link", Required: true,
			NameLocalizations:        map[string]string{"tr": "şarkı"},
			DescriptionLocalizations: map[string]string{"tr": "Şarkı adı veya bağlantısı"},
		}},
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid localized command: %v", err)
	}
	for name, mutate := range map[string]func(*ApplicationCommand){
		"uppercase name":     func(c *ApplicationCommand) { c.NameLocalizations["tr"] = "Oynat" },
		"long name":          func(c *ApplicationCommand) { c.NameLocalizations["tr"] = strings.Repeat("a", 33) },
		"bad locale":         func(c *ApplicationCommand) { c.NameLocalizations["Türkçe"] = "oynat" },
		"empty description":  func(c *ApplicationCommand) { c.DescriptionLocalizations["tr"] = "" },
		"long description":   func(c *ApplicationCommand) { c.DescriptionLocalizations["tr"] = strings.Repeat("a", 101) },
		"option name":        func(c *ApplicationCommand) { c.Options[0].NameLocalizations["tr"] = "Şarkı" },
		"option description": func(c *ApplicationCommand) { c.Options[0].DescriptionLocalizations["tr"] = strings.Repeat("a", 101) },
		"choice name": func(c *ApplicationCommand) {
			c.Options[0].Choices = []CommandChoice{{Name: "a", Value: "a", NameLocalizations: map[string]string{"tr": strings.Repeat("a", 101)}}}
		},
	} {
		command := valid
		command.NameLocalizations = map[string]string{"tr": "oynat"}
		command.DescriptionLocalizations = map[string]string{"tr": "Bir şarkı çal"}
		option := valid.Options[0]
		option.NameLocalizations = map[string]string{"tr": "şarkı"}
		option.DescriptionLocalizations = map[string]string{"tr": "Şarkı adı"}
		command.Options = []CommandOption{option}
		mutate(&command)
		if command.Validate() == nil {
			t.Errorf("%s was accepted", name)
		}
	}
	menu := ApplicationCommand{Type: CommandTypeUser, Name: "Profile", NameLocalizations: map[string]string{"tr": "Profil"}}
	if err := menu.Validate(); err != nil {
		t.Fatalf("localized context menu command: %v", err)
	}
	menu.NameLocalizations["tr"] = ""
	if menu.Validate() == nil {
		t.Fatal("empty context menu localization was accepted")
	}
}

func TestCommandLocalizationsSerialize(t *testing.T) {
	data, err := json.Marshal(ApplicationCommand{
		Name: "play", Description: "Play",
		NameLocalizations: map[string]string{"tr": "oynat"},
		Options:           []CommandOption{{Type: OptionTypeString, Name: "query", Description: "Song", NameLocalizations: map[string]string{"tr": "şarkı"}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"name_localizations":{"tr":"oynat"}`, `"name_localizations":{"tr":"şarkı"}`} {
		if !strings.Contains(string(data), want) {
			t.Errorf("%s missing from %s", want, data)
		}
	}
}
