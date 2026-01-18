package grpc

import (
	"context"
	"runtime/debug"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"microservice-template/pkg/logger"
)

// loggingInterceptor logs gRPC requests at Info level and errors at Error level.
func loggingInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		method := "unknown"
		if info != nil {
			method = info.FullMethod
		}

		logger.Log().WithField("method", method).Info("grpc request")

		resp, err := handler(ctx, req)

		if err != nil {
			logger.Log().WithField("method", method).Errorf("grpc error: %v", err)
		}

		return resp, err
	}
}

// recoveryInterceptor recovers from panics in gRPC handlers and returns an Internal error.
func recoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		_ *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Log().Errorf("grpc panic recovered: %v\n%s", r, debug.Stack())
				err = status.Errorf(codes.Internal, "internal server error: %v", r)
			}
		}()

		return handler(ctx, req)
	}
}
