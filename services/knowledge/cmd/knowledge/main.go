package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
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
			log.Fatalf("[knowledge.db] не удалось подключиться к PostgreSQL: %v", err)
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
			log.Fatalf("[knowledge.s3] не удалось инициализировать S3-хранилище: %v", err)
		}

		if err := s3Store.EnsureBucket(context.Background()); err != nil {
			log.Fatalf("[knowledge.s3] не удалось подготовить bucket: %v", err)
		}

		repo := repository.NewRepo(pool)
		uc = usecase.NewDocumentUsecase(repo, s3Store, maxFileSize)

		log.Println("[knowledge.app] PostgreSQL и S3 настроены")
	} else {
		log.Println("[knowledge.db] DATABASE_DSN не задан, документные ручки недоступны")
	}

	// --- gRPC-сервер ---

	serverImpl := transportgrpc.NewKnowledgeServiceServer(uc)
	grpcServer := grpc.NewServer()
	knowledgev1.RegisterKnowledgeServiceServer(grpcServer, serverImpl)

	grpcLis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("[knowledge.grpc] не удалось открыть порт gRPC: %v", err)
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
		log.Fatalf("[knowledge.http] не удалось зарегистрировать knowledge-обработчики: %v", err)
	}

	grpcConn, err := grpc.NewClient("127.0.0.1:"+grpcPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("[knowledge.http-upload] не удалось создать локальный gRPC-клиент: %v", err)
	}
	closer.Add(func() { _ = grpcConn.Close() })

	uploadHandlers := transporthttpupload.New(knowledgev1.NewKnowledgeServiceClient(grpcConn), uc)
	mux.HandleFunc("/knowledge/api/v1/documents", uploadHandlers.CreateDocument(gwMux))
	mux.HandleFunc("/knowledge/api/v1/documents/", uploadHandlers.UploadVersion(gwMux))
	mux.Handle("/knowledge/api/", gwMux)

	docs.RegisterAt(mux, "Knowledge", "/knowledge/docs")

	// --- Запуск ---

	app := application.New()

	httpServer := &http.Server{Addr: ":" + port, Handler: mux}

	log.Printf("[knowledge.grpc] слушает порт :%s", grpcPort)
	log.Printf("[knowledge.http] слушает порт :%s", port)
	log.Printf("[knowledge.docs] Swagger UI: http://localhost/knowledge/docs")

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
		log.Fatalf("[knowledge.app] приложение остановлено с ошибкой: %v", err)
	}
}
