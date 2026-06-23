package grpc

import (
	"context"
	"strings"

	"secure-rag-platform/services/iam/internal/model"
	"secure-rag-platform/services/iam/internal/usecase"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func requireUC(uc IAMUsecaseContract) error {
	if uc == nil {
		return status.Error(codes.Unavailable, "service not configured")
	}

	return nil
}

func authenticate(uc IAMUsecaseContract, ctx context.Context) (*usecase.Principal, *model.SubjectContext, error) {
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
