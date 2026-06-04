package internal_test

import (
	"context"
	"sync"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/struckchure/rt-comms/internal"
	"github.com/struckchure/rt-comms/ws"
)

func TestHubRegisterReplaceDeliver(t *testing.T) {
	t.Parallel()
	h, _ := internal.NewHub(nil)
	ctx := context.Background()

	c1 := ws.NewClient("alice", &websocket.Conn{})
	h.Register("alice", c1)
	if !h.IsOnline("alice") {
		t.Fatal("expected alice online")
	}

	c2 := ws.NewClient("alice", &websocket.Conn{})
	h.Register("alice", c2)
	select {
	case <-c1.Done:
	default:
		t.Fatal("expected old client closed on replace")
	}

	payload := []byte(`{"type":"message"}`)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case got := <-c2.Send:
			if string(got) != string(payload) {
				t.Errorf("payload mismatch: %s", got)
			}
		case <-c2.Done:
			t.Fatal("client closed before deliver")
		}
	}()
	_ = h.DeliverTo(ctx, "alice", payload)
	wg.Wait()

	_ = h.DeliverTo(ctx, "bob", payload)
	_ = h.Shutdown(ctx)
}
