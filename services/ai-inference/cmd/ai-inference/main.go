package main

import (
	"context"
	"errors"
	"flag"
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
)

const defaultGRPCPort = "9094"

func main() {
	modelsConfigPath := flag.String("config", config.DefaultModelsConfigPath, "Path to models config JSON file")
	flag.Parse()

	runtimeCfg, err := config.LoadFromFile(*modelsConfigPath)
	if err != nil {
		log.Fatalf("[ai-inference.config] не удалось загрузить конфигурацию моделей: %v", err)
	}

	grpcPort := runtimeCfg.GRPCPort
	if grpcPort == "" {
		grpcPort = defaultGRPCPort
	}

	providerTimeout := 180 * time.Second
	if runtimeCfg.ProviderTimeout != "" {
		var parsed time.Duration
		parsed, err = time.ParseDuration(runtimeCfg.ProviderTimeout)
		if err != nil {
			log.Fatalf("[ai-inference.config] не удалось разобрать provider_timeout: %v", err)
		}
		providerTimeout = parsed
	}

	providerSet := []usecase.Provider{
		provider.NewOpenAICompatProvider(providerTimeout),
	}
	inferenceService := usecase.NewService(runtimeCfg.ModelAliases, providerSet, log.Default())

	startupCheckTimeout := providerTimeout + 5*time.Second
	startupCtx, cancel := context.WithTimeout(context.Background(), startupCheckTimeout)
	defer cancel()
	err = inferenceService.CheckDependencies(startupCtx)
	if err != nil {
		log.Fatalf("[ai-inference.health] проверка зависимостей не прошла: %v", err)
	}

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(loggingInterceptor))
	aiinferencev1.RegisterAIInferenceServiceServer(grpcServer, transportgrpc.NewAIInferenceServiceServer(inferenceService))
	aiinferencev1.RegisterGenerationServiceServer(grpcServer, transportgrpc.NewGenerationServiceServer(inferenceService))
	aiinferencev1.RegisterEmbeddingServiceServer(grpcServer, transportgrpc.NewEmbeddingServiceServer(inferenceService))

	grpcLis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("[ai-inference.grpc] не удалось открыть порт gRPC: %v", err)
	}

	app := application.New()
	app.Add(func() error {
		return grpcServer.Serve(grpcLis)
	})

	closer.Add(grpcServer.GracefulStop)
	closer.Add(grpcLis.Close)

	log.Printf("[ai-inference.grpc] слушает порт :%s", grpcPort)

	err = app.Run()
	if err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("[ai-inference.app] приложение остановлено с ошибкой: %v", err)
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
		log.Printf("[ai-inference.grpc] метод=%s длительность=%s ошибка=%v", info.FullMethod, time.Since(startedAt), err)
		return nil, err
	}

	log.Printf("[ai-inference.grpc] метод=%s длительность=%s", info.FullMethod, time.Since(startedAt))
	return resp, nil
}
