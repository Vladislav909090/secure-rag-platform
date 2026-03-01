package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"example.com/project/gateway/internal/docs"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)

	if spec, err := os.ReadFile("/etc/swagger.json"); err == nil {
		docs.Register(mux, "Gateway", spec)
	} else {
		log.Printf("warning: swagger spec not loaded: %v", err)
	}

	// TODO: подключить grpc-gateway runtime mux для маршрутизации HTTP -> gRPC
	// на базе сгенерированных stubs из gen/v1 (make proto:gen:gateway).
	// Пример:
	//   import gatewayv1 "example.com/project/gateway/gen/v1"
	//   gwMux := runtime.NewServeMux()
	//   gatewayv1.RegisterGatewayServiceHandlerFromEndpoint(ctx, gwMux, grpcEndpoint, opts)
	//   mux.Handle("/v1/", gwMux)

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
