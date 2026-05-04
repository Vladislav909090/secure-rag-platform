package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	gatewayv1 "secure-rag-platform/services/gateway/gen/v1"
	application "secure-rag-platform/services/gateway/internal/app"
	"secure-rag-platform/services/gateway/internal/closer"
	"secure-rag-platform/services/gateway/internal/config"
	"secure-rag-platform/services/gateway/internal/docs"
	transportgrpc "secure-rag-platform/services/gateway/internal/transport/grpc"
	"secure-rag-platform/services/gateway/internal/usecase"
	iamv1 "secure-rag-platform/services/iam/gen/v1"
	knowledgev1 "secure-rag-platform/services/knowledge/gen/v1"
	ragv1 "secure-rag-platform/services/rag/gen/v1"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	logger := slog.Default()

	port := config.GetValue(config.Port)
	if port == "" {
		port = "8080"
	}

	grpcPort := config.GetValue(config.GRPCPort)
	if grpcPort == "" {
		grpcPort = "9090"
	}

	defaults := usecase.Defaults{
		TopK:                 int32(parseInt(config.GetValue(config.DefaultTopK), 3)),
		EmbeddingModelAlias:  valueOrDefault(config.GetValue(config.DefaultEmbed), "embed.default"),
		GenerationModelAlias: valueOrDefault(config.GetValue(config.DefaultGenerate), "chat.default"),
	}

	disableAuth := parseBool(config.GetValue(config.DisableAuth))
	disableFilter := parseBool(config.GetValue(config.DisableDocFilter))

	uc := buildUsecase(defaults, disableAuth, disableFilter, logger)

	serverImpl := transportgrpc.NewServer(uc)
	grpcServer := grpc.NewServer()
	gatewayv1.RegisterGatewayServiceServer(grpcServer, serverImpl)
	gatewayv1.RegisterGatewayAuthServiceServer(grpcServer, serverImpl)
	gatewayv1.RegisterGatewayKnowledgeServiceServer(grpcServer, serverImpl)
	gatewayv1.RegisterGatewayRAGServiceServer(grpcServer, serverImpl)

	grpcLis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		fatal(logger, "не удалось открыть порт gRPC", err)
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
	if err := gatewayv1.RegisterGatewayServiceHandlerServer(context.Background(), gwMux, serverImpl); err != nil {
		fatal(logger, "не удалось зарегистрировать служебные обработчики", err)
	}
	if err := gatewayv1.RegisterGatewayAuthServiceHandlerServer(context.Background(), gwMux, serverImpl); err != nil {
		fatal(logger, "не удалось зарегистрировать auth-обработчики", err)
	}
	if err := gatewayv1.RegisterGatewayKnowledgeServiceHandlerServer(context.Background(), gwMux, serverImpl); err != nil {
		fatal(logger, "не удалось зарегистрировать knowledge-обработчики", err)
	}
	if err := gatewayv1.RegisterGatewayRAGServiceHandlerServer(context.Background(), gwMux, serverImpl); err != nil {
		fatal(logger, "не удалось зарегистрировать rag-обработчики", err)
	}
	mux.Handle("/gateway/", gwMux)

	docs.RegisterAt(mux, "Gateway", "/gateway/docs")

	app := application.New()

	httpServer := &http.Server{Addr: ":" + port, Handler: mux}

	logger.Info("gRPC слушает порт", "component", "gateway.grpc", "port", grpcPort)
	logger.Info("HTTP слушает порт", "component", "gateway.http", "port", port)
	logger.Info("Swagger UI доступен", "component", "gateway.docs", "url", "http://localhost/gateway/docs")

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
		fatal(logger, "приложение остановлено с ошибкой", err)
	}
}

func buildUsecase(
	defaults usecase.Defaults,
	disableAuth bool,
	disableFilter bool,
	logger *slog.Logger,
) *usecase.Service {
	ragAddr := valueOrDefault(config.GetValue(config.RAGGRPC), "127.0.0.1:9093")
	ragConn, err := grpc.NewClient(ragAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fatal(logger, "не удалось создать gRPC-клиент RAG", err)
	}
	closer.Add(func() { _ = ragConn.Close() })

	knowledgeAddr := valueOrDefault(config.GetValue(config.KnowledgeGRPC), "127.0.0.1:9092")
	knowledgeConn, err := grpc.NewClient(knowledgeAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fatal(logger, "не удалось создать gRPC-клиент knowledge", err)
	}
	closer.Add(func() { _ = knowledgeConn.Close() })

	var iamClient iamv1.InternalIAMServiceClient
	var authClient iamv1.AuthServiceClient
	if !disableAuth {
		iamAddr := valueOrDefault(config.GetValue(config.IAMGRPC), "127.0.0.1:9091")
		iamConn, connErr := grpc.NewClient(iamAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if connErr != nil {
			fatal(logger, "не удалось создать gRPC-клиент IAM", connErr)
		}
		closer.Add(func() { _ = iamConn.Close() })
		iamClient = iamv1.NewInternalIAMServiceClient(iamConn)
		authClient = iamv1.NewAuthServiceClient(iamConn)
	}

	return usecase.NewService(
		ragv1.NewRAGServiceClient(ragConn),
		knowledgev1.NewKnowledgeServiceClient(knowledgeConn),
		iamClient,
		authClient,
		usecase.NewOPAAuthorizer(config.GetValue(config.OPAURL)),
		defaults,
		disableAuth,
		disableFilter,
		logger,
	)
}

func fatal(logger *slog.Logger, message string, err error) {
	logger.Error(message, "error", err)
	os.Exit(1)
}

func valueOrDefault(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func parseBool(raw string) bool {
	flag, err := strconv.ParseBool(raw)
	if err != nil {
		return false
	}
	return flag
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
