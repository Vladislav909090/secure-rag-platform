package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"secure-rag-platform/services/iam/internal/model"

	"github.com/jackc/pgx/v5"
)

// CreateSessionInput содержит данные, необходимые для создания сессии токена обновления.
type CreateSessionInput struct {
	UserID           string
	RefreshTokenHash string
	ExpiresAt        time.Time
}

func scanSession(row pgx.Row) (*model.UserSession, error) {
	var s model.UserSession
	if err := row.Scan(
		&s.ID,
		&s.UserID,
		&s.RefreshTokenHash,
		&s.ExpiresAt,
		&s.RevokedAt,
		&s.CreatedAt,
		&s.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &s, nil
}

// CreateSession создает новую пользовательскую сессию.
func (r *Repo) CreateSession(ctx context.Context, input CreateSessionInput) (*model.UserSession, error) {
	session, err := scanSession(r.pool.QueryRow(ctx, `
		INSERT INTO user_sessions (user_id, refresh_token_hash, expires_at, revoked_at, updated_at)
		VALUES ($1, $2, $3, NULL, NOW())
		ON CONFLICT (user_id) DO UPDATE
		SET refresh_token_hash = EXCLUDED.refresh_token_hash,
		    expires_at = EXCLUDED.expires_at,
		    revoked_at = NULL,
		    updated_at = NOW()
		RETURNING id, user_id, refresh_token_hash, expires_at, revoked_at, created_at, updated_at
	`, input.UserID, input.RefreshTokenHash, input.ExpiresAt))
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	return session, nil
}

// GetSessionByID возвращает сессию по идентификатору.
func (r *Repo) GetSessionByID(ctx context.Context, sessionID string) (*model.UserSession, error) {
	session, err := scanSession(r.pool.QueryRow(ctx, `
		SELECT id, user_id, refresh_token_hash, expires_at, revoked_at, created_at, updated_at
		FROM user_sessions
		WHERE id = $1
	`, sessionID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("query session by id: %w", err)
	}
	return session, nil
}

// GetSessionByRefreshHash возвращает сессию по хешу токена обновления.
func (r *Repo) GetSessionByRefreshHash(ctx context.Context, refreshHash string) (*model.UserSession, error) {
	session, err := scanSession(r.pool.QueryRow(ctx, `
		SELECT id, user_id, refresh_token_hash, expires_at, revoked_at, created_at, updated_at
		FROM user_sessions
		WHERE refresh_token_hash = $1
	`, refreshHash))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("query session by refresh hash: %w", err)
	}
	return session, nil
}

// RotateSessionRefreshHash ротирует хеш токена обновления и обновляет срок действия активной сессии.
func (r *Repo) RotateSessionRefreshHash(
	ctx context.Context,
	sessionID string,
	newRefreshHash string,
	expiresAt time.Time,
) (*model.UserSession, error) {
	session, err := scanSession(r.pool.QueryRow(ctx, `
		UPDATE user_sessions
		SET refresh_token_hash = $2,
		    expires_at = $3,
		    updated_at = NOW()
		WHERE id = $1
		  AND revoked_at IS NULL
		  AND expires_at > NOW()
		RETURNING id, user_id, refresh_token_hash, expires_at, revoked_at, created_at, updated_at
	`, sessionID, newRefreshHash, expiresAt))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("rotate session refresh hash: %w", err)
	}
	return session, nil
}

// RevokeSession отзывает одну сессию, если она еще не отозвана.
func (r *Repo) RevokeSession(ctx context.Context, sessionID string) (bool, error) {
	result, err := r.pool.Exec(ctx, `
		UPDATE user_sessions
		SET revoked_at = NOW(),
		    updated_at = NOW()
		WHERE id = $1
		  AND revoked_at IS NULL
	`, sessionID)
	if err != nil {
		return false, fmt.Errorf("revoke session: %w", err)
	}
	return result.RowsAffected() > 0, nil
}

// RevokeAllUserSessions отзывает все активные сессии пользователя.
func (r *Repo) RevokeAllUserSessions(ctx context.Context, userID string) (int64, error) {
	result, err := r.pool.Exec(ctx, `
		UPDATE user_sessions
		SET revoked_at = NOW(),
		    updated_at = NOW()
		WHERE user_id = $1
		  AND revoked_at IS NULL
		  AND expires_at > NOW()
	`, userID)
	if err != nil {
		return 0, fmt.Errorf("revoke all user sessions: %w", err)
	}
	return result.RowsAffected(), nil
}

// ListActiveUserSessions возвращает неотозванные и неистекшие сессии.
func (r *Repo) ListActiveUserSessions(ctx context.Context, userID string) ([]*model.UserSession, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, refresh_token_hash, expires_at, revoked_at, created_at, updated_at
		FROM user_sessions
		WHERE user_id = $1
		  AND revoked_at IS NULL
		  AND expires_at > NOW()
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("query active sessions: %w", err)
	}
	defer rows.Close()

	sessions := make([]*model.UserSession, 0)
	for rows.Next() {
		session, err := scanSession(rows)
		if err != nil {
			return nil, fmt.Errorf("scan session: %w", err)
		}
		sessions = append(sessions, session)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sessions: %w", err)
	}

	return sessions, nil
}
