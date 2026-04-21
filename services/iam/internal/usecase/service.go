package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"secure-rag-platform/services/iam/internal/model"
	"secure-rag-platform/services/iam/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

const (
	RoleUser            = "user"
	RoleKnowledgeEditor = "knowledge_editor"
	RoleAccessAdmin     = "access_admin"
	RoleSuperAdmin      = "super_admin"

	tokenTypeBearer = "Bearer"

	subjectCacheKeyPrefix = "iam:subjectctx:"
	loginRateKeyPrefix    = "iam:ratelimit:login:"
	refreshRateKeyPrefix  = "iam:ratelimit:refresh:"
)

var fixedRoleSet = map[string]struct{}{
	RoleUser:            {},
	RoleKnowledgeEditor: {},
	RoleAccessAdmin:     {},
	RoleSuperAdmin:      {},
}

type accessTokenClaims struct {
	SessionID string   `json:"sid"`
	Roles     []string `json:"roles"`
	CtxVer    int64    `json:"ctx_ver"`
	jwt.RegisteredClaims
}

// IAMUsecase содержит бизнес-логику IAM.
type IAMUsecase struct {
	repo  *repository.Repo
	redis *redis.Client
	cfg   Config
}

// NewIAMUsecase создает слой бизнес-логики IAM с переданными зависимостями.
func NewIAMUsecase(repo *repository.Repo, redisClient *redis.Client, cfg Config) *IAMUsecase {
	defaults := DefaultConfig()
	if cfg.JWTIssuer == "" {
		cfg.JWTIssuer = defaults.JWTIssuer
	}
	if cfg.JWTAudience == "" {
		cfg.JWTAudience = defaults.JWTAudience
	}
	if cfg.AccessTokenTTL <= 0 {
		cfg.AccessTokenTTL = defaults.AccessTokenTTL
	}
	if cfg.RefreshTokenTTL <= 0 {
		cfg.RefreshTokenTTL = defaults.RefreshTokenTTL
	}
	if cfg.SubjectCacheTTL <= 0 {
		cfg.SubjectCacheTTL = defaults.SubjectCacheTTL
	}
	if cfg.LoginRateLimit <= 0 {
		cfg.LoginRateLimit = defaults.LoginRateLimit
	}
	if cfg.LoginRateWindow <= 0 {
		cfg.LoginRateWindow = defaults.LoginRateWindow
	}
	if cfg.RefreshRateLimit <= 0 {
		cfg.RefreshRateLimit = defaults.RefreshRateLimit
	}
	if cfg.RefreshRateWindow <= 0 {
		cfg.RefreshRateWindow = defaults.RefreshRateWindow
	}
	if cfg.JWTSecret == "" {
		cfg.JWTSecret = uuid.NewString()
		log.Printf("[iam] JWT secret is empty; generated ephemeral secret for current run")
	}

	return &IAMUsecase{
		repo:  repo,
		redis: redisClient,
		cfg:   cfg,
	}
}

func normalizeRoleCodes(roleCodes []string) ([]string, error) {
	out := make([]string, 0, len(roleCodes))
	seen := make(map[string]struct{}, len(roleCodes))
	for _, roleCode := range roleCodes {
		code := strings.TrimSpace(roleCode)
		if code == "" {
			continue
		}
		if _, ok := fixedRoleSet[code]; !ok {
			return nil, ErrInvalidArgument
		}
		if _, exists := seen[code]; exists {
			continue
		}
		seen[code] = struct{}{}
		out = append(out, code)
	}
	if len(out) == 0 {
		out = []string{RoleUser}
	}
	slices.Sort(out)
	return out, nil
}

func hashPassword(password string) (string, error) {
	if strings.TrimSpace(password) == "" {
		return "", ErrInvalidArgument
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

func checkPassword(hash string, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func randomToken(size int) (string, error) {
	if size <= 0 {
		size = 32
	}

	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate random bytes: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func hashOpaqueToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (uc *IAMUsecase) issueAccessToken(subject *model.SubjectContext, sessionID string) (string, int64, string, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(uc.cfg.AccessTokenTTL)
	tokenID := uuid.NewString()

	claims := accessTokenClaims{
		SessionID: sessionID,
		Roles:     append([]string(nil), subject.Roles...),
		CtxVer:    subject.CtxVer,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    uc.cfg.JWTIssuer,
			Subject:   subject.UserID,
			Audience:  jwt.ClaimStrings{uc.cfg.JWTAudience},
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        tokenID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(uc.cfg.JWTSecret))
	if err != nil {
		return "", 0, "", fmt.Errorf("sign access token: %w", err)
	}

	expiresIn := int64(expiresAt.Sub(now).Seconds())
	if expiresIn < 0 {
		expiresIn = 0
	}

	return signed, expiresIn, tokenID, nil
}

func (uc *IAMUsecase) parseAccessToken(accessToken string) (*accessTokenClaims, error) {
	parsedToken, err := jwt.ParseWithClaims(
		accessToken,
		&accessTokenClaims{},
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, ErrInvalidToken
			}
			return []byte(uc.cfg.JWTSecret), nil
		},
		jwt.WithIssuer(uc.cfg.JWTIssuer),
		jwt.WithAudience(uc.cfg.JWTAudience),
	)
	if err != nil {
		return nil, fmt.Errorf("parse access token: %w", err)
	}

	claims, ok := parsedToken.Claims.(*accessTokenClaims)
	if !ok || !parsedToken.Valid {
		return nil, ErrInvalidToken
	}

	if claims.Subject == "" || claims.SessionID == "" {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func subjectCacheKey(userID string) string {
	return subjectCacheKeyPrefix + userID
}

func (uc *IAMUsecase) getSubjectContextFromCache(ctx context.Context, userID string) (*model.SubjectContext, bool) {
	if uc.redis == nil {
		return nil, false
	}

	raw, err := uc.redis.Get(ctx, subjectCacheKey(userID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, false
		}
		log.Printf("[iam] redis get subject context failed: %v", err)
		return nil, false
	}

	var cached model.SubjectContext
	if err := json.Unmarshal([]byte(raw), &cached); err != nil {
		log.Printf("[iam] redis unmarshal subject context failed: %v", err)
		return nil, false
	}

	return &cached, true
}

func (uc *IAMUsecase) storeSubjectContextInCache(ctx context.Context, subject *model.SubjectContext) {
	if uc.redis == nil || subject == nil {
		return
	}

	raw, err := json.Marshal(subject)
	if err != nil {
		log.Printf("[iam] redis marshal subject context failed: %v", err)
		return
	}

	if err := uc.redis.Set(ctx, subjectCacheKey(subject.UserID), raw, uc.cfg.SubjectCacheTTL).Err(); err != nil {
		log.Printf("[iam] redis set subject context failed: %v", err)
	}
}

// InvalidateSubjectContextCache очищает кеш контекста пользователя.
func (uc *IAMUsecase) InvalidateSubjectContextCache(ctx context.Context, userID string) {
	if uc.redis == nil || userID == "" {
		return
	}
	if err := uc.redis.Del(ctx, subjectCacheKey(userID)).Err(); err != nil {
		log.Printf("[iam] redis invalidate subject context failed: %v", err)
	}
}

func (uc *IAMUsecase) checkRateLimit(ctx context.Context, key string, limit int, window time.Duration) error {
	if uc.redis == nil || limit <= 0 || window <= 0 || key == "" {
		return nil
	}

	count, err := uc.redis.Incr(ctx, key).Result()
	if err != nil {
		log.Printf("[iam] redis rate limit incr failed: %v", err)
		return nil
	}

	if count == 1 {
		if err := uc.redis.Expire(ctx, key, window).Err(); err != nil {
			log.Printf("[iam] redis rate limit expire failed: %v", err)
		}
	}

	if count > int64(limit) {
		return ErrRateLimited
	}
	return nil
}

func (uc *IAMUsecase) bumpContextVersion(ctx context.Context, userID string) (int64, error) {
	ctxVer, err := uc.repo.IncrementContextVersion(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return 0, ErrNotFound
		}
		return 0, err
	}
	uc.InvalidateSubjectContextCache(ctx, userID)
	return ctxVer, nil
}

func hasRole(roles []string, required string) bool {
	for _, role := range roles {
		if role == required {
			return true
		}
	}
	return false
}

func hasAnyRole(roles []string, required ...string) bool {
	for _, r := range required {
		if hasRole(roles, r) {
			return true
		}
	}
	return false
}

func mergeAttributes(current map[string]any, patch map[string]any) map[string]any {
	merged := map[string]any{}
	for k, v := range current {
		merged[k] = v
	}
	for k, v := range patch {
		merged[k] = v
	}
	return merged
}

func sanitizeOptional(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

// GetSubjectContext возвращает нормализованный контекст с резервным чтением из кеша.
func (uc *IAMUsecase) GetSubjectContext(ctx context.Context, userID string) (*model.SubjectContext, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrInvalidArgument
	}

	if cached, ok := uc.getSubjectContextFromCache(ctx, userID); ok {
		return cached, nil
	}

	subject, err := uc.repo.GetSubjectContext(ctx, userID)
	if err != nil {
		return nil, err
	}
	if subject == nil {
		return nil, ErrNotFound
	}

	uc.storeSubjectContextInCache(ctx, subject)
	return subject, nil
}

// AuthenticateAccessToken проверяет токен доступа и возвращает данные принципала с актуальным контекстом субъекта.
func (uc *IAMUsecase) AuthenticateAccessToken(ctx context.Context, accessToken string) (*Principal, *model.SubjectContext, error) {
	claims, err := uc.parseAccessToken(accessToken)
	if err != nil {
		return nil, nil, ErrUnauthorized
	}

	subject, err := uc.GetSubjectContext(ctx, claims.Subject)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, nil, ErrUnauthorized
		}
		return nil, nil, err
	}

	if !subject.IsActive {
		return nil, nil, ErrInactiveUser
	}
	if claims.CtxVer != subject.CtxVer {
		return nil, nil, ErrInvalidToken
	}

	session, err := uc.repo.GetSessionByID(ctx, claims.SessionID)
	if err != nil {
		return nil, nil, err
	}
	if session == nil || session.UserID != subject.UserID {
		return nil, nil, ErrInvalidToken
	}
	if session.RevokedAt != nil {
		return nil, nil, ErrSessionRevoked
	}
	if session.ExpiresAt.Before(time.Now().UTC()) {
		return nil, nil, ErrSessionExpired
	}

	principal := &Principal{
		UserID:    subject.UserID,
		SessionID: claims.SessionID,
		TokenID:   claims.ID,
		Roles:     append([]string(nil), subject.Roles...),
		CtxVer:    subject.CtxVer,
		ExpiresAt: claims.ExpiresAt.Time,
	}
	return principal, subject, nil
}

// BootstrapSuperAdmin создает начального суперадмина, если он отсутствует.
func (uc *IAMUsecase) BootstrapSuperAdmin(ctx context.Context, login string, password string) (generatedPassword string, created bool, err error) {
	exists, err := uc.repo.HasUserWithRole(ctx, RoleSuperAdmin)
	if err != nil {
		return "", false, fmt.Errorf("check super admin existence: %w", err)
	}
	if exists {
		return "", false, nil
	}

	login = strings.TrimSpace(login)
	if login == "" {
		return "", false, ErrInvalidArgument
	}

	pwd := strings.TrimSpace(password)
	if pwd == "" {
		pwd, err = randomToken(24)
		if err != nil {
			return "", false, err
		}
		generatedPassword = pwd
	}

	active := true
	_, err = uc.CreateUser(ctx, CreateUserInput{
		Login:      login,
		Password:   pwd,
		IsActive:   &active,
		RoleCodes:  []string{RoleSuperAdmin, RoleUser},
		Attributes: map[string]any{},
	})
	if err != nil {
		if errors.Is(err, ErrUserExists) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("create bootstrap super admin: %w", err)
	}

	return generatedPassword, true, nil
}
