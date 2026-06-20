package usecase

import "time"

// Config задает бизнес-настройки IAM
type Config struct {
	JWTSecret         string
	JWTIssuer         string
	JWTAudience       string
	AccessTokenTTL    time.Duration
	RefreshTokenTTL   time.Duration
	SubjectCacheTTL   time.Duration
	LoginRateLimit    int
	LoginRateWindow   time.Duration
	RefreshRateLimit  int
	RefreshRateWindow time.Duration
}

// DefaultConfig возвращает безопасные значения по умолчанию для рабочих настроек IAM
func DefaultConfig() Config {
	return Config{
		JWTIssuer:         "secure-rag-iam",
		JWTAudience:       "secure-rag-gateway",
		AccessTokenTTL:    15 * time.Minute,
		RefreshTokenTTL:   7 * 24 * time.Hour,
		SubjectCacheTTL:   2 * time.Minute,
		LoginRateLimit:    10,
		LoginRateWindow:   time.Minute,
		RefreshRateLimit:  20,
		RefreshRateWindow: time.Minute,
	}
}
