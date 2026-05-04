package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net"
	"os"
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
	logger := slog.Default()

	modelsConfigPath := flag.String("config", config.DefaultModelsConfigPath, "Path to models config JSON file")
	flag.Parse()

	runtimeCfg, err := config.LoadFromFile(*modelsConfigPath)
	if err != nil {
		fatal(logger, "не удалось загрузить конфигурацию моделей", err)
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
			fatal(logger, "не удалось разобрать provider_timeout", err)
		}
		providerTimeout = parsed
	}

	providerSet := []usecase.Provider{
		provider.NewOpenAICompatProvider(providerTimeout),
	}
	inferenceService := usecase.NewService(runtimeCfg.ModelAliases, providerSet, logger)

	startupCheckTimeout := providerTimeout + 5*time.Second
	startupCtx, cancel := context.WithTimeout(context.Background(), startupCheckTimeout)
	defer cancel()
	err = inferenceService.CheckDependencies(startupCtx)
	if err != nil {
		fatal(logger, "проверка зависимостей не прошла", err)
	}

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(loggingInterceptor(logger)))
	aiinferencev1.RegisterAIInferenceServiceServer(grpcServer, transportgrpc.NewAIInferenceServiceServer(inferenceService))
	aiinferencev1.RegisterGenerationServiceServer(grpcServer, transportgrpc.NewGenerationServiceServer(inferenceService))
	aiinferencev1.RegisterEmbeddingServiceServer(grpcServer, transportgrpc.NewEmbeddingServiceServer(inferenceService))

	grpcLis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		fatal(logger, "не удалось открыть порт gRPC", err)
	}

	app := application.New()
	app.Add(func() error {
		return grpcServer.Serve(grpcLis)
	})

	closer.Add(grpcServer.GracefulStop)
	closer.Add(grpcLis.Close)

	logger.Info("gRPC слушает порт", "component", "ai-inference.grpc", "port", grpcPort)

	err = app.Run()
	if err != nil && !errors.Is(err, context.Canceled) {
		fatal(logger, "приложение остановлено с ошибкой", err)
	}
}

func fatal(logger *slog.Logger, message string, err error) {
	logger.Error(message, "error", err)
	os.Exit(1)
}

func loggingInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		startedAt := time.Now()
		resp, err := handler(ctx, req)
		if err != nil {
			logger.ErrorContext(ctx, "gRPC метод завершился с ошибкой",
				"component", "ai-inference.grpc",
				"method", info.FullMethod,
				"duration", time.Since(startedAt),
				"error", err,
			)
			return nil, err
		}

		logger.InfoContext(ctx, "gRPC метод выполнен",
			"component", "ai-inference.grpc",
			"method", info.FullMethod,
			"duration", time.Since(startedAt),
		)
		return resp, nil
	}
}
