package handlers

import (
	"bufio"
	"fmt"
	"strings"
	"time"

	"TEMPLATE_MODULE_PATH/web/sse"

	"github.com/charmbracelet/log"
	"github.com/gofiber/fiber/v2"
)

// defaultHeartbeatInterval controls how often a keep-alive comment is written
// to an idle SSE connection so proxies don't close it.
const defaultHeartbeatInterval = 30 * time.Second

// Channel-subscription limits to prevent abuse via unbounded subscriptions.
const (
	maxChannelIDLen      = 64
	maxChannelsPerClient = 10
)

// SSEHandler serves the SSE stream and health endpoints. Construct it with the
// shared hub: handlers.SSEHandler{App: app, Hub: hub}.
type SSEHandler struct {
	App *App
	Hub *sse.Hub

	// HeartbeatInterval overrides the heartbeat frequency. Zero uses the
	// default. Exposed mainly for testing.
	HeartbeatInterval time.Duration
}

func (h *SSEHandler) heartbeatInterval() time.Duration {
	if h.HeartbeatInterval > 0 {
		return h.HeartbeatInterval
	}
	return defaultHeartbeatInterval
}

// Stream handles GET /events/stream. The browser opens an EventSource here and
// subscribes to the channels named in the ?channels=a,b query parameter. The
// connection stays open; the hub fans matching events to it until the client
// disconnects or the reaper evicts it.
//
// Channels are accepted as-is (after length/count validation). When you add
// auth, restrict the accepted channels to those the authenticated user may
// subscribe to — e.g. only allow a "user:<id>" channel for the logged-in user.
func (h *SSEHandler) Stream(c *fiber.Ctx) error {
	channels := validateChannels(strings.Split(c.Query("channels"), ","))
	if len(channels) == 0 {
		return c.Status(fiber.StatusBadRequest).SendString("no valid channels requested")
	}

	client := &sse.Client{
		IPAddress:   c.IP(),
		Channels:    channels,
		Events:      make(chan sse.SSEEvent, sse.ClientBufferSize),
		Done:        make(chan struct{}),
		ConnectedAt: time.Now(),
	}
	h.Hub.Register(client)

	// SSE needs the connection held open and flushed incrementally, which
	// Fiber's normal response pipeline can't do — hand the raw writer to
	// SetBodyStreamWriter and format the events ourselves.
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no") // disable proxy buffering (nginx)

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		// Defers run LIFO; this order guarantees CloseDone → Unregister → log.
		defer func() {
			reason := client.GetEvictReason()
			if reason == "" {
				reason = "client_disconnect"
			}
			h.Hub.Unregister(client, reason)
		}()
		defer client.CloseDone()

		// Initial comment confirms the stream is alive.
		if _, err := fmt.Fprint(w, ":ok\n\n"); err != nil {
			return
		}
		if err := w.Flush(); err != nil {
			return
		}

		heartbeat := time.NewTicker(h.heartbeatInterval())
		defer heartbeat.Stop()

		for {
			select {
			case <-client.Done:
				return // reaper evicted us (lifetime/idle)
			case event, ok := <-client.Events:
				if !ok {
					return // hub closed the channel
				}
				if err := writeSSEEvent(w, event); err != nil {
					log.Debug("SSE write error, closing connection", "error", err)
					return
				}
				client.TouchActivity()
			case <-heartbeat.C:
				if _, err := fmt.Fprint(w, ":heartbeat\n\n"); err != nil {
					return
				}
				if err := w.Flush(); err != nil {
					return
				}
				client.TouchActivity()
			}
		}
	})

	return nil
}

// Health handles GET /events/health, returning hub metrics as JSON.
func (h *SSEHandler) Health(c *fiber.Ctx) error {
	return c.JSON(h.Hub.Health())
}

// validateChannels trims, de-dupes, and caps the requested channel list.
func validateChannels(raw []string) []string {
	seen := make(map[string]struct{}, len(raw))
	valid := make([]string, 0, len(raw))
	for _, ch := range raw {
		ch = strings.TrimSpace(ch)
		if ch == "" || len(ch) > maxChannelIDLen {
			continue
		}
		if _, dup := seen[ch]; dup {
			continue
		}
		seen[ch] = struct{}{}
		valid = append(valid, ch)
		if len(valid) >= maxChannelsPerClient {
			break
		}
	}
	return valid
}

// writeSSEEvent formats and flushes a single SSE message to w.
func writeSSEEvent(w *bufio.Writer, event sse.SSEEvent) error {
	if event.Event != "" {
		if _, err := fmt.Fprintf(w, "event: %s\n", sanitizeSSEField(event.Event)); err != nil {
			return err
		}
	}
	if event.ID != "" {
		if _, err := fmt.Fprintf(w, "id: %s\n", sanitizeSSEField(event.ID)); err != nil {
			return err
		}
	}
	// data: lines must not contain raw newlines — use the SSE multi-line form.
	data := strings.ReplaceAll(event.Data, "\n", "\ndata: ")
	if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
		return err
	}
	return w.Flush()
}

// sanitizeSSEField strips CR/LF from SSE field values (event, id), which the
// SSE spec forbids.
func sanitizeSSEField(s string) string {
	s = strings.ReplaceAll(s, "\r", "")
	return strings.ReplaceAll(s, "\n", "")
}
