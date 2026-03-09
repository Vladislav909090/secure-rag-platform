package config

import "os"

type Key string

const (
	Port            Key = "PORT"
	GRPCPort        Key = "GRPC_PORT"
)

func GetValue(key Key) string {
	return os.Getenv(string(key))
}
