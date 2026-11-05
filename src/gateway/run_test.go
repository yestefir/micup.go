package gateway

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestRunIdentifiesAndDeliversEvents(t *testing.T) {
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("v") != ProtocolVersion || r.URL.Query().Get("encoding") != Encoding {
			t.Error("Gateway protocol query missing")
		}
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer ws.Close()
		_ = ws.WriteJSON(map[string]any{"op": OpHello, "d": map[string]int{"heartbeat_interval": 5000}})
		var identify struct {
			Op   int `json:"op"`
			Data struct {
				Token string `json:"token"`
				Shard []int  `json:"shard"`
			} `json:"d"`
		}
		if ws.ReadJSON(&identify) != nil || identify.Op != OpIdentify || identify.Data.Token != "fixture" || len(identify.Data.Shard) != 2 || identify.Data.Shard[0] != 1 || identify.Data.Shard[1] != 2 {
			t.Error("invalid Identify packet")
			return
		}
		_ = ws.WriteJSON(map[string]any{"op": OpDispatch, "s": 17, "t": EventReady, "d": map[string]string{"session_id": "session"}})
		var sent struct {
			Op   int               `json:"op"`
			Data map[string]string `json:"d"`
		}
		if ws.ReadJSON(&sent) != nil || sent.Op != OpVoiceStateUpdate || sent.Data["guild_id"] != "guild" {
			t.Error("Gateway event could not send an opcode")
		}
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	want := errors.New("handler complete")
	resume := &Resume{}
	err := Run(ctx, Options{URL: "ws" + strings.TrimPrefix(server.URL, "http"), Token: "fixture", ShardID: 1, ShardCount: 2, Resume: resume}, func(ctx context.Context, event Event) error {
		if event.Name != EventReady || event.Sequence == nil || *event.Sequence != 17 || event.ShardID != 1 {
			t.Fatal("Gateway event fields were lost")
		}
		var ready struct {
			SessionID string `json:"session_id"`
		}
		if event.Decode(&ready) != nil || ready.SessionID != "session" {
			t.Fatal("Gateway event data was changed")
		}
		if err := event.Send(ctx, OpVoiceStateUpdate, map[string]string{"guild_id": "guild", "channel_id": "voice"}); err != nil {
			t.Fatal(err)
		}
		return want
	})
	if !errors.Is(err, want) {
		t.Fatalf("Gateway returned the wrong handler result: %v", err)
	}
	if resume.Resumable() {
		t.Fatal("READY was acknowledged although its handler failed")
	}
}

func TestOptionsValidation(t *testing.T) {
	for name, options := range map[string]Options{
		"no token":  {URL: "wss://micup.gg/gateway"},
		"bad url":   {URL: "https://micup.gg/gateway", Token: "t"},
		"bad shard": {URL: "wss://micup.gg/gateway", Token: "t", ShardID: 2, ShardCount: 2},
	} {
		if _, err := options.normalize(); err == nil {
			t.Fatalf("%s accepted", name)
		}
	}
}

func TestCloseErrorClassification(t *testing.T) {
	if !(&CloseError{Code: 4004}).Fatal() || (&CloseError{Code: 4000}).Fatal() {
		t.Fatal("fatal close codes misclassified")
	}
	if !(&CloseError{Code: 4009}).ResetsSession() || (&CloseError{Code: 4000}).ResetsSession() {
		t.Fatal("session-reset close codes misclassified")
	}
}

func TestRunAcknowledgesOnlyProcessedEvents(t *testing.T) {
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer ws.Close()
		_ = ws.WriteJSON(map[string]any{"op": OpHello, "d": map[string]int{"heartbeat_interval": 5000}})
		var identify map[string]any
		_ = ws.ReadJSON(&identify)
		_ = ws.WriteJSON(map[string]any{"op": OpDispatch, "s": 1, "t": EventReady, "d": map[string]string{"session_id": "s"}})
		_ = ws.WriteJSON(map[string]any{"op": OpDispatch, "s": 2, "t": EventMessageCreate, "d": map[string]string{}})
		_, _, _ = ws.ReadMessage()
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	failed := errors.New("handler failed")
	resume := &Resume{}
	err := Run(ctx, Options{URL: "ws" + strings.TrimPrefix(server.URL, "http"), Token: "fixture", Resume: resume}, func(_ context.Context, event Event) error {
		if event.Name == EventMessageCreate {
			return failed
		}
		return nil
	})
	if !errors.Is(err, failed) {
		t.Fatalf("handler error was not returned: %v", err)
	}
	if last := resume.last(); last == nil || *last != 1 {
		t.Fatalf("failed event was acknowledged: %v", last)
	}
}
