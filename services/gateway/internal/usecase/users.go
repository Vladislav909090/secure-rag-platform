package usecase

import (
	"context"
	"fmt"
	"strings"

	iamv1 "secure-rag-platform/services/iam/gen/v1"

	"google.golang.org/protobuf/types/known/structpb"
)

// CreateUser создаёт пользователя через IAM. Доступно администратору доступа или суперадмину.
func (s *Service) CreateUser(ctx context.Context, req CreateUserRequest, accessToken string) (*User, error) {
	if !s.Ready() {
		return nil, ErrNotConfigured
	}
	if s.users == nil {
		return nil, ErrNotConfigured
	}

	subject, err := s.validateAccessToken(ctx, accessToken)
	if err != nil {
		return nil, err
	}
	if err = requireAdmin(subject); err != nil {
		return nil, err
	}

	login := strings.TrimSpace(req.Login)
	if login == "" || strings.TrimSpace(req.Password) == "" {
		return nil, ErrInvalidRequest
	}

	attrs, err := structpb.NewStruct(req.Attributes)
	if err != nil {
		return nil, ErrInvalidRequest
	}

	outgoingCtx, err := outgoingAuthContext(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	createReq := &iamv1.CreateUserRequest{
		Login:      login,
		Password:   req.Password,
		RoleCodes:  append([]string(nil), req.RoleCodes...),
		Attributes: attrs,
	}
	if req.IsActive != nil {
		createReq.IsActive = req.IsActive
	}

	resp, err := s.users.CreateUser(outgoingCtx, createReq)
	if err != nil {
		return nil, mapIAMError(err, "create user")
	}

	return userFromProto(resp.GetUser()), nil
}

// UpdateUser обновляет пользователя через IAM. Доступно администратору доступа или суперадмину.
func (s *Service) UpdateUser(ctx context.Context, req UpdateUserRequest, accessToken string) (*User, error) {
	if !s.Ready() {
		return nil, ErrNotConfigured
	}
	if s.users == nil {
		return nil, ErrNotConfigured
	}

	subject, err := s.validateAccessToken(ctx, accessToken)
	if err != nil {
		return nil, err
	}
	if err = requireAdmin(subject); err != nil {
		return nil, err
	}

	userID := strings.TrimSpace(req.UserID)
	if userID == "" || (req.Login == nil && req.Password == nil && req.IsActive == nil) {
		return nil, ErrInvalidRequest
	}

	outgoingCtx, err := outgoingAuthContext(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	updateReq := &iamv1.UpdateUserRequest{
		UserId:   userID,
		Login:    req.Login,
		Password: req.Password,
		IsActive: req.IsActive,
	}
	resp, err := s.users.UpdateUser(outgoingCtx, updateReq)
	if err != nil {
		return nil, mapIAMError(err, "update user")
	}

	return userFromProto(resp.GetUser()), nil
}

// ListRoles возвращает фиксированные роли IAM через gateway.
func (s *Service) ListRoles(ctx context.Context, accessToken string) ([]Role, error) {
	if !s.Ready() {
		return nil, ErrNotConfigured
	}
	if s.roles == nil {
		return nil, ErrNotConfigured
	}

	subject, err := s.validateAccessToken(ctx, accessToken)
	if err != nil {
		return nil, err
	}
	if err = requireAdmin(subject); err != nil {
		return nil, err
	}

	outgoingCtx, err := outgoingAuthContext(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	resp, err := s.roles.ListRoles(outgoingCtx, &iamv1.ListRolesRequest{})
	if err != nil {
		return nil, mapIAMError(err, "list roles")
	}

	return rolesFromProto(resp.GetRoles()), nil
}

// SetUserRoles полностью заменяет роли пользователя через IAM.
func (s *Service) SetUserRoles(ctx context.Context, userID string, roleCodes []string, accessToken string) (*UserRolesResult, error) {
	if !s.Ready() {
		return nil, ErrNotConfigured
	}
	if s.roles == nil {
		return nil, ErrNotConfigured
	}

	subject, err := s.validateAccessToken(ctx, accessToken)
	if err != nil {
		return nil, err
	}
	if err = requireAdmin(subject); err != nil {
		return nil, err
	}

	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrInvalidRequest
	}

	outgoingCtx, err := outgoingAuthContext(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	resp, err := s.roles.SetUserRoles(outgoingCtx, &iamv1.SetUserRolesRequest{
		UserId:    userID,
		RoleCodes: append([]string(nil), roleCodes...),
	})
	if err != nil {
		return nil, mapIAMError(err, "set user roles")
	}

	return &UserRolesResult{
		UserID: resp.GetUserId(),
		Roles:  rolesFromProto(resp.GetRoles()),
		CtxVer: resp.GetCtxVer(),
	}, nil
}

// ReplaceUserAttributes полностью заменяет attributes пользователя через IAM.
func (s *Service) ReplaceUserAttributes(ctx context.Context, userID string, attrs map[string]any, accessToken string) (*UserAttributesResult, error) {
	if !s.Ready() {
		return nil, ErrNotConfigured
	}
	if s.attributes == nil {
		return nil, ErrNotConfigured
	}

	subject, err := s.validateAccessToken(ctx, accessToken)
	if err != nil {
		return nil, err
	}
	if err = requireAdmin(subject); err != nil {
		return nil, err
	}

	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrInvalidRequest
	}

	protoAttrs, err := structpb.NewStruct(attrs)
	if err != nil {
		return nil, ErrInvalidRequest
	}

	outgoingCtx, err := outgoingAuthContext(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	resp, err := s.attributes.ReplaceUserAttributes(outgoingCtx, &iamv1.ReplaceUserAttributesRequest{
		UserId:     userID,
		Attributes: protoAttrs,
	})
	if err != nil {
		return nil, mapIAMError(err, "replace user attributes")
	}

	out := map[string]any{}
	if resp.GetAttributes() != nil {
		out = resp.GetAttributes().AsMap()
	}
	return &UserAttributesResult{UserID: resp.GetUserId(), Attributes: out, CtxVer: resp.GetCtxVer()}, nil
}

func mapIAMError(err error, operation string) error {
	mapped := mapUpstreamError(err, operation)
	if mapped != nil {
		return mapped
	}
	return fmt.Errorf("%s: %w", operation, err)
}

func userFromProto(user *iamv1.User) *User {
	if user == nil {
		return &User{}
	}
	attrs := map[string]any{}
	if user.GetAttributes() != nil {
		attrs = user.GetAttributes().AsMap()
	}
	return &User{
		ID:         user.GetId(),
		Login:      user.GetLogin(),
		IsActive:   user.GetIsActive(),
		CtxVer:     user.GetCtxVer(),
		Roles:      append([]string(nil), user.GetRoles()...),
		Attributes: attrs,
		CreatedAt:  user.GetCreatedAt(),
		UpdatedAt:  user.GetUpdatedAt(),
	}
}

func rolesFromProto(roles []*iamv1.Role) []Role {
	out := make([]Role, 0, len(roles))
	for _, role := range roles {
		if role == nil {
			continue
		}
		out = append(out, Role{
			ID:          role.GetId(),
			Code:        role.GetCode(),
			Name:        role.GetName(),
			Description: role.GetDescription(),
			CreatedAt:   role.GetCreatedAt(),
		})
	}
	return out
}
