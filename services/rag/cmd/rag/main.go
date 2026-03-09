package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"

	ragv1 "secure-rag-platform/services/rag/gen/v1"
	application "secure-rag-platform/services/rag/internal/app"
	"secure-rag-platform/services/rag/internal/closer"
	"secure-rag-platform/services/rag/internal/config"
	"secure-rag-platform/services/rag/internal/docs"
	transportgrpc "secure-rag-platform/services/rag/internal/transport/grpc"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

func main() {
	port := config.GetValue(config.Port)
	if port == "" {
		port = "8083"
	}

	grpcPort := config.GetValue(config.GRPCPort)
	if grpcPort == "" {
		grpcPort = "9093"
	}

	serverImpl := &transportgrpc.RAGServiceServerImpl{}
	grpcServer := grpc.NewServer()
	ragv1.RegisterRAGServiceServer(grpcServer, serverImpl)

	grpcLis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("failed to listen grpc: %v", err)
	}

	mux := http.NewServeMux()

	gwMux := runtime.NewServeMux()
	if err := ragv1.RegisterRAGServiceHandlerServer(context.Background(), gwMux, serverImpl); err != nil {
		log.Fatalf("failed to register rag handlers: %v", err)
	}
	mux.Handle("/v1/", gwMux)

	docs.Register(mux, "RAG")

	app := application.New()

	httpServer := &http.Server{Addr: ":" + port, Handler: mux}

	log.Printf("rag grpc listening on :%s", grpcPort)
	log.Printf("rag listening on :%s", port)
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
