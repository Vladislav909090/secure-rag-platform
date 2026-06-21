package usecase

import (
	"testing"

	iamv1 "secure-rag-platform/api/gen/go/iam/v1"

	"google.golang.org/protobuf/types/known/structpb"
)

func TestGatewayIAMProtoConversions(t *testing.T) {
	attrs, err := structpb.NewStruct(map[string]any{"team": "platform"})
	if err != nil {
		t.Fatalf("NewStruct() error = %v", err)
	}
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
	if user.ID != "u1" || user.Attributes["team"] != "platform" || user.CtxVer != 3 {
		t.Fatalf("unexpected user: %#v", user)
	}
	if empty := userFromProto(nil); empty.ID != "" {
		t.Fatalf("nil user should produce empty user, got %#v", empty)
	}

	roles := rolesFromProto([]*iamv1.Role{
		nil,
		{Id: 1, Code: roleUser, Name: "User", Description: "default", CreatedAt: "created"},
	})
	if len(roles) != 1 || roles[0].Code != roleUser || roles[0].Name != "User" {
		t.Fatalf("unexpected roles: %#v", roles)
	}
}
