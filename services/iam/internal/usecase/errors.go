package usecase

import "errors"

var (
	ErrInvalidArgument    = errors.New("invalid argument")
	ErrInvalidToken       = errors.New("invalid token")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrNotFound           = errors.New("not found")
	ErrUserExists         = errors.New("user already exists")
	ErrInactiveUser       = errors.New("user is inactive")
	ErrRateLimited        = errors.New("rate limit exceeded")
	ErrSessionExpired     = errors.New("session expired")
	ErrSessionRevoked     = errors.New("session revoked")
)
