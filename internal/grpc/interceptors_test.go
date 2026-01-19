package grpc

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestLoggingInterceptor_Success(t *testing.T) {
	t.Parallel()

	interceptor := loggingInterceptor()

	handler := func(_ context.Context, _ interface{}) (interface{}, error) {
		return "ok", nil
	}

	resp, err := interceptor(context.Background(), "req", &grpc.UnaryServerInfo{FullMethod: "/test.Method"}, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "ok" {
		t.Fatalf("unexpected response: %v", resp)
	}
}

func TestLoggingInterceptor_Error(t *testing.T) {
	t.Parallel()

	interceptor := loggingInterceptor()

	handlerErr := errors.New("boom")
	handler := func(_ context.Context, _ interface{}) (interface{}, error) {
		return nil, handlerErr
	}

	resp, err := interceptor(context.Background(), "req", &grpc.UnaryServerInfo{FullMethod: "/test.Method"}, handler)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, handlerErr) {
		t.Fatalf("expected wrapped error, got %v", err)
	}
	if resp != nil {
		t.Fatalf("expected nil response, got %v", resp)
	}
}

func TestRecoveryInterceptor_NoPanic(t *testing.T) {
	t.Parallel()

	interceptor := recoveryInterceptor()

	handler := func(_ context.Context, _ interface{}) (interface{}, error) {
		return "ok", nil
	}

	resp, err := interceptor(context.Background(), "req", &grpc.UnaryServerInfo{FullMethod: "/test.Method"}, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "ok" {
		t.Fatalf("unexpected response: %v", resp)
	}
}

func TestRecoveryInterceptor_Panic(t *testing.T) {
	t.Parallel()

	interceptor := recoveryInterceptor()

	handler := func(_ context.Context, _ interface{}) (interface{}, error) {
		panic("boom")
	}

	_, err := interceptor(context.Background(), "req", &grpc.UnaryServerInfo{FullMethod: "/test.Method"}, handler)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected grpc status error, got %v", err)
	}

	if st.Code() != codes.Internal {
		t.Fatalf("expected Internal code, got %v", st.Code())
	}

	if st.Message() == "" {
		t.Fatalf("expected non-empty error message")
	}
}
