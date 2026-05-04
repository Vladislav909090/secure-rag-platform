// Package docs предоставляет Swagger UI для отображения OpenAPI-спецификации.
package docs

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

const uiTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>API Docs — %s</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
  <style>html{box-sizing:border-box}*,*::before,*::after{box-sizing:inherit}body{margin:0;background:#fafafa}</style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-standalone-preset.js"></script>
  <script>
	const spec = JSON.parse(%s);
    SwaggerUIBundle({
	  spec: spec,
      dom_id: '#swagger-ui',
      presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
      layout: 'StandaloneLayout'
    });
  </script>
</body>
</html>`

const dockerSpecPath = "/etc/swagger.json"
const defaultSpecFileName = "knowledge.swagger.json"

func resolveSpecPath() string {
	serviceLocalPath := filepath.Join("gen", "openapiv2", "v1", defaultSpecFileName)
	if _, err := os.Stat(serviceLocalPath); err == nil {
		return serviceLocalPath
	}

	workspaceRootPath := filepath.Join("services", "knowledge", "gen", "openapiv2", "v1", defaultSpecFileName)
	if _, err := os.Stat(workspaceRootPath); err == nil {
		return workspaceRootPath
	}

	return dockerSpecPath
}

func loadSpecJSON() ([]byte, error) {
	specPath := resolveSpecPath()
	specJSON, err := os.ReadFile(specPath)
	if err != nil {
		return nil, fmt.Errorf("swagger spec not loaded from %q: %w", specPath, err)
	}

	overlaid, err := applyMultipartUploadOverlay(specJSON)
	if err != nil {
		return nil, err
	}

	return overlaid, nil
}

func applyMultipartUploadOverlay(specJSON []byte) ([]byte, error) {
	var root map[string]any
	if err := json.Unmarshal(specJSON, &root); err != nil {
		return nil, fmt.Errorf("unmarshal swagger spec: %w", err)
	}

	paths, _ := root["paths"].(map[string]any)
	if paths == nil {
		paths = map[string]any{}
		root["paths"] = paths
	}

	pathDocuments, _ := paths["/knowledge/api/v1/documents"].(map[string]any)
	if pathDocuments == nil {
		pathDocuments = map[string]any{}
		paths["/knowledge/api/v1/documents"] = pathDocuments
	}
	pathDocuments["post"] = map[string]any{
		"summary":     "CreateDocument загружает файл как multipart и стримит его в gRPC",
		"operationId": "KnowledgeService_CreateDocumentMultipart",
		"tags":        []any{"KnowledgeService"},
		"consumes":    []any{"multipart/form-data"},
		"produces":    []any{"application/json"},
		"parameters": []any{
			map[string]any{"name": "title", "in": "formData", "required": true, "type": "string"},
			map[string]any{"name": "description", "in": "formData", "required": false, "type": "string"},
			map[string]any{
				"name":        "attributes",
				"in":          "formData",
				"required":    false,
				"type":        "string",
				"description": "JSON object as string",
			},
			map[string]any{"name": "file", "in": "formData", "required": true, "type": "file"},
		},
		"responses": map[string]any{
			"200": map[string]any{
				"description": "OK",
				"schema": map[string]any{
					"$ref": "#/definitions/v1CreateDocumentResponse",
				},
			},
			"400": map[string]any{"description": "Bad Request", "schema": map[string]any{"$ref": "#/definitions/rpcStatus"}},
		},
	}

	out, err := json.Marshal(root)
	if err != nil {
		return nil, fmt.Errorf("marshal swagger spec: %w", err)
	}

	return out, nil
}

// RegisterAt регистрирует Swagger UI по заданному пути.
func RegisterAt(mux *http.ServeMux, serviceName string, docsPath string) {
	specJSON, err := loadSpecJSON()
	if err != nil {
		slog.Warn("Swagger UI недоступен",
			"component", "knowledge.docs",
			"error", err,
		)
		return
	}

	html := fmt.Sprintf(uiTemplate, serviceName, strconv.Quote(string(specJSON)))

	mux.HandleFunc(docsPath, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	})
}
