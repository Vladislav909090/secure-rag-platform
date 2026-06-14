package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	aiinferencev1 "secure-rag-platform/api/gen/go/aiinference/v1"
	application "secure-rag-platform/services/ai-inference/internal/app"
	"secure-rag-platform/services/ai-inference/internal/closer"
	"secure-rag-platform/services/ai-inference/internal/config"
	"secure-rag-platform/services/ai-inference/internal/docs"
	"secure-rag-platform/services/ai-inference/internal/provider"
	transportgrpc "secure-rag-platform/services/ai-inference/internal/transport/grpc"
	"secure-rag-platform/services/ai-inference/internal/usecase"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
)

const defaultGRPCPort = "9094"
const defaultHTTPPort = "8084"

func main() {
	logger := slog.Default()

	runtimeCfg, err := config.Load()
	if err != nil {
		fatal(logger, "не удалось загрузить конфигурацию моделей", err)
	}

	grpcPort := runtimeCfg.GRPCPort
	if grpcPort == "" {
		grpcPort = defaultGRPCPort
	}

	httpPort := runtimeCfg.HTTPPort
	if httpPort == "" {
		httpPort = defaultHTTPPort
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

	aiServerImpl := transportgrpc.NewAIInferenceServiceServer(inferenceService)
	genServerImpl := transportgrpc.NewGenerationServiceServer(inferenceService)
	embServerImpl := transportgrpc.NewEmbeddingServiceServer(inferenceService)

	aiinferencev1.RegisterAIInferenceServiceServer(grpcServer, aiServerImpl)
	aiinferencev1.RegisterGenerationServiceServer(grpcServer, genServerImpl)
	aiinferencev1.RegisterEmbeddingServiceServer(grpcServer, embServerImpl)

	grpcLis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		fatal(logger, "не удалось открыть порт gRPC", err)
	}

	// gRPC-gateway (HTTP -> gRPC)
	mux := http.NewServeMux()
	docs.RegisterAt(mux, "AI Inference", "/ai-inference/docs")
	gatewayMux := runtime.NewServeMux(
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

	gwCtx := context.Background()
	if err = aiinferencev1.RegisterAIInferenceServiceHandlerServer(gwCtx, gatewayMux, aiServerImpl); err != nil {
		fatal(logger, "не удалось зарегистрировать AIInferenceService HTTP handlers", err)
	}
	if err = aiinferencev1.RegisterGenerationServiceHandlerServer(gwCtx, gatewayMux, genServerImpl); err != nil {
		fatal(logger, "не удалось зарегистрировать GenerationService HTTP handlers", err)
	}
	if err = aiinferencev1.RegisterEmbeddingServiceHandlerServer(gwCtx, gatewayMux, embServerImpl); err != nil {
		fatal(logger, "не удалось зарегистрировать EmbeddingService HTTP handlers", err)
	}

	mux.Handle("/ai-inference/", gatewayMux)
	httpServer := &http.Server{Addr: ":" + httpPort, Handler: mux}

	app := application.New()
	app.Add(func() error {
		return grpcServer.Serve(grpcLis)
	})
	app.Add(func() error {
		if err = httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})

	closer.Add(httpServer.Shutdown)
	closer.Add(grpcServer.GracefulStop)
	closer.Add(grpcLis.Close)

	logger.Info("gRPC слушает порт", "component", "ai-inference.grpc", "port", grpcPort)
	logger.Info("HTTP слушает порт", "component", "ai-inference.http", "port", httpPort)
	logger.Info("Swagger UI доступен", "component", "ai-inference.docs", "url", "http://localhost/ai-inference/docs")

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
