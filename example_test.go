package micup_test

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	micup "github.com/yestefir/micup.go"
)

func ExampleNewEventRouter() {
	router := micup.NewEventRouter()
	router.Command("queue add", func(ctx context.Context, i micup.Interaction) error {
		fmt.Println("add to queue")
		return nil
	})
	router.Component("page:*", func(ctx context.Context, i micup.Interaction) error {
		fmt.Println("change page", i.CustomID())
		return nil
	})

	add := micup.Interaction{
		Type: micup.InteractionTypeApplicationCommand,
		Data: json.RawMessage(`{"name":"queue","options":[{"type":1,"name":"add"}]}`),
	}
	page := micup.Interaction{
		Type: micup.InteractionTypeMessageComponent,
		Data: json.RawMessage(`{"custom_id":"page:3","component_type":2}`),
	}
	unknown := micup.Interaction{
		Type: micup.InteractionTypeApplicationCommand,
		Data: json.RawMessage(`{"name":"other"}`),
	}
	ctx := context.Background()
	fmt.Println(router.Handle(ctx, add))
	fmt.Println(router.Handle(ctx, page))
	fmt.Println(router.Handle(ctx, unknown) == micup.ErrNoRoute)
	// Output:
	// add to queue
	// <nil>
	// change page page:3
	// <nil>
	// true
}

func ExampleNewEventMux() {
	mux := micup.NewEventMux()
	mux.On(micup.EventMessageCreate, micup.Listen(func(ctx context.Context, m micup.MessageCreate) error {
		fmt.Println(m.Content)
		return nil
	}))
	event := micup.Event{Name: micup.EventMessageCreate, Data: json.RawMessage(`{"id":"1","channel_id":"2","content":"!ping"}`)}
	fmt.Println(mux.Handle(context.Background(), event))
	// Output:
	// !ping
	// <nil>
}

func ExampleEventMux_OnGuildTagChange() {
	mux := micup.NewEventMux()
	mux.OnGuildTagChange(func(ctx context.Context, change micup.GuildTagChange) error {
		if change.Cleared() {
			fmt.Printf("%s removed their tag\n", change.UserID)
			return nil
		}
		fmt.Printf("%s now shows the tag of guild %s\n", change.UserID, change.Current)
		return nil
	})
	update := func(tag string) micup.Event {
		data := `{"guild_id":"10","roles":[],"user":{"id":"20","username":"u","tag_guild_id":` + tag + `}}`
		return micup.Event{Name: micup.EventGuildMemberUpdate, Data: json.RawMessage(data)}
	}
	ctx := context.Background()
	fmt.Println(mux.Handle(ctx, update(`"30"`)))
	fmt.Println(mux.Handle(ctx, update(`"30"`)))
	fmt.Println(mux.Handle(ctx, update(`null`)))
	// Output:
	// 20 now shows the tag of guild 30
	// <nil>
	// <nil>
	// 20 removed their tag
	// <nil>
}

func ExampleFormatTimestamp() {
	fmt.Println(micup.FormatTimestamp(time.Unix(1700000000, 0), micup.TimestampRelative))
	// Output:
	// <t:1700000000:R>
}

func ExampleNewCatalog() {
	catalog, err := micup.NewCatalog("en-US", micup.Messages{
		"en-US": {"play.name": "play", "play.description": "Play a song", "playing": "Now playing %s"},
		"tr":    {"play.name": "oynat", "play.description": "Bir şarkı çal", "playing": "%s çalınıyor"},
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	command := micup.ApplicationCommand{
		Name:                     catalog.TextIn("en-US", "play.name"),
		NameLocalizations:        catalog.Localizations("play.name"),
		Description:              catalog.TextIn("en-US", "play.description"),
		DescriptionLocalizations: catalog.Localizations("play.description"),
	}
	fmt.Println(command.Validate(), command.NameLocalizations["tr"])

	interaction := micup.Interaction{Locale: "tr"}
	fmt.Println(catalog.Text(interaction, "playing", "Ezgi"))
	fmt.Println(catalog.Text(micup.Interaction{Locale: "de"}, "playing", "Ezgi"))
	// Output:
	// <nil> oynat
	// Ezgi çalınıyor
	// Now playing Ezgi
}
