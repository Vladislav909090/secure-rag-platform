package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"example.com/project/rag/internal/docs"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)

	if spec, err := os.ReadFile("/etc/swagger.json"); err == nil {
		docs.Register(mux, "RAG", spec)
	} else {
		log.Printf("warning: swagger spec not loaded: %v", err)
	}

	// TODO: подключить gRPC сервер и grpc-gateway для rag.
	// На базе сгенерированных stubs из gen/v1 (make proto:gen:rag).
	// Пример:
	//   import ragv1 "example.com/project/rag/gen/v1"
	//   grpcServer := grpc.NewServer()
	//   ragv1.RegisterRAGServiceServer(grpcServer, &server{})

	log.Printf("rag listening on :%s", port)
	log.Printf("docs: http://localhost:%s/docs", port)

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "rag"})
}
