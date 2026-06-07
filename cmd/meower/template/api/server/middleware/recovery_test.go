package middleware

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRecoveryUnaryInterceptor(t *testing.T) {
	interceptor := RecoveryUnaryInterceptor()
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

	t.Run("normal handler passes through", func(t *testing.T) {
		handler := func(ctx context.Context, req any) (any, error) {
			return "ok", nil
		}
		resp, err := interceptor(context.Background(), nil, info, handler)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp != "ok" {
			t.Errorf("got resp %v, want \"ok\"", resp)
		}
	})

	t.Run("handler error passes through", func(t *testing.T) {
		handler := func(ctx context.Context, req any) (any, error) {
			return nil, status.Error(codes.NotFound, "not found")
		}
		resp, err := interceptor(context.Background(), nil, info, handler)
		if resp != nil {
			t.Errorf("got resp %v, want nil", resp)
		}
		if status.Code(err) != codes.NotFound {
			t.Errorf("got code %v, want NotFound", status.Code(err))
		}
	})

	t.Run("panic is recovered and returns Internal", func(t *testing.T) {
		handler := func(ctx context.Context, req any) (any, error) {
			panic("something went wrong")
		}
		resp, err := interceptor(context.Background(), nil, info, handler)
		if resp != nil {
			t.Errorf("got resp %v, want nil", resp)
		}
		if status.Code(err) != codes.Internal {
			t.Errorf("got code %v, want Internal", status.Code(err))
		}
		if msg := status.Convert(err).Message(); msg != "internal server error" {
			t.Errorf("got message %q, want \"internal server error\"", msg)
		}
	})
}

func TestRecoveryStreamInterceptor(t *testing.T) {
	interceptor := RecoveryStreamInterceptor()
	info := &grpc.StreamServerInfo{FullMethod: "/test.Service/StreamMethod"}

	t.Run("normal handler passes through", func(t *testing.T) {
		handler := func(srv any, stream grpc.ServerStream) error {
			return nil
		}
		if err := interceptor(nil, nil, info, handler); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("panic is recovered and returns Internal", func(t *testing.T) {
		handler := func(srv any, stream grpc.ServerStream) error {
			panic("stream panic")
		}
		err := interceptor(nil, nil, info, handler)
		if status.Code(err) != codes.Internal {
			t.Errorf("got code %v, want Internal", status.Code(err))
		}
		if msg := status.Convert(err).Message(); msg != "internal server error" {
			t.Errorf("got message %q, want \"internal server error\"", msg)
		}
	})
}
