package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"example.com/project/knowledge/internal/docs"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)

	if spec, err := os.ReadFile("/etc/swagger.json"); err == nil {
		docs.Register(mux, "Knowledge", spec)
	} else {
		log.Printf("warning: swagger spec not loaded: %v", err)
	}

	// TODO: подключить gRPC сервер и grpc-gateway для knowledge.
	// На базе сгенерированных stubs из gen/v1 (make proto:gen:knowledge).
	// Пример:
	//   import knowledgev1 "example.com/project/knowledge/gen/v1"
	//   grpcServer := grpc.NewServer()
	//   knowledgev1.RegisterKnowledgeServiceServer(grpcServer, &server{})

	log.Printf("knowledge listening on :%s", port)
	log.Printf("docs: http://localhost:%s/docs", port)

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "knowledge"})
}
