package usecase

import (
	"time"

	"secure-rag-platform/services/iam/internal/model"
)

// TokenPair возвращается из операций входа и обновления токена
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
	TokenType    string
}

// LoginInput содержит учетные данные для входа
type LoginInput struct {
	Login    string
	Password string
}

// RefreshTokenInput содержит токен обновления
type RefreshTokenInput struct {
	RefreshToken string
}

// Principal описывает идентификатор аутентифицированного токена
type Principal struct {
	UserID    string
	SessionID string
	TokenID   string
	Roles     []string
	CtxVer    int64
	ExpiresAt time.Time
}

// ValidateTokenResult возвращается после проверки токена доступа
type ValidateTokenResult struct {
	Valid         bool
	Reason        string
	Subject       *model.SubjectContext
	Principal     *Principal
	ExpiresAtUnix int64
}

// LogoutAllResult содержит результат массового отзыва сессий
type LogoutAllResult struct {
	UserID       string
	RevokedCount int64
	CtxVer       int64
}

// CreateUserInput содержит поля для создания нового IAM-пользователя
type CreateUserInput struct {
	Login      string
	Password   string
	IsActive   *bool
	RoleCodes  []string
	Attributes map[string]any
	CreatedBy  *string
}

// UpdateUserInput содержит изменяемые поля для обновления пользователя
type UpdateUserInput struct {
	UserID   string
	Login    *string
	Password *string
	IsActive *bool
}
