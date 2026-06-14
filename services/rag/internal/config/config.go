package config

import "os"

type Key string

const (
	Port            Key = "PORT"
	GRPCPort        Key = "GRPC_PORT"
	DatabaseDSN     Key = "DATABASE_DSN"
	LegacyDBDSN     Key = "DB_DSN"
	S3Endpoint      Key = "S3_ENDPOINT"
	S3Bucket        Key = "S3_BUCKET"
	S3AccessKey     Key = "S3_ACCESS_KEY"
	S3SecretKey     Key = "S3_SECRET_KEY"
	S3UseSSL        Key = "S3_USE_SSL"
	KnowledgeGRPC   Key = "KNOWLEDGE_GRPC_ADDR"
	AIInferenceGRPC Key = "AI_INFERENCE_GRPC_ADDR"
	ChunkSize       Key = "RAG_CHUNK_SIZE"
	ChunkOverlap    Key = "RAG_CHUNK_OVERLAP"
	DefaultTopK     Key = "RAG_DEFAULT_TOP_K"
	DefaultEmbed    Key = "RAG_DEFAULT_EMBEDDING_MODEL_ALIAS"
	DefaultGenerate Key = "RAG_DEFAULT_GENERATION_MODEL_ALIAS"
	IndexedEmbedDim Key = "RAG_INDEXED_EMBEDDING_DIMENSION"
)

func GetValue(key Key) string {
	return os.Getenv(string(key))
}
