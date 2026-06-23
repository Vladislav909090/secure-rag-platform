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

	notReady := NewMockRAGUsecase(t)
	notReady.EXPECT().Ready().Return(false)
	err = (&Server{uc: notReady}).requireUC()
	require.Error(t, err)
	assert.Equal(t, codes.Unavailable, status.Code(err))

	ready := NewMockRAGUsecase(t)
	ready.EXPECT().Ready().Return(true)
	err = (&Server{uc: ready}).requireUC()
	require.NoError(t, err)
}

func TestNewServer(t *testing.T) {
	t.Parallel()

	server := NewServer(nil)
	require.NotNil(t, server)
	assert.Nil(t, server.uc)
}
