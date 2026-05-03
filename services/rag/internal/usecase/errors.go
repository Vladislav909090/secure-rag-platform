package usecase

import "errors"

var (
	ErrNotConfigured  = errors.New("service not configured")
	ErrInvalidRequest = errors.New("invalid request")
	ErrNoContexts     = errors.New("no contexts found")
)
