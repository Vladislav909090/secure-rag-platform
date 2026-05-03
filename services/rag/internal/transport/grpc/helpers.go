package grpc

import (
	"errors"
	"log"

	"secure-rag-platform/services/rag/internal/usecase"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) requireUC() error {
	if s.uc == nil || !s.uc.Ready() {
		return status.Error(codes.Unavailable, "service not configured")
	}
	return nil
}

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, usecase.ErrNotConfigured):
		return status.Error(codes.Unavailable, err.Error())
	case errors.Is(err, usecase.ErrInvalidRequest):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, usecase.ErrNoContexts):
		return status.Error(codes.NotFound, err.Error())
	default:
		log.Printf("[rag.grpc] внутренняя ошибка: %v", err)
		return status.Error(codes.Internal, "internal error")
	}
}
