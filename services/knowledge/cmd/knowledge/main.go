package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"

	knowledgev1 "secure-rag-platform/services/knowledge/gen/v1"
	application "secure-rag-platform/services/knowledge/internal/app"
	"secure-rag-platform/services/knowledge/internal/closer"
	"secure-rag-platform/services/knowledge/internal/config"
	"secure-rag-platform/services/knowledge/internal/docs"
	transportgrpc "secure-rag-platform/services/knowledge/internal/transport/grpc"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

func main() {
	port := config.GetValue(config.Port)
	if port == "" {
		port = "8082"
	}

	grpcPort := config.GetValue(config.GRPCPort)
	if grpcPort == "" {
		grpcPort = "9092"
	}

	serverImpl := &transportgrpc.KnowledgeServiceServerImpl{}
	grpcServer := grpc.NewServer()
	knowledgev1.RegisterKnowledgeServiceServer(grpcServer, serverImpl)

	grpcLis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("failed to listen grpc: %v", err)
	}

	mux := http.NewServeMux()

	gwMux := runtime.NewServeMux()
	if err := knowledgev1.RegisterKnowledgeServiceHandlerServer(context.Background(), gwMux, serverImpl); err != nil {
		log.Fatalf("failed to register knowledge handlers: %v", err)
	}
	mux.Handle("/v1/", gwMux)

	docs.Register(mux, "Knowledge")

	app := application.New()

	httpServer := &http.Server{Addr: ":" + port, Handler: mux}

	log.Printf("knowledge grpc listening on :%s", grpcPort)
	log.Printf("knowledge listening on :%s", port)
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
