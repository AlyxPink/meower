package middleware

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

// Span attribute keys set by the trace-enrichment interceptors.
const (
	attrRPCMethod = "rpc.method"
)

// EnrichTraceUnary returns a unary server interceptor that decorates the active
// OTel span with request-scoped attributes (currently the full RPC method).
//
// Extension point: once the auth scaffold is enabled, pull the authenticated
// user out of ctx here and add an `enduser.id` attribute so traces can be
// filtered by user. Keep this interceptor *after* the auth interceptor in the
// chain so the user is already on the context.
func EnrichTraceUnary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		enrichSpan(ctx, info.FullMethod)
		return handler(ctx, req)
	}
}

// EnrichTraceStream is the stream-RPC counterpart of EnrichTraceUnary.
func EnrichTraceStream() grpc.StreamServerInterceptor {
	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		enrichSpan(ss.Context(), info.FullMethod)
		return handler(srv, ss)
	}
}

// enrichSpan sets attributes on the active span. It is a no-op when the span is
// not recording.
func enrichSpan(ctx context.Context, fullMethod string) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}
	if fullMethod != "" {
		span.SetAttributes(attribute.String(attrRPCMethod, fullMethod))
	}
}
