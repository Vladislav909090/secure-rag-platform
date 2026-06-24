package usecase

import (
	"context"
	"errors"
	"testing"

	iamv1 "secure-rag-platform/api/gen/go/iam/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestGatewayIAMProtoConversions(t *testing.T) {
	t.Parallel()

	attrs, err := structpb.NewStruct(map[string]any{"team": "platform"})
	require.NoError(t, err)
	user := userFromProto(&iamv1.User{
		Id:         "u1",
		Login:      "alice",
		IsActive:   true,
		CtxVer:     3,
		Roles:      []string{roleUser},
		Attributes: attrs,
		CreatedAt:  "created",
		UpdatedAt:  "updated",
	})
	assert.Equal(t, "u1", user.ID)
	assert.Equal(t, "platform", user.Attributes["team"])
	assert.Equal(t, int64(3), user.CtxVer)
	assert.Empty(t, userFromProto(nil).ID)

	roles := rolesFromProto([]*iamv1.Role{
		nil,
		{Id: 1, Code: roleUser, Name: "User", Description: "default", CreatedAt: "created"},
	})
	require.Len(t, roles, 1)
	assert.Equal(t, roleUser, roles[0].Code)
	assert.Equal(t, "User", roles[0].Name)
}

func TestGatewayServiceCreateUserUsesIAM(t *testing.T) {
	t.Parallel()

	active := false
	svc, deps := newGatewayTestService(t)
	expectValidToken(deps, "admin-token", gatewaySubject(roleAccessAdmin))
	deps.users.EXPECT().
		CreateUser(mock.MatchedBy(hasBearer("admin-token")), mock.MatchedBy(func(req *iamv1.CreateUserRequest) bool {
			return req.GetLogin() == "alice" &&
				req.GetPassword() == "secret" &&
				req.GetIsActive() == active &&
				req.GetAttributes().AsMap()["team"] == "platform"
		})).
		Return(&iamv1.CreateUserResponse{User: &iamv1.User{Id: "u1", Login: "alice", IsActive: active}}, nil)

	got, err := svc.CreateUser(context.Background(), CreateUserRequest{
		Login:      " alice ",
		Password:   "secret",
		IsActive:   &active,
		RoleCodes:  []string{roleUser},
		Attributes: map[string]any{"team": "platform"},
	}, "admin-token")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "u1", got.ID)
	assert.False(t, got.IsActive)
}

func TestGatewayServiceCreateUserRejectsNonAdmin(t *testing.T) {
	t.Parallel()

	svc, deps := newGatewayTestService(t)
	expectValidToken(deps, "user-token", gatewaySubject(roleUser))

	got, err := svc.CreateUser(context.Background(), CreateUserRequest{Login: "alice", Password: "secret"}, "user-token")
	require.ErrorIs(t, err, ErrForbidden)
	assert.Nil(t, got)
}

func TestGatewayServiceListRolesSetRolesAndReplaceAttributes(t *testing.T) {
	t.Parallel()

	svc, deps := newGatewayTestService(t)
	expectValidToken(deps, "admin-token", gatewaySubject(roleSuperAdmin))
	expectValidToken(deps, "admin-token", gatewaySubject(roleSuperAdmin))
	expectValidToken(deps, "admin-token", gatewaySubject(roleSuperAdmin))
	deps.roles.EXPECT().
		ListRoles(mock.MatchedBy(hasBearer("admin-token")), mock.Anything).
		Return(&iamv1.ListRolesResponse{Roles: []*iamv1.Role{{Id: 1, Code: roleUser, Name: "User"}}}, nil)
	deps.roles.EXPECT().
		SetUserRoles(mock.MatchedBy(hasBearer("admin-token")), mock.MatchedBy(func(req *iamv1.SetUserRolesRequest) bool {
			return req.GetUserId() == "u1" && len(req.GetRoleCodes()) == 1 && req.GetRoleCodes()[0] == roleUser
		})).
		Return(&iamv1.SetUserRolesResponse{UserId: "u1", Roles: []*iamv1.Role{{Id: 1, Code: roleUser}}, CtxVer: 4}, nil)
	deps.attributes.EXPECT().
		ReplaceUserAttributes(mock.MatchedBy(hasBearer("admin-token")), mock.MatchedBy(func(req *iamv1.ReplaceUserAttributesRequest) bool {
			return req.GetUserId() == "u1" && req.GetAttributes().AsMap()["team"] == "rag"
		})).
		Return(&iamv1.ReplaceUserAttributesResponse{
			UserId: "u1",
			Attributes: mustStruct(map[string]any{
				"team": "rag",
			}),
			CtxVer: 5,
		}, nil)

	roles, err := svc.ListRoles(context.Background(), "admin-token")
	require.NoError(t, err)
	require.Len(t, roles, 1)
	assert.Equal(t, roleUser, roles[0].Code)

	roleResult, err := svc.SetUserRoles(context.Background(), " u1 ", []string{roleUser}, "admin-token")
	require.NoError(t, err)
	assert.Equal(t, int64(4), roleResult.CtxVer)

	attrResult, err := svc.ReplaceUserAttributes(context.Background(), " u1 ", map[string]any{"team": "rag"}, "admin-token")
	require.NoError(t, err)
	assert.Equal(t, "rag", attrResult.Attributes["team"])
	assert.Equal(t, int64(5), attrResult.CtxVer)
}

func TestGatewayServiceUpdateUserMapsIAMNotFound(t *testing.T) {
	t.Parallel()

	login := "alice2"
	svc, deps := newGatewayTestService(t)
	expectValidToken(deps, "admin-token", gatewaySubject(roleAccessAdmin))
	deps.users.EXPECT().
		UpdateUser(mock.MatchedBy(hasBearer("admin-token")), mock.Anything).
		Return((*iamv1.UpdateUserResponse)(nil), status.Error(codes.NotFound, "missing"))

	got, err := svc.UpdateUser(context.Background(), UpdateUserRequest{UserID: "u1", Login: &login}, "admin-token")
	require.ErrorIs(t, err, ErrNotFound)
	assert.Nil(t, got)
}

func TestGatewayMapIAMErrorWrapsUnknown(t *testing.T) {
	t.Parallel()

	err := errors.New("boom")
	got := mapIAMError(err, "update user")
	require.ErrorIs(t, got, err)
	assert.Contains(t, got.Error(), "update user")
}

func mustStruct(attrs map[string]any) *structpb.Struct {
	out, err := structpb.NewStruct(attrs)
	if err != nil {
		panic(err)
	}

	return out
}
