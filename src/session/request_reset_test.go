package session_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yestefir/micup.go/src/gateway"
	"github.com/yestefir/micup.go/src/internal/fixture"
	"github.com/yestefir/micup.go/src/model"
	"github.com/yestefir/micup.go/src/session"
)

func TestPendingRequestsEndWhenGatewaySessionIsReplaced(t *testing.T) {
	for _, transition := range []string{"invalid_session", "new_ready", "invalid_sequence_close", "session_timeout_close"} {
		for _, kind := range []string{"members", "guild_counts", "channel_counts"} {
			t.Run(transition+"/"+kind, func(t *testing.T) {
				p := fixture.New(t)
				p.HoldMembers, p.HoldCounts = true, true
				allowReidentify := make(chan struct{})
				var identifies atomic.Int64
				s, err := session.New(session.Config{Endpoint: p.GatewayURL(), Token: fixture.Token, BotID: fixture.BotID, BeforeIdentify: func(ctx context.Context) error {
					if identifies.Add(1) == 1 {
						return nil
					}
					select {
					case <-allowReidentify:
						return nil
					case <-ctx.Done():
						return ctx.Err()
					}
				}}, session.Options{MinBackoff: time.Millisecond})
				if err != nil {
					t.Fatal(err)
				}
				run(t, s)
				fixture.Eventually(t, s.Ready)
				request := func(ctx context.Context) error {
					var err error
					switch kind {
					case "members":
						_, err = s.RequestGuildMembers(ctx, model.MembersRequest{GuildID: fixture.GuildID})
					case "guild_counts":
						_, err = s.RequestGuildCounts(ctx, []string{fixture.GuildID})
					case "channel_counts":
						_, err = s.RequestChannelMemberCounts(ctx, fixture.GuildID, []string{fixture.VoiceID})
					}
					return err
				}
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				done := make(chan error, 1)
				go func() { done <- request(ctx) }()
				fixture.Eventually(t, func() bool { return p.MembersRequested.Load()+p.CountsRequested.Load() == 1 })
				switch transition {
				case "invalid_session":
					p.Send(map[string]any{"op": gateway.OpInvalidSession, "d": false})
				case "new_ready":
					p.Event(gateway.EventReady, map[string]any{"session_id": "replacement", "guilds": []any{}})
				case "invalid_sequence_close":
					p.CloseGateway(gateway.CloseInvalidSequence)
				case "session_timeout_close":
					p.CloseGateway(gateway.CloseSessionTimeout)
				}
				if err := <-done; !errors.Is(err, gateway.ErrClosed) {
					t.Errorf("old session request survived replacement: %v", err)
				}
				close(allowReidentify)
				fixture.Eventually(t, s.Ready)
				p.AllowRequestResponses()
				freshCtx, stopFresh := context.WithTimeout(context.Background(), 3*time.Second)
				defer stopFresh()
				if err := request(freshCtx); err != nil {
					t.Fatalf("replacement session could not answer a fresh request: %v", err)
				}
			})
		}
	}
}
