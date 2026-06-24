package grpc

import (
	"testing"
	"time"

	"secure-rag-platform/services/iam/internal/model"
	"secure-rag-platform/services/iam/internal/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestIAMProtoConversions(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 6, 21, 12, 30, 0, 0, time.UTC)
	revokedAt := now.Add(time.Hour)

	subject := subjectToProto(&model.SubjectContext{
		UserID:     "u1",
		Login:      "alice",
		IsActive:   true,
		Roles:      []string{usecase.RoleUser},
		Attributes: map[string]any{"department": "search"},
		CtxVer:     7,
	})
	assert.Equal(t, "u1", subject.GetUserId())
	assert.Equal(t, "search", subject.GetAttributes().AsMap()["department"])

	user := userToProto(&model.UserView{
		ID:         "u1",
		Login:      "alice",
		IsActive:   true,
		CtxVer:     7,
		Roles:      []string{usecase.RoleUser},
		Attributes: map[string]any{"level": 2},
		CreatedAt:  now,
		UpdatedAt:  now.Add(time.Minute),
	})
	assert.Equal(t, "u1", user.GetId())
	assert.Equal(t, now.Format(time.RFC3339), user.GetCreatedAt())

	role := roleToProto(&model.Role{
		ID:          1,
		Code:        usecase.RoleUser,
		Name:        "User",
		Description: "Default user",
		CreatedAt:   now,
	})
	assert.Equal(t, usecase.RoleUser, role.GetCode())
	assert.Equal(t, now.Format(time.RFC3339), role.GetCreatedAt())

	roles := rolesToProto([]*model.Role{{ID: 1, Code: usecase.RoleUser, CreatedAt: now}, nil})
	require.Len(t, roles, 2)
	assert.Equal(t, usecase.RoleUser, roles[0].GetCode())
	assert.Nil(t, roles[1])

	session := sessionToProto(&model.UserSession{
		ID:        "s1",
		UserID:    "u1",
		ExpiresAt: now.Add(24 * time.Hour),
		RevokedAt: &revokedAt,
		CreatedAt: now,
		UpdatedAt: now.Add(time.Minute),
	})
	assert.Equal(t, "s1", session.GetId())
	assert.Equal(t, revokedAt.Format(time.RFC3339), session.GetRevokedAt())
}

func TestIAMProtoConversionsNilAndFallbacks(t *testing.T) {
	t.Parallel()

	assert.Nil(t, subjectToProto(nil))
	assert.Nil(t, userToProto(nil))
	assert.Nil(t, roleToProto(nil))
	assert.Nil(t, sessionToProto(nil))

	empty := mapToStruct(nil)
	assert.Empty(t, empty.GetFields())

	fallback := mapToStruct(map[string]any{"bad": func() {}})
	assert.Empty(t, fallback.GetFields())

	assert.Empty(t, structToMap(nil))

	value, err := structpb.NewStruct(map[string]any{"a": "b"})
	require.NoError(t, err)
	assert.Equal(t, "b", structToMap(value)["a"])
}
