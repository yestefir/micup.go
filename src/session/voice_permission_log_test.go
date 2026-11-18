package session_test

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"sync"
	"testing"

	"github.com/yestefir/micup.go/src/internal/fixture"
	"github.com/yestefir/micup.go/src/session"
)

type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func TestJoinVoiceLogsTheMissingPermission(t *testing.T) {
	p := fixture.New(t)
	out := &syncBuffer{}
	s := newSession(t, p, session.Options{Logger: slog.New(slog.NewTextHandler(out, nil))})
	run(t, s)
	fixture.Eventually(t, s.Ready)
	if _, err := s.JoinVoice(context.Background(), fixture.GuildID, fixture.LockedID, session.VoiceJoin{}); err == nil {
		t.Fatal("join into a locked channel was sent")
	}
	log := out.String()
	for _, want := range []string{"level=WARN", "bot lacks permission to join the voice channel", "guild=" + fixture.GuildID, "channel=" + fixture.LockedID} {
		if !strings.Contains(log, want) {
			t.Errorf("log %q is missing %q", log, want)
		}
	}
}
