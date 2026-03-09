package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"

	gatewayv1 "secure-rag-platform/services/gateway/gen/v1"
	application "secure-rag-platform/services/gateway/internal/app"
	"secure-rag-platform/services/gateway/internal/closer"
	"secure-rag-platform/services/gateway/internal/config"
	"secure-rag-platform/services/gateway/internal/docs"
	transportgrpc "secure-rag-platform/services/gateway/internal/transport/grpc"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

func main() {
	port := config.GetValue(config.Port)
	if port == "" {
		port = "8080"
	}

	grpcPort := config.GetValue(config.GRPCPort)
	if grpcPort == "" {
		grpcPort = "9090"
	}

	serverImpl := &transportgrpc.GatewayServiceServerImpl{}
	grpcServer := grpc.NewServer()
	gatewayv1.RegisterGatewayServiceServer(grpcServer, serverImpl)

	grpcLis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("failed to listen grpc: %v", err)
	}

	mux := http.NewServeMux()

	gwMux := runtime.NewServeMux()
	if err := gatewayv1.RegisterGatewayServiceHandlerServer(context.Background(), gwMux, serverImpl); err != nil {
		log.Fatalf("failed to register gateway handlers: %v", err)
	}
	mux.Handle("/v1/", gwMux)

	docs.Register(mux, "Gateway")

	app := application.New()

	httpServer := &http.Server{Addr: ":" + port, Handler: mux}

	log.Printf("gateway grpc listening on :%s", grpcPort)
	log.Printf("gateway listening on :%s", port)
	log.Printf("docs: http://localhost:%s/docs", port)

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
