package grpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRAGServerRequireUC(t *testing.T) {
	t.Parallel()

	err := (&Server{}).requireUC()
	require.Error(t, err)
	assert.Equal(t, codes.Unavailable, status.Code(err))

	err = (&Server{uc: &mockRAGUsecase{t: t, ready: func() bool { return false }}}).requireUC()
	require.Error(t, err)
	assert.Equal(t, codes.Unavailable, status.Code(err))

	err = (&Server{uc: &mockRAGUsecase{t: t}}).requireUC()
	require.NoError(t, err)
}

func TestNewServer(t *testing.T) {
	t.Parallel()

	server := NewServer(nil)
	require.NotNil(t, server)
	assert.Nil(t, server.uc)
}
