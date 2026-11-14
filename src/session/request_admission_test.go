package session

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/yestefir/micup.go/src/model"
)

type requestWaitContext struct {
	context.Context
	waiting chan struct{}
	once    sync.Once
}

func (c *requestWaitContext) Done() <-chan struct{} {
	c.once.Do(func() { close(c.waiting) })
	return c.Context.Done()
}

func TestQueuedRequestsRecheckReadinessBeforeRegistration(t *testing.T) {
	for _, members := range []bool{false, true} {
		name := "counts"
		if members {
			name = "members"
		}
		t.Run(name, func(t *testing.T) {
			s, err := New(Config{Endpoint: "wss://localhost", Token: "fixture"}, Options{})
			if err != nil {
				t.Fatal(err)
			}
			s.ready.Store(true)
			slots := s.countSlots
			if members {
				slots = s.memberSlots
			}
			for range cap(slots) {
				slots <- struct{}{}
			}
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			observed := &requestWaitContext{Context: ctx, waiting: make(chan struct{})}
			done := make(chan error, 1)
			go func() {
				var err error
				if members {
					_, err = s.RequestGuildMembers(observed, model.MembersRequest{GuildID: "g"})
				} else {
					_, err = s.RequestGuildCounts(observed, []string{"g"})
				}
				done <- err
			}()
			select {
			case <-observed.waiting:
			case <-ctx.Done():
				t.Fatal("request did not enter the occupied slot wait")
			}
			s.ready.Store(false)
			<-slots
			if err := <-done; !errors.Is(err, ErrNotReady) {
				t.Errorf("queued request did not reject unready admission: %v", err)
			}
			if len(s.memberRequests) != 0 || len(s.countRequests) != 0 {
				t.Error("unready request left a registration")
			}
		})
	}
}
