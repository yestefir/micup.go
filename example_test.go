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

func ExampleFormatTimestamp() {
	fmt.Println(micup.FormatTimestamp(time.Unix(1700000000, 0), micup.TimestampRelative))
	// Output:
	// <t:1700000000:R>
}
