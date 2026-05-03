package grpc

import (
	"errors"
	"log"

	"secure-rag-platform/services/gateway/internal/usecase"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

func (s *Server) requireUC() error {
	if s.uc == nil || !s.uc.Ready() {
		return status.Error(codes.Unavailable, "service not configured")
	}
	return nil
}

func mapToStruct(value map[string]any) *structpb.Struct {
	if value == nil {
		value = map[string]any{}
	}
	out, err := structpb.NewStruct(value)
	if err != nil {
		return &structpb.Struct{Fields: map[string]*structpb.Value{}}
	}
	return out
}

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, usecase.ErrNotConfigured):
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
		log.Printf("[gateway] internal error: %v", err)
		return status.Error(codes.Internal, "internal error")
	}
}
