package grpc

import (
	"context"
	"errors"
	"strings"
	"time"

	pb "secure-rag-platform/api/gen/go/iam/v1"
	"secure-rag-platform/services/iam/internal/model"
	"secure-rag-platform/services/iam/internal/usecase"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

func requireUC(uc *usecase.IAMUsecase) error {
	if uc == nil {
		return status.Error(codes.Unavailable, "service not configured")
	}

	return nil
}

func authenticate(uc *usecase.IAMUsecase, ctx context.Context) (*usecase.Principal, *model.SubjectContext, error) {
	if err := requireUC(uc); err != nil {
		return nil, nil, err
	}

	accessToken, err := extractBearerToken(ctx)
	if err != nil {
		return nil, nil, err
	}

	principal, subject, err := uc.AuthenticateAccessToken(ctx, accessToken)
	if err != nil {
		return nil, nil, err
	}

	return principal, subject, nil
}

func hasRole(roles []string, roleCode string) bool {
	for _, role := range roles {
		if role == roleCode {
			return true
		}
	}

	return false
}

func isAdmin(principal *usecase.Principal) bool {
	if principal == nil {
		return false
	}

	return hasRole(principal.Roles, usecase.RoleAccessAdmin) || hasRole(principal.Roles, usecase.RoleSuperAdmin)
}

func canAccessUser(principal *usecase.Principal, userID string) bool {
	if principal == nil {
		return false
	}
	if principal.UserID == userID {
		return true
	}

	return isAdmin(principal)
}

func extractBearerToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", usecase.ErrUnauthorized
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		values = md.Get("Authorization")
	}
	if len(values) == 0 {
		return "", usecase.ErrUnauthorized
	}

	value := strings.TrimSpace(values[0])
	if value == "" {
		return "", usecase.ErrUnauthorized
	}

	parts := strings.SplitN(value, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", usecase.ErrUnauthorized
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", usecase.ErrUnauthorized
	}

	return token, nil
}

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
