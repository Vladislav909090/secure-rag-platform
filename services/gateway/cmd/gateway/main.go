package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"

	gatewayv1 "example.com/project/gateway/gen/v1"
	"example.com/project/gateway/internal/docs"
	transportgrpc "example.com/project/gateway/internal/transport/grpc"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	grpcPort := os.Getenv("GRPC_PORT")
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

	go func() {
		log.Printf("gateway grpc listening on :%s", grpcPort)
		if serveErr := grpcServer.Serve(grpcLis); serveErr != nil {
			log.Fatalf("failed to serve grpc: %v", serveErr)
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)

	gwMux := runtime.NewServeMux()
	if err := gatewayv1.RegisterGatewayServiceHandlerServer(context.Background(), gwMux, serverImpl); err != nil {
		log.Fatalf("failed to register gateway handlers: %v", err)
	}
	mux.Handle("/v1/", gwMux)

	if spec, err := os.ReadFile("/etc/swagger.json"); err == nil {
		docs.Register(mux, "Gateway", spec)
	} else {
		log.Printf("warning: swagger spec not loaded: %v", err)
	}

	log.Printf("gateway listening on :%s", port)
	log.Printf("docs: http://localhost:%s/docs", port)

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "gateway"})
}
