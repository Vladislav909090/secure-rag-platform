package config

import "os"

type Key string

const (
	Port             Key = "PORT"
	GRPCPort         Key = "GRPC_PORT"
	IAMGRPC          Key = "IAM_GRPC_ADDR"
	KnowledgeGRPC    Key = "KNOWLEDGE_GRPC_ADDR"
	RAGGRPC          Key = "RAG_GRPC_ADDR"
	DisableAuth      Key = "DISABLE_AUTH"
	DisableDocFilter Key = "DISABLE_DOC_FILTER"
	OPAURL           Key = "OPA_URL"
	DefaultTopK      Key = "GATEWAY_DEFAULT_TOP_K"
	DefaultEmbed     Key = "GATEWAY_DEFAULT_EMBEDDING_MODEL_ALIAS"
	DefaultGenerate  Key = "GATEWAY_DEFAULT_GENERATION_MODEL_ALIAS"
)

func GetValue(key Key) string {
	return os.Getenv(string(key))
}
