package grpc

import (
	"context"

	pb "secure-rag-platform/api/gen/go/gateway/v1"
	"secure-rag-platform/services/gateway/internal/usecase"
)

func (s *Server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.TokenPairResponse, error) {
	if err := s.requireUC(); err != nil {
		return nil, err
	}

	tokens, err := s.uc.Login(ctx, usecase.LoginRequest{
		Login:    req.GetLogin(),
		Password: req.GetPassword(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return tokenPairToProto(tokens), nil
}

func (s *Server) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.TokenPairResponse, error) {
	if err := s.requireUC(); err != nil {
		return nil, err
	}

	tokens, err := s.uc.RefreshToken(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, toGRPCError(err)
	}

	return tokenPairToProto(tokens), nil
}

func (s *Server) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	_ = req

	if err := s.requireUC(); err != nil {
		return nil, err
	}

	revoked, err := s.uc.Logout(ctx, extractAccessToken(ctx))
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.LogoutResponse{Revoked: revoked}, nil
}

func (s *Server) GetMe(ctx context.Context, req *pb.GetMeRequest) (*pb.GetMeResponse, error) {
	_ = req

	if err := s.requireUC(); err != nil {
		return nil, err
	}

	me, err := s.uc.GetMe(ctx, extractAccessToken(ctx))
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.GetMeResponse{Me: subjectToProto(me)}, nil
}

func tokenPairToProto(tokens *usecase.TokenPair) *pb.TokenPairResponse {
	if tokens == nil {
		return &pb.TokenPairResponse{}
	}
	return &pb.TokenPairResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
		TokenType:    tokens.TokenType,
	}
}

func subjectToProto(subject *usecase.SubjectContext) *pb.SubjectContext {
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
