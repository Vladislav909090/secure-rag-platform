package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"

	iamv1 "example.com/project/iam/gen/v1"
	"example.com/project/iam/internal/docs"
	transportgrpc "example.com/project/iam/internal/transport/grpc"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "9091"
	}

	serverImpl := &transportgrpc.IAMServiceServerImpl{}
	grpcServer := grpc.NewServer()
	iamv1.RegisterIAMServiceServer(grpcServer, serverImpl)

	grpcLis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("failed to listen grpc: %v", err)
	}

	go func() {
		log.Printf("iam grpc listening on :%s", grpcPort)
		if serveErr := grpcServer.Serve(grpcLis); serveErr != nil {
			log.Fatalf("failed to serve grpc: %v", serveErr)
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)

	gwMux := runtime.NewServeMux()
	if err := iamv1.RegisterIAMServiceHandlerServer(context.Background(), gwMux, serverImpl); err != nil {
		log.Fatalf("failed to register iam handlers: %v", err)
	}
	mux.Handle("/v1/", gwMux)

	if spec, err := os.ReadFile("/etc/swagger.json"); err == nil {
		docs.Register(mux, "IAM", spec)
	} else {
		log.Printf("warning: swagger spec not loaded: %v", err)
	}

	log.Printf("iam listening on :%s", port)
	log.Printf("docs: http://localhost:%s/docs", port)

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "iam"})
}
