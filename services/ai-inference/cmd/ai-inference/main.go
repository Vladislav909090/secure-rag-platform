package main

import (
	"context"
	"errors"
	"log"
	"net"
	"time"

	aiinferencev1 "secure-rag-platform/services/ai-inference/gen/v1"
	application "secure-rag-platform/services/ai-inference/internal/app"
	"secure-rag-platform/services/ai-inference/internal/closer"
	"secure-rag-platform/services/ai-inference/internal/config"
	"secure-rag-platform/services/ai-inference/internal/provider"
	transportgrpc "secure-rag-platform/services/ai-inference/internal/transport/grpc"
	"secure-rag-platform/services/ai-inference/internal/usecase"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	grpcHealthV1 "google.golang.org/grpc/health/grpc_health_v1"
)

const defaultGRPCPort = "9094"

func main() {
	runtimeCfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	grpcPort := runtimeCfg.GRPCPort
	if grpcPort == "" {
		grpcPort = defaultGRPCPort
	}

	providerSet := []usecase.Provider{
		provider.NewOpenAICompatProvider(60 * time.Second),
	}
	inferenceService := usecase.NewService(runtimeCfg.ModelAliases, providerSet, log.Default())

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(loggingInterceptor))
	aiinferencev1.RegisterAIInferenceServiceServer(grpcServer, transportgrpc.NewAIInferenceServiceServer())
	aiinferencev1.RegisterGenerationServiceServer(grpcServer, transportgrpc.NewGenerationServiceServer(inferenceService))
	aiinferencev1.RegisterEmbeddingServiceServer(grpcServer, transportgrpc.NewEmbeddingServiceServer(inferenceService))

	healthServer := health.NewServer()
	grpcHealthV1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpcHealthV1.HealthCheckResponse_SERVING)

	grpcLis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("failed to listen grpc: %v", err)
	}

	app := application.New()
	app.Add(func() error {
		return grpcServer.Serve(grpcLis)
	})

	closer.Add(grpcServer.GracefulStop)
	closer.Add(grpcLis.Close)

	log.Printf("ai-inference grpc listening on :%s", grpcPort)

	if err := app.Run(); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("application stopped with error: %v", err)
	}
}

func loggingInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	startedAt := time.Now()
	resp, err := handler(ctx, req)
	if err != nil {
		log.Printf("[grpc] method=%s duration=%s error=%v", info.FullMethod, time.Since(startedAt), err)
		return nil, err
	}

	log.Printf("[grpc] method=%s duration=%s", info.FullMethod, time.Since(startedAt))
	return resp, nil
}
