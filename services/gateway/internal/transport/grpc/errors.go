package grpc

import (
	"errors"
	"log/slog"

	"secure-rag-platform/services/gateway/internal/usecase"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, usecase.ErrNotConfigured),
		errors.Is(err, usecase.ErrPolicyRequired),
		errors.Is(err, usecase.ErrPolicyUnavailable):
		return status.Error(codes.Unavailable, err.Error())
	case errors.Is(err, usecase.ErrInvalidRequest):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, usecase.ErrUnauthorized):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, usecase.ErrForbidden):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, usecase.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		slog.Error("внутренняя ошибка gRPC",
			"component", "gateway.grpc",
			"error", err,
		)

		return status.Error(codes.Internal, "internal error")
	}
}
