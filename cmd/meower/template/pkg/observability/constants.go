package observability

// Service names identify each component in traces and metrics. They are used as
// the OpenTelemetry service.name resource attribute, so a trace spanning web →
// api shows which component each span came from.
const (
	ServiceNameAPI = "TEMPLATE_PROJECT_NAME-api"
	ServiceNameWeb = "TEMPLATE_PROJECT_NAME-web"
)

// Common span attribute keys. Use these constants rather than raw strings so
// attribute names stay consistent across the codebase (and so a rename is one
// edit, not a grep-and-pray).
const (
	AttrUserID    = "user.id"
	AttrRequestID = "request.id"
	AttrMethod    = "rpc.method"
	AttrService   = "rpc.service"
)
