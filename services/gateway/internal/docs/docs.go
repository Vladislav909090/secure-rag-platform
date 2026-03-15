// Package docs предоставляет Swagger UI для отображения OpenAPI-спецификации.
package docs

import (
	"fmt"
	"log"
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
const defaultSpecFileName = "gateway.swagger.json"

func resolveSpecPath() string {
	serviceLocalPath := filepath.Join("gen", "openapiv2", "v1", defaultSpecFileName)
	if _, err := os.Stat(serviceLocalPath); err == nil {
		return serviceLocalPath
	}

	workspaceRootPath := filepath.Join("services", "gateway", "gen", "openapiv2", "v1", defaultSpecFileName)
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

	return specJSON, nil
}

// RegisterAt регистрирует Swagger UI по заданному пути.
func RegisterAt(mux *http.ServeMux, serviceName string, docsPath string) {
	specJSON, err := loadSpecJSON()
	if err != nil {
		log.Printf("warning: %v", err)
		return
	}

	html := fmt.Sprintf(uiTemplate, serviceName, strconv.Quote(string(specJSON)))

	mux.HandleFunc(docsPath, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	})
}
