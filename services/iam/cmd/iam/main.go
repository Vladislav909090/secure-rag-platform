package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"

	iamv1 "secure-rag-platform/services/iam/gen/v1"
	application "secure-rag-platform/services/iam/internal/app"
	"secure-rag-platform/services/iam/internal/closer"
	"secure-rag-platform/services/iam/internal/config"
	"secure-rag-platform/services/iam/internal/docs"
	"secure-rag-platform/services/iam/internal/repository"
	transportgrpc "secure-rag-platform/services/iam/internal/transport/grpc"
	"secure-rag-platform/services/iam/internal/usecase"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

const (
	defaultHTTPPort            = "8081"
	defaultGRPCPort            = "9091"
	defaultBootstrapAdminLogin = "superadmin"
)

func main() {
	ctx := context.Background()

	port := config.GetValue(config.Port)
	if port == "" {
		port = defaultHTTPPort
	}

	grpcPort := config.GetValue(config.GRPCPort)
	if grpcPort == "" {
		grpcPort = defaultGRPCPort
	}

	// --- инфраструктура: PostgreSQL и Redis ---

	databaseDSN := config.GetFirstValue(config.DatabaseDSN, config.LegacyDBDSN)
	if databaseDSN == "" {
		log.Fatalf("DATABASE_DSN (or DB_DSN) must be set")
	}

	pool, err := pgxpool.New(ctx, databaseDSN)
	if err != nil {
		log.Fatalf("failed to connect to PostgreSQL: %v", err)
	}
	closer.Add(func() { pool.Close() })

	var redisClient *redis.Client
	if redisAddr := config.GetValue(config.RedisAddr); redisAddr != "" {
		candidate := redis.NewClient(&redis.Options{
			Addr:     redisAddr,
			Password: config.GetValue(config.RedisPassword),
			DB:       0,
		})

		if pingErr := candidate.Ping(ctx).Err(); pingErr != nil {
			log.Printf("[iam] failed to connect to Redis, cache/rate-limit disabled: %v", pingErr)
			_ = candidate.Close()
		} else {
			redisClient = candidate
			closer.Add(func() { _ = redisClient.Close() })
			log.Printf("[iam] Redis connected: %s", redisAddr)
		}
	}

	// --- слой бизнес-логики и инициализация суперадмина ---

	ucCfg := usecase.DefaultConfig()
	ucCfg.JWTSecret = config.GetValue(config.JWTSecret)
	ucCfg.JWTIssuer = config.GetValue(config.JWTIssuer)
	ucCfg.JWTAudience = config.GetValue(config.JWTAudience)

	repo := repository.NewRepo(pool)
	uc := usecase.NewIAMUsecase(repo, redisClient, ucCfg)

	bootstrapLogin := config.GetValue(config.BootstrapAdminLogin)
	if bootstrapLogin == "" {
		bootstrapLogin = defaultBootstrapAdminLogin
	}

	bootstrapPassword, created, err := uc.BootstrapSuperAdmin(ctx, bootstrapLogin, config.GetValue(config.BootstrapAdminPassword))
	if err != nil {
		log.Fatalf("failed to bootstrap super admin: %v", err)
	}
	if created {
		if bootstrapPassword != "" {
			log.Printf("[iam] bootstrap super admin created: login=%s password=%s", bootstrapLogin, bootstrapPassword)
		} else {
			log.Printf("[iam] bootstrap super admin created: login=%s", bootstrapLogin)
		}
	}

	// --- gRPC сервер ---

	iamServiceServer := transportgrpc.NewIAMServiceServer(uc)
	authServiceServer := transportgrpc.NewAuthServiceServer(uc)
	userServiceServer := transportgrpc.NewUserServiceServer(uc)
	roleServiceServer := transportgrpc.NewRoleServiceServer(uc)
	attributeServiceServer := transportgrpc.NewAttributeServiceServer(uc)
	sessionServiceServer := transportgrpc.NewSessionServiceServer(uc)
	internalIAMServiceServer := transportgrpc.NewInternalIAMServiceServer(uc)

	grpcServer := grpc.NewServer()
	iamv1.RegisterIAMServiceServer(grpcServer, iamServiceServer)
	iamv1.RegisterAuthServiceServer(grpcServer, authServiceServer)
	iamv1.RegisterUserServiceServer(grpcServer, userServiceServer)
	iamv1.RegisterRoleServiceServer(grpcServer, roleServiceServer)
	iamv1.RegisterAttributeServiceServer(grpcServer, attributeServiceServer)
	iamv1.RegisterSessionServiceServer(grpcServer, sessionServiceServer)
	iamv1.RegisterInternalIAMServiceServer(grpcServer, internalIAMServiceServer)

	grpcLis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("failed to listen on gRPC port: %v", err)
	}

	// --- HTTP-маршрутизация ---

	mux := http.NewServeMux()

	gwMux := runtime.NewServeMux()
	if err := iamv1.RegisterIAMServiceHandlerServer(ctx, gwMux, iamServiceServer); err != nil {
		log.Fatalf("failed to register IAM handlers: %v", err)
	}
	if err := iamv1.RegisterAuthServiceHandlerServer(ctx, gwMux, authServiceServer); err != nil {
		log.Fatalf("failed to register Auth handlers: %v", err)
	}
	if err := iamv1.RegisterUserServiceHandlerServer(ctx, gwMux, userServiceServer); err != nil {
		log.Fatalf("failed to register User handlers: %v", err)
	}
	if err := iamv1.RegisterRoleServiceHandlerServer(ctx, gwMux, roleServiceServer); err != nil {
		log.Fatalf("failed to register Role handlers: %v", err)
	}
	if err := iamv1.RegisterAttributeServiceHandlerServer(ctx, gwMux, attributeServiceServer); err != nil {
		log.Fatalf("failed to register Attribute handlers: %v", err)
	}
	if err := iamv1.RegisterSessionServiceHandlerServer(ctx, gwMux, sessionServiceServer); err != nil {
		log.Fatalf("failed to register Session handlers: %v", err)
	}
	if err := iamv1.RegisterInternalIAMServiceHandlerServer(ctx, gwMux, internalIAMServiceServer); err != nil {
		log.Fatalf("failed to register InternalIAM handlers: %v", err)
	}
	mux.Handle("/iam/", gwMux)

	docs.RegisterAt(mux, "IAM", "/iam/docs")

	// --- запуск ---

	app := application.New()

	httpServer := &http.Server{Addr: ":" + port, Handler: mux}

	log.Printf("iam gRPC listening on :%s", grpcPort)
	log.Printf("iam listening on :%s", port)
	log.Printf("docs: http://localhost/iam/docs")

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
