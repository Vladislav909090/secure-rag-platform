package grpc

import (
	"errors"
	"testing"

	"secure-rag-platform/services/iam/internal/usecase"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestIAMToGRPCError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		err  error
		code codes.Code
	}{
		{usecase.ErrInvalidArgument, codes.InvalidArgument},
		{usecase.ErrRateLimited, codes.ResourceExhausted},
		{usecase.ErrForbidden, codes.PermissionDenied},
		{usecase.ErrNotFound, codes.NotFound},
		{usecase.ErrUserExists, codes.AlreadyExists},
		{usecase.ErrInactiveUser, codes.FailedPrecondition},
		{usecase.ErrUnauthorized, codes.Unauthenticated},
		{errors.New("boom"), codes.Internal},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.code.String(), func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.code, status.Code(toGRPCError(tt.err)))
		})
	}
}
