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

type serverSet struct {
	iam         *transportgrpc.IAMServiceServerImpl
	auth        *transportgrpc.AuthServiceServerImpl
	user        *transportgrpc.UserServiceServerImpl
	role        *transportgrpc.RoleServiceServerImpl
	attribute   *transportgrpc.AttributeServiceServerImpl
	session     *transportgrpc.SessionServiceServerImpl
	internalIAM *transportgrpc.InternalIAMServiceServerImpl
}

func main() {
	ctx := context.Background()

	pool, err := connectPostgres(ctx)
	if err != nil {
		log.Fatal(err)
	}
	closer.Add(func() { pool.Close() })

	repo := repository.NewRepo(pool)
	redisClient := connectRedis(ctx)
	uc := usecase.NewIAMUsecase(repo, redisClient, usecaseConfig())

	err = bootstrapSuperAdmin(ctx, uc)
	if err != nil {
		log.Fatal(err)
	}

	servers := newServerSet(uc)
	grpcServer := newGRPCServer(servers)

	grpcPort := valueOrDefault(config.GetValue(config.GRPCPort), defaultGRPCPort)
	grpcLis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("failed to listen on gRPC port: %v", err)
	}

	mux := newHTTPMux(ctx, servers)
	port := valueOrDefault(config.GetValue(config.Port), defaultHTTPPort)
	httpServer := &http.Server{Addr: ":" + port, Handler: mux}

	app := application.New()
	registerRunners(app, grpcServer, grpcLis, httpServer)

	log.Printf("iam gRPC listening on :%s", grpcPort)
	log.Printf("iam listening on :%s", port)
	log.Printf("docs: http://localhost/iam/docs")

	if err := app.Run(); err != nil {
		log.Fatalf("application stopped with error: %v", err)
	}
}

func valueOrDefault(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func connectPostgres(ctx context.Context) (*pgxpool.Pool, error) {
	databaseDSN := config.GetFirstValue(config.DatabaseDSN, config.LegacyDBDSN)
	if databaseDSN == "" {
		return nil, errors.New("DATABASE_DSN (or DB_DSN) must be set")
	}

	pool, err := pgxpool.New(ctx, databaseDSN)
	if err != nil {
		return nil, errors.New("failed to connect to PostgreSQL: " + err.Error())
	}
	return pool, nil
}

func connectRedis(ctx context.Context) *redis.Client {
	redisAddr := config.GetValue(config.RedisAddr)
	if redisAddr == "" {
		return nil
	}

	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: config.GetValue(config.RedisPassword),
		DB:       0,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("[iam] failed to connect to Redis, cache/rate-limit disabled: %v", err)
		_ = client.Close()
		return nil
	}

	closer.Add(func() { _ = client.Close() })
	log.Printf("[iam] Redis connected: %s", redisAddr)
	return client
}

func usecaseConfig() usecase.Config {
	ucCfg := usecase.DefaultConfig()
	ucCfg.JWTSecret = config.GetValue(config.JWTSecret)
	ucCfg.JWTIssuer = config.GetValue(config.JWTIssuer)
	ucCfg.JWTAudience = config.GetValue(config.JWTAudience)
	return ucCfg
}

func bootstrapSuperAdmin(ctx context.Context, uc *usecase.IAMUsecase) error {
	bootstrapLogin := valueOrDefault(config.GetValue(config.BootstrapAdminLogin), defaultBootstrapAdminLogin)

	bootstrapPassword, created, err := uc.BootstrapSuperAdmin(ctx, bootstrapLogin, config.GetValue(config.BootstrapAdminPassword))
	if err != nil {
		return errors.New("failed to bootstrap super admin: " + err.Error())
	}
	if !created {
		return nil
	}

	if bootstrapPassword != "" {
		log.Printf("[iam] bootstrap super admin created: login=%s password=%s", bootstrapLogin, bootstrapPassword)
		return nil
	}

	log.Printf("[iam] bootstrap super admin created: login=%s", bootstrapLogin)
	return nil
}

func newServerSet(uc *usecase.IAMUsecase) serverSet {
	return serverSet{
		iam:         transportgrpc.NewIAMServiceServer(uc),
		auth:        transportgrpc.NewAuthServiceServer(uc),
		user:        transportgrpc.NewUserServiceServer(uc),
		role:        transportgrpc.NewRoleServiceServer(uc),
		attribute:   transportgrpc.NewAttributeServiceServer(uc),
		session:     transportgrpc.NewSessionServiceServer(uc),
		internalIAM: transportgrpc.NewInternalIAMServiceServer(uc),
	}
}

func newGRPCServer(servers serverSet) *grpc.Server {
	grpcServer := grpc.NewServer()
	iamv1.RegisterIAMServiceServer(grpcServer, servers.iam)
	iamv1.RegisterAuthServiceServer(grpcServer, servers.auth)
	iamv1.RegisterUserServiceServer(grpcServer, servers.user)
	iamv1.RegisterRoleServiceServer(grpcServer, servers.role)
	iamv1.RegisterAttributeServiceServer(grpcServer, servers.attribute)
	iamv1.RegisterSessionServiceServer(grpcServer, servers.session)
	iamv1.RegisterInternalIAMServiceServer(grpcServer, servers.internalIAM)
	return grpcServer
}

func newHTTPMux(ctx context.Context, servers serverSet) *http.ServeMux {
	mux := http.NewServeMux()
	gwMux := runtime.NewServeMux()

	registerGatewayHandlers(ctx, gwMux, servers)
	mux.Handle("/iam/", gwMux)
	docs.RegisterAt(mux, "IAM", "/iam/docs")

	return mux
}

func registerGatewayHandlers(ctx context.Context, gwMux *runtime.ServeMux, servers serverSet) {
	if err := iamv1.RegisterIAMServiceHandlerServer(ctx, gwMux, servers.iam); err != nil {
		log.Fatalf("failed to register IAM handlers: %v", err)
	}
	if err := iamv1.RegisterAuthServiceHandlerServer(ctx, gwMux, servers.auth); err != nil {
		log.Fatalf("failed to register Auth handlers: %v", err)
	}
	if err := iamv1.RegisterUserServiceHandlerServer(ctx, gwMux, servers.user); err != nil {
		log.Fatalf("failed to register User handlers: %v", err)
	}
	if err := iamv1.RegisterRoleServiceHandlerServer(ctx, gwMux, servers.role); err != nil {
		log.Fatalf("failed to register Role handlers: %v", err)
	}
	if err := iamv1.RegisterAttributeServiceHandlerServer(ctx, gwMux, servers.attribute); err != nil {
		log.Fatalf("failed to register Attribute handlers: %v", err)
	}
	if err := iamv1.RegisterSessionServiceHandlerServer(ctx, gwMux, servers.session); err != nil {
		log.Fatalf("failed to register Session handlers: %v", err)
	}
	if err := iamv1.RegisterInternalIAMServiceHandlerServer(ctx, gwMux, servers.internalIAM); err != nil {
		log.Fatalf("failed to register InternalIAM handlers: %v", err)
	}
}

func registerRunners(app *application.App, grpcServer *grpc.Server, grpcLis net.Listener, httpServer *http.Server) {
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
}
