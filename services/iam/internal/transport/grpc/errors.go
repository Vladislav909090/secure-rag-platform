package grpc

import (
	"errors"

	"secure-rag-platform/services/iam/internal/usecase"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, usecase.ErrInvalidArgument):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, usecase.ErrRateLimited):
		return status.Error(codes.ResourceExhausted, err.Error())
	case errors.Is(err, usecase.ErrForbidden):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, usecase.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, usecase.ErrUserExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, usecase.ErrInactiveUser):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, usecase.ErrInvalidToken),
		errors.Is(err, usecase.ErrInvalidCredentials),
		errors.Is(err, usecase.ErrUnauthorized),
		errors.Is(err, usecase.ErrSessionExpired),
		errors.Is(err, usecase.ErrSessionRevoked):
		return status.Error(codes.Unauthenticated, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
