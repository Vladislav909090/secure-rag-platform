package config

import "os"

type Key string

const (
	Port        Key = "PORT"
	GRPCPort    Key = "GRPC_PORT"
	DatabaseDSN Key = "DATABASE_DSN"
	S3Endpoint  Key = "S3_ENDPOINT"
	S3Bucket    Key = "S3_BUCKET"
	S3AccessKey Key = "S3_ACCESS_KEY"
	S3SecretKey Key = "S3_SECRET_KEY"
	S3UseSSL    Key = "S3_USE_SSL"
	MaxFileSize Key = "MAX_FILE_SIZE"
)

func GetValue(key Key) string {
	return os.Getenv(string(key))
}
