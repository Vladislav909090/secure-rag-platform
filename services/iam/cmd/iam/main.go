package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"example.com/project/iam/internal/docs"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)

	if spec, err := os.ReadFile("/etc/swagger.json"); err == nil {
		docs.Register(mux, "IAM", spec)
	} else {
		log.Printf("warning: swagger spec not loaded: %v", err)
	}

	// TODO: подключить gRPC сервер и grpc-gateway для iam.
	// На базе сгенерированных stubs из gen/v1 (make proto:gen:iam).
	// Пример:
	//   import iamv1 "example.com/project/iam/gen/v1"
	//   grpcServer := grpc.NewServer()
	//   iamv1.RegisterIAMServiceServer(grpcServer, &server{})

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
