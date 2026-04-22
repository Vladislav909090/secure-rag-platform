package model

import "time"

// User хранит учетную запись субъекта IAM.
type User struct {
	ID           string
	Login        string
	PasswordHash string
	IsActive     bool
	CtxVer       int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Role - фиксированная RBAC-роль.
type Role struct {
	ID          int64
	Code        string
	Name        string
	Description string
	CreatedAt   time.Time
}

// SubjectContext - нормализованный рабочий контекст безопасности.
type SubjectContext struct {
	UserID     string
	Login      string
	IsActive   bool
	Roles      []string
	Attributes map[string]any
	CtxVer     int64
}

// UserView предоставляет данные пользователя вместе с ролями и атрибутами.
type UserView struct {
	ID         string
	Login      string
	IsActive   bool
	CtxVer     int64
	Roles      []string
	Attributes map[string]any
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// UserSession хранит состояние сессии токена обновления.
type UserSession struct {
	ID               string
	UserID           string
	RefreshTokenHash string
	ExpiresAt        time.Time
	RevokedAt        *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
