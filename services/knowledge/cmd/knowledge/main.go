package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"

	knowledgev1 "secure-rag-platform/services/knowledge/gen/v1"
	application "secure-rag-platform/services/knowledge/internal/app"
	"secure-rag-platform/services/knowledge/internal/closer"
	"secure-rag-platform/services/knowledge/internal/config"
	"secure-rag-platform/services/knowledge/internal/docs"
	"secure-rag-platform/services/knowledge/internal/repository"
	"secure-rag-platform/services/knowledge/internal/storage"
	transportgrpc "secure-rag-platform/services/knowledge/internal/transport/grpc"
	transporthttpupload "secure-rag-platform/services/knowledge/internal/transport/httpupload"
	"secure-rag-platform/services/knowledge/internal/usecase"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
)

const defaultMaxFileSize = 100 * 1024 * 1024 // 100 MB

func main() {
	logger := slog.Default()

	port := config.GetValue(config.Port)
	if port == "" {
		port = "8082"
	}

	grpcPort := config.GetValue(config.GRPCPort)
	if grpcPort == "" {
		grpcPort = "9092"
	}

	maxFileSize := int64(defaultMaxFileSize)
	if v := config.GetValue(config.MaxFileSize); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			maxFileSize = n
		}
	}

	// --- Опциональная инфраструктура: БД + S3 ---

	var uc *usecase.DocumentUsecase

	if dbDSN := config.GetValue(config.DatabaseDSN); dbDSN != "" {
		pool, err := pgxpool.New(context.Background(), dbDSN)
		if err != nil {
			fatal(logger, "не удалось подключиться к PostgreSQL", err)
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
			fatal(logger, "не удалось инициализировать S3-хранилище", err)
		}

		if err := s3Store.EnsureBucket(context.Background()); err != nil {
			fatal(logger, "не удалось подготовить bucket", err)
		}

		repo := repository.NewRepo(pool)
		uc = usecase.NewDocumentUsecase(repo, s3Store, maxFileSize)

		logger.Info("PostgreSQL и S3 настроены", "component", "knowledge.app")
	} else {
		logger.Warn("DATABASE_DSN не задан, документные ручки недоступны", "component", "knowledge.db")
	}

	// --- gRPC-сервер ---

	serverImpl := transportgrpc.NewKnowledgeServiceServer(uc)
	grpcServer := grpc.NewServer()
	knowledgev1.RegisterKnowledgeServiceServer(grpcServer, serverImpl)

	grpcLis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		fatal(logger, "не удалось открыть порт gRPC", err)
	}

	// --- HTTP mux ---

	mux := http.NewServeMux()

	// gRPC-gateway
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
	err = knowledgev1.RegisterKnowledgeServiceHandlerServer(context.Background(), gwMux, serverImpl)
	if err != nil {
		fatal(logger, "не удалось зарегистрировать knowledge-обработчики", err)
	}

	grpcConn, err := grpc.NewClient("127.0.0.1:"+grpcPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fatal(logger, "не удалось создать локальный gRPC-клиент", err)
	}
	closer.Add(func() { _ = grpcConn.Close() })

	uploadHandlers := transporthttpupload.New(knowledgev1.NewKnowledgeServiceClient(grpcConn), uc)
	mux.HandleFunc("/knowledge/api/v1/documents", uploadHandlers.CreateDocument(gwMux))
	mux.HandleFunc("/knowledge/api/v1/documents/", uploadHandlers.DocumentFiles(gwMux))
	mux.Handle("/knowledge/api/", gwMux)

	docs.RegisterAt(mux, "Knowledge", "/knowledge/docs")

	// --- Запуск ---

	app := application.New()

	httpServer := &http.Server{Addr: ":" + port, Handler: mux}

	logger.Info("gRPC слушает порт", "component", "knowledge.grpc", "port", grpcPort)
	logger.Info("HTTP слушает порт", "component", "knowledge.http", "port", port)
	logger.Info("Swagger UI доступен", "component", "knowledge.docs", "url", "http://localhost/knowledge/docs")

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

func fatal(logger *slog.Logger, message string, err error) {
	logger.Error(message, "error", err)
	os.Exit(1)
}
