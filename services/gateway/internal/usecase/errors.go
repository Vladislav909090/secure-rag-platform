package usecase

import "errors"

var (
	ErrNotConfigured     = errors.New("service not configured")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
	ErrNotFound          = errors.New("not found")
	ErrInvalidRequest    = errors.New("invalid request")
	ErrPolicyRequired    = errors.New("policy authorizer required")
	ErrPolicyUnavailable = errors.New("policy authorizer unavailable")
)
