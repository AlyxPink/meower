package urls

// SSEStream is the server-sent events endpoint. Clients subscribe to one or
// more channels via a query parameter, e.g. ?channels=meows,user:{id}. The
// channel naming scheme is up to the application; the SSE hub (web/sse) fans
// out published events to whichever connections subscribed to a channel.
type SSEStream struct{}

func (SSEStream) URL() string     { return "/events/stream" }
func (SSEStream) Pattern() string { return "/events/stream" }
func (SSEStream) Name() string    { return "events.stream" }

// SSEHealth exposes SSE hub health metrics (connection and channel counts) for
// monitoring and debugging.
type SSEHealth struct{}

func (SSEHealth) URL() string     { return "/events/health" }
func (SSEHealth) Pattern() string { return "/events/health" }
func (SSEHealth) Name() string    { return "events.health" }
