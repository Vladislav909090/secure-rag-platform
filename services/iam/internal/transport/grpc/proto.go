package grpc

import (
	"time"

	pb "secure-rag-platform/api/gen/go/iam/v1"
	"secure-rag-platform/services/iam/internal/model"

	"google.golang.org/protobuf/types/known/structpb"
)

func structToMap(value *structpb.Struct) map[string]any {
	if value == nil {
		return map[string]any{}
	}

	return value.AsMap()
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

func subjectToProto(subject *model.SubjectContext) *pb.SubjectContext {
	if subject == nil {
		return nil
	}

	return &pb.SubjectContext{
		UserId:     subject.UserID,
		Login:      subject.Login,
		IsActive:   subject.IsActive,
		Roles:      append([]string(nil), subject.Roles...),
		Attributes: mapToStruct(subject.Attributes),
		CtxVer:     subject.CtxVer,
	}
}

func userToProto(view *model.UserView) *pb.User {
	if view == nil {
		return nil
	}

	return &pb.User{
		Id:         view.ID,
		Login:      view.Login,
		IsActive:   view.IsActive,
		CtxVer:     view.CtxVer,
		Roles:      append([]string(nil), view.Roles...),
		Attributes: mapToStruct(view.Attributes),
		CreatedAt:  view.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  view.UpdatedAt.Format(time.RFC3339),
	}
}

func roleToProto(role *model.Role) *pb.Role {
	if role == nil {
		return nil
	}

	return &pb.Role{
		Id:          role.ID,
		Code:        role.Code,
		Name:        role.Name,
		Description: role.Description,
		CreatedAt:   role.CreatedAt.Format(time.RFC3339),
	}
}

func rolesToProto(roles []*model.Role) []*pb.Role {
	out := make([]*pb.Role, 0, len(roles))
	for _, role := range roles {
		out = append(out, roleToProto(role))
	}

	return out
}

func sessionToProto(session *model.UserSession) *pb.UserSession {
	if session == nil {
		return nil
	}

	out := &pb.UserSession{
		Id:        session.ID,
		UserId:    session.UserID,
		ExpiresAt: session.ExpiresAt.Format(time.RFC3339),
		CreatedAt: session.CreatedAt.Format(time.RFC3339),
		UpdatedAt: session.UpdatedAt.Format(time.RFC3339),
	}
	if session.RevokedAt != nil {
		out.RevokedAt = session.RevokedAt.Format(time.RFC3339)
	}

	return out
}
