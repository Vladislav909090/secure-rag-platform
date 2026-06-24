package grpc

import (
	"errors"
	"log/slog"

	"secure-rag-platform/services/rag/internal/usecase"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, usecase.ErrNotConfigured):
		return status.Error(codes.Unavailable, err.Error())
	case errors.Is(err, usecase.ErrInvalidRequest):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, usecase.ErrNoContexts):
		return status.Error(codes.NotFound, err.Error())
	default:
		slog.Error("внутренняя ошибка gRPC",
			"component", "rag.grpc",
			"error", err,
		)

		return status.Error(codes.Internal, "internal error")
	}
}
