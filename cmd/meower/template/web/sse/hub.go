// Package sse provides an in-process server-sent-events hub: browsers subscribe
// to named channels, application code publishes events to a channel, and the
// hub fans each event out to every connection subscribed to it. It is the
// real-time backbone for HTMX-driven live updates.
//
// The hub is in-process: it fans out within a single web server instance. To
// fan out across multiple instances, put a shared bus (e.g. Redis pub/sub) in
// front of Broadcast — subscribe each instance to the bus and call Broadcast
// from the subscription. That's left out of the base to keep it dependency-free.
package sse

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/charmbracelet/log"
)

// reaperInterval controls how often the reaper checks for stale connections.
const reaperInterval = 60 * time.Second

// Connection lifecycle constants.
const (
	// ClientBufferSize is the per-connection outgoing event buffer. Events are
	// dropped (not blocked) when a slow client's buffer fills.
	ClientBufferSize = 64
	// MaxConnectionLifetime caps how long a single SSE connection may stay open.
	MaxConnectionLifetime = 24 * time.Hour
	// ConnectionIdleTimeout evicts connections with no delivered activity.
	ConnectionIdleTimeout = 5 * time.Minute
)

// SSEEvent is what gets written to the browser. Data is typically an HTML
// fragment (for an HTMX swap) or JSON.
type SSEEvent struct {
	Event string // SSE event type (e.g. "meow-created")
	ID    string // event ID
	Data  string // HTML fragment or JSON
}

// Hub manages SSE connections and fans events out to subscribed clients.
type Hub struct {
	mu sync.RWMutex
	// channels maps a channel key → the set of clients subscribed to it.
	// Channel keys are arbitrary strings chosen by the application, e.g.
	// "meows" for a global feed or "user:<id>" for a per-user channel.
	channels map[string]map[*Client]struct{}

	// Stats
	totalConnections atomic.Int64
	totalEvents      atomic.Int64
	droppedEvents    atomic.Int64
}

// Client represents a single browser SSE connection.
type Client struct {
	UserID      string        // optional: identifies the connected user, for logging
	IPAddress   string        // optional: client IP, for logging
	Channels    []string      // subscribed channel keys
	Events      chan SSEEvent // buffered outgoing events
	Done        chan struct{} // closed when the client disconnects
	ConnectedAt time.Time

	// closed is set before closing Events so safeSend can avoid sending on a
	// closed channel without relying on recover().
	closed atomic.Bool

	// lastActivityNano is the Unix-nanosecond time of the last delivered event
	// or heartbeat. Updated atomically so Broadcast/handlers don't need the hub
	// mutex to record liveness.
	lastActivityNano atomic.Int64

	// evictReason is set by the reaper before it closes Done so the handler's
	// deferred Unregister reads the correct reason. atomic.Pointer guarantees
	// cross-goroutine visibility.
	evictReason atomic.Pointer[string]

	// once-guards make Unregister/CloseDone safe to call from both the reaper
	// and the handler's deferred cleanup without double-close panics or counter
	// corruption.
	closeEventsOnce sync.Once
	closeDoneOnce   sync.Once
	unregisterOnce  sync.Once
}

// CloseDone idempotently closes the Done channel.
func (c *Client) CloseDone() {
	c.closeDoneOnce.Do(func() { close(c.Done) })
}

// LastActivity returns the time of the most recent delivery or heartbeat.
func (c *Client) LastActivity() time.Time {
	nano := c.lastActivityNano.Load()
	if nano == 0 {
		return c.ConnectedAt
	}
	return time.Unix(0, nano)
}

// TouchActivity records now as the last activity time. Call it from the handler
// after a successful write+flush to the browser, so the idle reaper reflects
// real browser liveness rather than buffer depth.
func (c *Client) TouchActivity() {
	c.lastActivityNano.Store(time.Now().UnixNano())
}

// SetEvictReason atomically stores the eviction reason.
func (c *Client) SetEvictReason(reason string) {
	c.evictReason.Store(&reason)
}

// GetEvictReason atomically loads the eviction reason, returning "" if unset.
func (c *Client) GetEvictReason() string {
	if p := c.evictReason.Load(); p != nil {
		return *p
	}
	return ""
}

// NewHub creates a new, empty SSE hub.
func NewHub() *Hub {
	return &Hub{
		channels: make(map[string]map[*Client]struct{}),
	}
}

// Register adds a client to the hub for each of its subscribed channels.
func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, ch := range client.Channels {
		if h.channels[ch] == nil {
			h.channels[ch] = make(map[*Client]struct{})
		}
		h.channels[ch][client] = struct{}{}
	}

	h.totalConnections.Add(1)

	log.Info("SSE client connected",
		"user_id", client.UserID,
		"ip_address", client.IPAddress,
		"channels", client.Channels,
		"total_connections", h.totalConnections.Load())
}

// Unregister removes a client from the hub and closes its Events channel.
// reason is a short string for the audit log (e.g. "client_disconnect",
// "lifetime_exceeded", "idle_timeout"). Safe to call multiple times.
func (h *Hub) Unregister(client *Client, reason string) {
	client.unregisterOnce.Do(func() {
		h.mu.Lock()
		defer h.mu.Unlock()

		for _, ch := range client.Channels {
			if clients, ok := h.channels[ch]; ok {
				delete(clients, client)
				if len(clients) == 0 {
					delete(h.channels, ch)
				}
			}
		}

		h.totalConnections.Add(-1)

		client.closed.Store(true)
		client.closeEventsOnce.Do(func() { close(client.Events) })

		log.Info("SSE client disconnected",
			"user_id", client.UserID,
			"reason", reason,
			"duration", time.Since(client.ConnectedAt).Truncate(time.Second).String(),
			"total_connections", h.totalConnections.Load())
	})
}

// Broadcast sends an event to every client subscribed to the given channel.
// It drops the event for any client whose buffer is full rather than blocking.
func (h *Hub) Broadcast(channel string, event SSEEvent) {
	h.mu.RLock()
	clients := h.channels[channel]
	// Copy the set under read lock so we don't hold the lock during sends.
	targets := make([]*Client, 0, len(clients))
	for c := range clients {
		targets = append(targets, c)
	}
	h.mu.RUnlock()

	for _, client := range targets {
		if client.closed.Load() {
			continue
		}
		if safeSend(client.Events, event) {
			h.totalEvents.Add(1)
			// We intentionally do NOT TouchActivity here: buffering an event is
			// not proof the browser received it. Only the handler touches
			// activity after a successful write+flush.
		} else {
			h.droppedEvents.Add(1)
			log.Warn("SSE event dropped (buffer full or client closed)",
				"user_id", client.UserID,
				"channel", channel)
		}
	}
}

// safeSend attempts a non-blocking send on ch, returning true if delivered and
// false if the buffer was full. The caller must check client.closed first.
func safeSend(ch chan SSEEvent, ev SSEEvent) bool {
	select {
	case ch <- ev:
		return true
	default:
		return false
	}
}

// Stats returns current connection and event-delivery counts.
func (h *Hub) Stats() (connections int64, events int64) {
	return h.totalConnections.Load(), h.totalEvents.Load()
}

// HealthInfo is a snapshot of hub health metrics, suitable for a /health
// endpoint or monitoring scrape.
type HealthInfo struct {
	ActiveConnections int64 `json:"active_connections"`
	TotalEvents       int64 `json:"total_events"`
	DroppedEvents     int64 `json:"dropped_events"`
}

// Health returns a snapshot of hub health metrics.
func (h *Hub) Health() HealthInfo {
	return HealthInfo{
		ActiveConnections: h.totalConnections.Load(),
		TotalEvents:       h.totalEvents.Load(),
		DroppedEvents:     h.droppedEvents.Load(),
	}
}

// StartReaper runs a goroutine that evicts clients exceeding
// MaxConnectionLifetime or idle past ConnectionIdleTimeout. It returns when ctx
// is cancelled. Start it once at server startup.
func (h *Hub) StartReaper(stop <-chan struct{}) {
	ticker := time.NewTicker(reaperInterval)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			h.reap()
		}
	}
}

// reap evicts clients past their lifetime or idle timeout. Expired (lifetime)
// takes priority over idle when a client matches both. Must not hold the mutex.
func (h *Hub) reap() {
	now := time.Now()

	h.mu.RLock()
	seen := make(map[*Client]struct{})
	var expired, idle []*Client
	for _, clients := range h.channels {
		for c := range clients {
			if _, ok := seen[c]; ok {
				continue
			}
			seen[c] = struct{}{}
			age := now.Sub(c.ConnectedAt)
			idleFor := now.Sub(c.LastActivity())
			if age > MaxConnectionLifetime {
				expired = append(expired, c)
			} else if idleFor > ConnectionIdleTimeout {
				idle = append(idle, c)
			}
		}
	}
	h.mu.RUnlock()

	evict := func(c *Client, reason, logMsg string, fields ...any) {
		args := append([]any{"user_id", c.UserID}, fields...)
		log.Info(logMsg, args...)

		// Set the reason before closing Done so the handler's deferred
		// Unregister reads the right value, then Unregister promptly (the
		// once-guard makes the handler's later call a no-op).
		c.SetEvictReason(reason)
		c.CloseDone()
		h.Unregister(c, reason)
	}

	for _, c := range expired {
		evict(c, "lifetime_exceeded", "SSE reaper: closing connection (lifetime exceeded)",
			"age", time.Since(c.ConnectedAt).Truncate(time.Second).String(),
			"limit", MaxConnectionLifetime.String())
	}
	for _, c := range idle {
		evict(c, "idle_timeout", "SSE reaper: closing connection (idle timeout)",
			"idle_for", time.Since(c.LastActivity()).Truncate(time.Second).String(),
			"limit", ConnectionIdleTimeout.String())
	}
}
