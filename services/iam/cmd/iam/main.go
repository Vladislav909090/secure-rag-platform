package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"

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
	logger := slog.Default()
	ctx := context.Background()

	pool, err := connectPostgres(ctx)
	if err != nil {
		fatal(logger, "не удалось подготовить PostgreSQL", err)
	}
	closer.Add(func() { pool.Close() })

	repo := repository.NewRepo(pool)
	redisClient := connectRedis(ctx, logger)
	uc := usecase.NewIAMUsecase(repo, redisClient, usecaseConfig(), logger)

	err = bootstrapSuperAdmin(ctx, uc, logger)
	if err != nil {
		fatal(logger, "не удалось подготовить начального администратора", err)
	}

	servers := newServerSet(uc)
	grpcServer := newGRPCServer(servers)

	grpcPort := valueOrDefault(config.GetValue(config.GRPCPort), defaultGRPCPort)
	grpcLis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		fatal(logger, "не удалось открыть порт gRPC", err)
	}

	mux := newHTTPMux(ctx, servers, logger)
	port := valueOrDefault(config.GetValue(config.Port), defaultHTTPPort)
	httpServer := &http.Server{Addr: ":" + port, Handler: mux}

	app := application.New()
	registerRunners(app, grpcServer, grpcLis, httpServer)

	logger.Info("gRPC слушает порт", "component", "iam.grpc", "port", grpcPort)
	logger.Info("HTTP слушает порт", "component", "iam.http", "port", port)
	logger.Info("Swagger UI доступен", "component", "iam.docs", "url", "http://localhost/iam/docs")

	if err := app.Run(); err != nil {
		fatal(logger, "приложение остановлено с ошибкой", err)
	}
}

func fatal(logger *slog.Logger, message string, err error) {
	logger.Error(message, "error", err)
	os.Exit(1)
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

func connectRedis(ctx context.Context, logger *slog.Logger) *redis.Client {
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
		logger.WarnContext(ctx, "Redis недоступен, кеш и rate limit отключены",
			"component", "iam.cache",
			"error", err,
		)
		_ = client.Close()
		return nil
	}

	closer.Add(func() { _ = client.Close() })
	logger.InfoContext(ctx, "Redis подключен", "component", "iam.cache", "addr", redisAddr)
	return client
}

func usecaseConfig() usecase.Config {
	ucCfg := usecase.DefaultConfig()
	ucCfg.JWTSecret = config.GetValue(config.JWTSecret)
	ucCfg.JWTIssuer = config.GetValue(config.JWTIssuer)
	ucCfg.JWTAudience = config.GetValue(config.JWTAudience)
	return ucCfg
}

func bootstrapSuperAdmin(ctx context.Context, uc *usecase.IAMUsecase, logger *slog.Logger) error {
	bootstrapLogin := valueOrDefault(config.GetValue(config.BootstrapAdminLogin), defaultBootstrapAdminLogin)

	bootstrapPassword, created, err := uc.BootstrapSuperAdmin(
		ctx,
		bootstrapLogin,
		config.GetValue(config.BootstrapAdminPassword),
	)
	if err != nil {
		return errors.New("failed to bootstrap super admin: " + err.Error())
	}
	if !created {
		return nil
	}

	if bootstrapPassword != "" {
		logger.InfoContext(ctx, "создан начальный superadmin",
			"component", "iam.bootstrap",
			"login", bootstrapLogin,
			"password", bootstrapPassword,
		)
		return nil
	}

	logger.InfoContext(ctx, "создан начальный superadmin",
		"component", "iam.bootstrap",
		"login", bootstrapLogin,
	)
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

func newHTTPMux(ctx context.Context, servers serverSet, logger *slog.Logger) *http.ServeMux {
	mux := http.NewServeMux()
	gwMux := runtime.NewServeMux()

	registerGatewayHandlers(ctx, gwMux, servers, logger)
	mux.Handle("/iam/", gwMux)
	docs.RegisterAt(mux, "IAM", "/iam/docs")

	return mux
}

func registerGatewayHandlers(ctx context.Context, gwMux *runtime.ServeMux, servers serverSet, logger *slog.Logger) {
	if err := iamv1.RegisterIAMServiceHandlerServer(ctx, gwMux, servers.iam); err != nil {
		fatal(logger, "не удалось зарегистрировать IAM-обработчики", err)
	}
	if err := iamv1.RegisterAuthServiceHandlerServer(ctx, gwMux, servers.auth); err != nil {
		fatal(logger, "не удалось зарегистрировать auth-обработчики", err)
	}
	if err := iamv1.RegisterUserServiceHandlerServer(ctx, gwMux, servers.user); err != nil {
		fatal(logger, "не удалось зарегистрировать user-обработчики", err)
	}
	if err := iamv1.RegisterRoleServiceHandlerServer(ctx, gwMux, servers.role); err != nil {
		fatal(logger, "не удалось зарегистрировать role-обработчики", err)
	}
	if err := iamv1.RegisterAttributeServiceHandlerServer(ctx, gwMux, servers.attribute); err != nil {
		fatal(logger, "не удалось зарегистрировать attribute-обработчики", err)
	}
	if err := iamv1.RegisterSessionServiceHandlerServer(ctx, gwMux, servers.session); err != nil {
		fatal(logger, "не удалось зарегистрировать session-обработчики", err)
	}
	if err := iamv1.RegisterInternalIAMServiceHandlerServer(ctx, gwMux, servers.internalIAM); err != nil {
		fatal(logger, "не удалось зарегистрировать internal IAM-обработчики", err)
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
