package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"strconv"

	aiinferencev1 "secure-rag-platform/services/ai-inference/gen/v1"
	knowledgev1 "secure-rag-platform/services/knowledge/gen/v1"
	ragv1 "secure-rag-platform/services/rag/gen/v1"
	application "secure-rag-platform/services/rag/internal/app"
	"secure-rag-platform/services/rag/internal/closer"
	"secure-rag-platform/services/rag/internal/config"
	"secure-rag-platform/services/rag/internal/docs"
	"secure-rag-platform/services/rag/internal/repository"
	"secure-rag-platform/services/rag/internal/storage"
	transportgrpc "secure-rag-platform/services/rag/internal/transport/grpc"
	"secure-rag-platform/services/rag/internal/usecase"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pgvectorpgx "github.com/pgvector/pgvector-go/pgx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	port := config.GetValue(config.Port)
	if port == "" {
		port = "8083"
	}

	grpcPort := config.GetValue(config.GRPCPort)
	if grpcPort == "" {
		grpcPort = "9093"
	}

	chunkSize := parseInt(config.GetValue(config.ChunkSize), 800)
	chunkOverlap := parseInt(config.GetValue(config.ChunkOverlap), 100)
	defaultTopK := int32(parseInt(config.GetValue(config.DefaultTopK), 3))
	defaultEmbed := config.GetValue(config.DefaultEmbed)
	if defaultEmbed == "" {
		defaultEmbed = "embed.default"
	}
	defaultGenerate := config.GetValue(config.DefaultGenerate)
	if defaultGenerate == "" {
		defaultGenerate = "chat.default"
	}

	uc := buildUsecase(chunkSize, chunkOverlap, defaultTopK, defaultEmbed, defaultGenerate)

	serverImpl := transportgrpc.NewServer(uc)
	grpcServer := grpc.NewServer()
	ragv1.RegisterRAGServiceServer(grpcServer, serverImpl)

	grpcLis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("failed to listen grpc: %v", err)
	}

	mux := http.NewServeMux()

	gwMux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames:   true,
				EmitUnpopulated: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
	)
	if err := ragv1.RegisterRAGServiceHandlerServer(context.Background(), gwMux, serverImpl); err != nil {
		log.Fatalf("failed to register rag handlers: %v", err)
	}
	mux.Handle("/rag/", gwMux)

	docs.RegisterAt(mux, "RAG", "/rag/docs")

	app := application.New()

	httpServer := &http.Server{Addr: ":" + port, Handler: mux}

	log.Printf("rag grpc listening on :%s", grpcPort)
	log.Printf("rag listening on :%s", port)
	log.Printf("docs: http://localhost/rag/docs")

	app.Add(func() error {
		return grpcServer.Serve(grpcLis)
	})
	app.Add(func() error {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})

	closer.Add(httpServer.Shutdown)
	closer.Add(grpcServer.GracefulStop)
	closer.Add(grpcLis.Close)

	if err := app.Run(); err != nil {
		log.Fatalf("application stopped with error: %v", err)
	}
}

func buildUsecase(chunkSize int, chunkOverlap int, topK int32, embedAlias string, genAlias string) *usecase.Service {
	dbDSN := config.GetValue(config.DatabaseDSN)
	if dbDSN == "" {
		dbDSN = config.GetValue(config.LegacyDBDSN)
	}
	if dbDSN == "" {
		log.Println("rag: DATABASE_DSN not set, indexing/query endpoints unavailable")
		return nil
	}

	cfg, err := pgxpool.ParseConfig(dbDSN)
	if err != nil {
		log.Fatalf("failed to parse database dsn: %v", err)
	}
	cfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		return pgvectorpgx.RegisterTypes(ctx, conn)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	closer.Add(func() { pool.Close() })

	s3Store, err := storage.NewS3Storage(
		config.GetValue(config.S3Endpoint),
		config.GetValue(config.S3AccessKey),
		config.GetValue(config.S3SecretKey),
		config.GetValue(config.S3Bucket),
		config.GetValue(config.S3UseSSL) == "true",
	)
	if err != nil {
		log.Fatalf("failed to init S3 storage: %v", err)
	}
	if err := s3Store.EnsureBucket(context.Background()); err != nil {
		log.Fatalf("failed to ensure S3 bucket: %v", err)
	}

	knowledgeAddr := config.GetValue(config.KnowledgeGRPC)
	if knowledgeAddr == "" {
		knowledgeAddr = "127.0.0.1:9092"
	}
	knowledgeConn, err := grpc.NewClient(knowledgeAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to knowledge gRPC: %v", err)
	}
	closer.Add(func() { _ = knowledgeConn.Close() })

	aiAddr := config.GetValue(config.AIInferenceGRPC)
	if aiAddr == "" {
		aiAddr = "127.0.0.1:9094"
	}
	aiConn, err := grpc.NewClient(aiAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to ai-inference gRPC: %v", err)
	}
	closer.Add(func() { _ = aiConn.Close() })

	repo := repository.NewRepo(pool)
	defaults := usecase.Defaults{
		EmbeddingModelAlias:  embedAlias,
		GenerationModelAlias: genAlias,
		ChunkSize:            chunkSize,
		ChunkOverlap:         chunkOverlap,
		TopK:                 topK,
	}

	uc := usecase.NewService(
		repo,
		s3Store,
		knowledgev1.NewKnowledgeServiceClient(knowledgeConn),
		aiinferencev1.NewEmbeddingServiceClient(aiConn),
		aiinferencev1.NewGenerationServiceClient(aiConn),
		defaults,
		log.Default(),
	)

	log.Println("rag: database, S3, and upstream services configured")
	return uc
}

func parseInt(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}

	val, err := strconv.Atoi(raw)
	if err != nil || val <= 0 {
		return fallback
	}
	return val
}
