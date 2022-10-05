package middleware

import (
	"context"

	"google.golang.org/grpc/codes"

	"github.com/integration-system/isp-kit/grpc"
	"github.com/integration-system/isp-kit/grpc/endpoint"
	"github.com/integration-system/isp-kit/grpc/isp"
	"github.com/integration-system/isp-kit/log"
	"google.golang.org/grpc/status"
)

func ErrorHandler(logger log.Logger) endpoint.Middleware {
	return func(next grpc.HandlerFunc) grpc.HandlerFunc {
		return func(ctx context.Context, message *isp.Message) (*isp.Message, error) {
			result, err := next(ctx, message)
			if err == nil {
				return result, nil
			}
			logger.Error(ctx, err)
			_, ok := status.FromError(err)
			if ok {
				return result, err
			}
			// hide error details to prevent potential security leaks
			return result, status.Error(codes.Internal, "Internal service error")
		}
	}
}
