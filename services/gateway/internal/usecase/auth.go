package usecase

import (
	"context"
	"strings"

	iamv1 "secure-rag-platform/api/gen/go/iam/v1"

	"google.golang.org/grpc/metadata"
)

const (
	roleUser            = "user"
	roleKnowledgeEditor = "knowledge_editor"
	roleAccessAdmin     = "access_admin"
	roleSuperAdmin      = "super_admin"
)

func (s *Service) Login(ctx context.Context, req LoginRequest) (*TokenPair, error) {
	if s.disableAuth {
		return nil, ErrNotConfigured
	}
	if s.auth == nil {
		return nil, ErrNotConfigured
	}

	login := strings.TrimSpace(req.Login)
	if login == "" || strings.TrimSpace(req.Password) == "" {
		return nil, ErrInvalidRequest
	}

	resp, err := s.auth.Login(ctx, &iamv1.LoginRequest{
		Login:    login,
		Password: req.Password,
	})
	if err != nil {
		return nil, ErrUnauthorized
	}

	return tokenPairFromProto(resp), nil
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	if s.disableAuth {
		return nil, ErrNotConfigured
	}
	if s.auth == nil {
		return nil, ErrNotConfigured
	}

	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil, ErrInvalidRequest
	}

	resp, err := s.auth.RefreshToken(ctx, &iamv1.RefreshTokenRequest{RefreshToken: refreshToken})
	if err != nil {
		return nil, ErrUnauthorized
	}

	return tokenPairFromProto(resp), nil
}

func (s *Service) Logout(ctx context.Context, accessToken string) (bool, error) {
	if s.disableAuth {
		return false, ErrNotConfigured
	}
	if s.auth == nil {
		return false, ErrNotConfigured
	}

	ctx, err := outgoingAuthContext(ctx, accessToken)
	if err != nil {
		return false, err
	}

	resp, err := s.auth.Logout(ctx, &iamv1.LogoutRequest{})
	if err != nil {
		return false, ErrUnauthorized
	}

	return resp.GetRevoked(), nil
}

func (s *Service) GetMe(ctx context.Context, accessToken string) (*SubjectContext, error) {
	if s.disableAuth {
		return nil, ErrNotConfigured
	}
	if s.auth == nil {
		return nil, ErrNotConfigured
	}

	ctx, err := outgoingAuthContext(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	resp, err := s.auth.GetMe(ctx, &iamv1.GetMeRequest{})
	if err != nil {
		return nil, ErrUnauthorized
	}

	return subjectFromProto(resp.GetMe()), nil
}

func (s *Service) validateAccessToken(ctx context.Context, token string) (*iamv1.SubjectContext, error) {
	if s.disableAuth {
		return nil, nil
	}
	if s.iam == nil {
		return nil, ErrNotConfigured
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return nil, ErrUnauthorized
	}

	resp, err := s.iam.ValidateAccessToken(ctx, &iamv1.ValidateAccessTokenRequest{AccessToken: token})
	if err != nil {
		return nil, ErrUnauthorized
	}
	if !resp.GetValid() {
		return nil, ErrUnauthorized
	}

	return resp.GetSubject(), nil
}

func requireAdmin(subject *iamv1.SubjectContext) error {
	if subject == nil {
		return nil
	}
	if hasAnyRole(subject, roleSuperAdmin, roleAccessAdmin) {
		return nil
	}
	return ErrForbidden
}

func requireDocumentEditor(subject *iamv1.SubjectContext) error {
	if subject == nil {
		return nil
	}
	if hasAnyRole(subject, roleSuperAdmin, roleKnowledgeEditor) {
		return nil
	}
	return ErrForbidden
}

func hasAnyRole(subject *iamv1.SubjectContext, allowed ...string) bool {
	if subject == nil {
		return false
	}
	for _, role := range subject.GetRoles() {
		for _, expected := range allowed {
			if role == expected {
				return true
			}
		}
	}
	return false
}

func outgoingAuthContext(ctx context.Context, accessToken string) (context.Context, error) {
	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return nil, ErrUnauthorized
	}
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+accessToken), nil
}

func tokenPairFromProto(resp *iamv1.TokenPairResponse) *TokenPair {
	if resp == nil {
		return &TokenPair{}
	}
	return &TokenPair{
		AccessToken:  resp.GetAccessToken(),
		RefreshToken: resp.GetRefreshToken(),
		ExpiresIn:    resp.GetExpiresIn(),
		TokenType:    resp.GetTokenType(),
	}
}

func subjectFromProto(subject *iamv1.SubjectContext) *SubjectContext {
	if subject == nil {
		return nil
	}
	attrs := map[string]any{}
	if subject.GetAttributes() != nil {
		attrs = subject.GetAttributes().AsMap()
	}
	return &SubjectContext{
		UserID:     subject.GetUserId(),
		Login:      subject.GetLogin(),
		IsActive:   subject.GetIsActive(),
		Roles:      append([]string(nil), subject.GetRoles()...),
		Attributes: attrs,
		CtxVer:     subject.GetCtxVer(),
	}
}
