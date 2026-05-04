package grpc

import (
	"errors"

	"secure-rag-platform/services/ai-inference/internal/usecase"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, usecase.ErrInvalidArgument):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, usecase.ErrAliasNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, usecase.ErrAliasTaskMismatch):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, usecase.ErrProviderNotConfigured):
		return status.Error(codes.Unavailable, err.Error())
	case errors.Is(err, usecase.ErrProviderFailed):
		return status.Error(codes.Unavailable, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
