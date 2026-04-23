package usecase

import "errors"

var (
	ErrInvalidArgument       = errors.New("invalid argument")
	ErrAliasNotFound         = errors.New("model alias not found")
	ErrAliasTaskMismatch     = errors.New("model alias task mismatch")
	ErrProviderNotConfigured = errors.New("provider not configured")
	ErrProviderFailed        = errors.New("provider request failed")
)
