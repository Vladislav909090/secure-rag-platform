package grpc

import (
	"errors"
	"log/slog"
	"strings"

	"secure-rag-platform/services/knowledge/internal/usecase"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, usecase.ErrDocumentNotFound),
		errors.Is(err, usecase.ErrFileNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, usecase.ErrDocumentDeleted),
		errors.Is(err, usecase.ErrDocumentNotDeleted):
		return status.Error(codes.Aborted, err.Error())
	case errors.Is(err, usecase.ErrFileTooLarge):
		return status.Error(codes.ResourceExhausted, err.Error())
	case errors.Is(err, usecase.ErrInvalidRequest):
		return status.Error(codes.InvalidArgument, err.Error())
	case strings.Contains(err.Error(), "upload to storage"):
		return status.Error(codes.Internal, err.Error())
	default:
		slog.Error("внутренняя ошибка gRPC",
			"component", "knowledge.grpc",
			"error", err,
		)

		return status.Error(codes.Internal, "internal error")
	}
}
