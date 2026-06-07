package sse

import (
	"testing"
	"time"
)

// newTestClient builds a client subscribed to the given channels with a small
// buffer, ready to Register.
func newTestClient(channels ...string) *Client {
	return &Client{
		Channels:    channels,
		Events:      make(chan SSEEvent, ClientBufferSize),
		Done:        make(chan struct{}),
		ConnectedAt: time.Now(),
	}
}

func TestBroadcastDeliversToSubscribers(t *testing.T) {
	h := NewHub()
	c := newTestClient("meows")
	h.Register(c)

	h.Broadcast("meows", SSEEvent{Event: "meow-created", Data: "<li>hi</li>"})

	select {
	case ev := <-c.Events:
		if ev.Event != "meow-created" || ev.Data != "<li>hi</li>" {
			t.Errorf("unexpected event: %+v", ev)
		}
	default:
		t.Fatal("expected an event, got none")
	}
}

func TestBroadcastSkipsOtherChannels(t *testing.T) {
	h := NewHub()
	c := newTestClient("meows")
	h.Register(c)

	h.Broadcast("other", SSEEvent{Event: "x"})

	select {
	case ev := <-c.Events:
		t.Fatalf("did not expect an event, got %+v", ev)
	default:
	}
}

func TestBroadcastFanOut(t *testing.T) {
	h := NewHub()
	a := newTestClient("meows")
	b := newTestClient("meows")
	h.Register(a)
	h.Register(b)

	h.Broadcast("meows", SSEEvent{Event: "ping"})

	for name, c := range map[string]*Client{"a": a, "b": b} {
		select {
		case <-c.Events:
		default:
			t.Errorf("client %s did not receive the event", name)
		}
	}

	conns, events := h.Stats()
	if conns != 2 {
		t.Errorf("got %d connections, want 2", conns)
	}
	if events != 2 {
		t.Errorf("got %d delivered events, want 2", events)
	}
}

func TestUnregisterRemovesClientAndClosesEvents(t *testing.T) {
	h := NewHub()
	c := newTestClient("meows")
	h.Register(c)

	h.Unregister(c, "client_disconnect")

	// Events channel is closed.
	if _, ok := <-c.Events; ok {
		t.Error("expected Events channel to be closed after Unregister")
	}

	// A subsequent broadcast must not panic or deliver.
	h.Broadcast("meows", SSEEvent{Event: "after-close"})

	conns, _ := h.Stats()
	if conns != 0 {
		t.Errorf("got %d connections after unregister, want 0", conns)
	}
}

func TestUnregisterIsIdempotent(t *testing.T) {
	h := NewHub()
	c := newTestClient("meows")
	h.Register(c)

	// Two unregisters (simulating reaper + handler race) must not double-count
	// or double-close.
	h.Unregister(c, "idle_timeout")
	h.Unregister(c, "client_disconnect")

	if conns, _ := h.Stats(); conns != 0 {
		t.Errorf("got %d connections after double unregister, want 0", conns)
	}
}

func TestBroadcastDropsWhenBufferFull(t *testing.T) {
	h := NewHub()
	c := &Client{
		Channels:    []string{"meows"},
		Events:      make(chan SSEEvent, 1), // tiny buffer
		Done:        make(chan struct{}),
		ConnectedAt: time.Now(),
	}
	h.Register(c)

	// First fills the buffer, second must be dropped (not block).
	h.Broadcast("meows", SSEEvent{Event: "1"})
	h.Broadcast("meows", SSEEvent{Event: "2"})

	if _, dropped := h.Health().DroppedEvents, h.Health().DroppedEvents; dropped != 1 {
		t.Errorf("got %d dropped events, want 1", h.Health().DroppedEvents)
	}
}

func TestHealthSnapshot(t *testing.T) {
	h := NewHub()
	c := newTestClient("meows")
	h.Register(c)
	h.Broadcast("meows", SSEEvent{Event: "x"})

	info := h.Health()
	if info.ActiveConnections != 1 {
		t.Errorf("ActiveConnections = %d, want 1", info.ActiveConnections)
	}
	if info.TotalEvents != 1 {
		t.Errorf("TotalEvents = %d, want 1", info.TotalEvents)
	}
}

func TestClientActivityTracking(t *testing.T) {
	c := newTestClient("meows")
	// Before any touch, LastActivity falls back to ConnectedAt.
	if !c.LastActivity().Equal(c.ConnectedAt) {
		t.Error("LastActivity should default to ConnectedAt")
	}
	c.TouchActivity()
	if !c.LastActivity().After(c.ConnectedAt) && c.LastActivity().Before(c.ConnectedAt) {
		t.Error("TouchActivity should advance LastActivity")
	}
}

func TestEvictReasonRoundTrip(t *testing.T) {
	c := newTestClient("meows")
	if c.GetEvictReason() != "" {
		t.Error("evict reason should be empty initially")
	}
	c.SetEvictReason("idle_timeout")
	if got := c.GetEvictReason(); got != "idle_timeout" {
		t.Errorf("got evict reason %q, want idle_timeout", got)
	}
}
