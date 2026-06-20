package config

import "os"

type Key string

const (
	Port                   Key = "PORT"
	GRPCPort               Key = "GRPC_PORT"
	DatabaseDSN            Key = "DATABASE_DSN"
	LegacyDBDSN            Key = "DB_DSN"
	RedisAddr              Key = "REDIS_ADDR"
	RedisPassword          Key = "REDIS_PASSWORD"
	JWTSecret              Key = "JWT_SECRET"
	JWTIssuer              Key = "JWT_ISSUER"
	JWTAudience            Key = "JWT_AUDIENCE"
	BootstrapAdminLogin    Key = "BOOTSTRAP_ADMIN_LOGIN"
	BootstrapAdminPassword Key = "BOOTSTRAP_ADMIN_PASSWORD"
)

func GetValue(key Key) string {
	return os.Getenv(string(key))
}

func GetFirstValue(keys ...Key) string {
	for _, key := range keys {
		if value := GetValue(key); value != "" {
			return value
		}
	}

	return ""
}
