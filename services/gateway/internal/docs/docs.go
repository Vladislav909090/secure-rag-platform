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

	// Добавляем Bearer auth schema для сгенерированных спецификаций без securityDefinitions.
	if (!spec.securityDefinitions) {
	  spec.securityDefinitions = {};
	}
	if (!spec.securityDefinitions.BearerAuth) {
	  spec.securityDefinitions.BearerAuth = {
		type: 'apiKey',
		name: 'Authorization',
		in: 'header',
		description: 'Bearer access token, пример: Bearer eyJ...'
	  };
	}
	if (!spec.security) {
	  spec.security = [{ BearerAuth: [] }];
	}

	const ensureBearer = (value) => {
	  if (!value || typeof value !== 'string') {
		return value;
	  }
	  const trimmed = value.trim();
	  if (!trimmed) {
		return trimmed;
	  }
	  if (/^Bearer\s+/i.test(trimmed)) {
		return trimmed;
	  }
	  return 'Bearer ' + trimmed;
	};

    SwaggerUIBundle({
	  spec: spec,
      dom_id: '#swagger-ui',
      presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
	  persistAuthorization: true,
	  requestInterceptor: (req) => {
		if (req && req.headers && req.headers.Authorization) {
		  req.headers.Authorization = ensureBearer(req.headers.Authorization);
		}
		return req;
	  },
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

	pathDocuments, _ := paths["/gateway/api/v1/documents"].(map[string]any)
	if pathDocuments == nil {
		pathDocuments = map[string]any{}
		paths["/gateway/api/v1/documents"] = pathDocuments
	}
	pathDocuments["post"] = map[string]any{
		"summary":     "CreateDocument загружает файл как multipart через gateway",
		"operationId": "GatewayKnowledgeService_CreateDocumentMultipart",
		"tags":        []any{"GatewayKnowledgeService"},
		"security":    []any{map[string]any{"BearerAuth": []any{}}},
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
					"$ref": "#/definitions/v1GetDocumentResponse",
				},
			},
			"400": map[string]any{"description": "Bad Request", "schema": map[string]any{"$ref": "#/definitions/rpcStatus"}},
			"401": map[string]any{"description": "Unauthorized", "schema": map[string]any{"$ref": "#/definitions/rpcStatus"}},
			"403": map[string]any{"description": "Forbidden", "schema": map[string]any{"$ref": "#/definitions/rpcStatus"}},
		},
	}

	pathDocument, _ := paths["/gateway/api/v1/documents/{document_uuid}"].(map[string]any)
	if pathDocument == nil {
		pathDocument = map[string]any{}
		paths["/gateway/api/v1/documents/{document_uuid}"] = pathDocument
	}
	pathDocument["patch"] = jsonOperation(
		"UpdateDocument обновляет title и/или description документа",
		"GatewayKnowledgeService_UpdateDocument",
		"GatewayKnowledgeService",
		[]any{
			pathParam("document_uuid"),
			bodyParam(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"title":       map[string]any{"type": "string"},
					"description": map[string]any{"type": "string"},
				},
			}),
		},
		map[string]any{"$ref": "#/definitions/v1GetDocumentResponse"},
	)

	pathDocumentAttrs, _ := paths["/gateway/api/v1/documents/{document_uuid}/attributes"].(map[string]any)
	if pathDocumentAttrs == nil {
		pathDocumentAttrs = map[string]any{}
		paths["/gateway/api/v1/documents/{document_uuid}/attributes"] = pathDocumentAttrs
	}
	pathDocumentAttrs["patch"] = jsonOperation(
		"UpdateDocumentAttributes заменяет attributes документа",
		"GatewayKnowledgeService_UpdateDocumentAttributes",
		"GatewayKnowledgeService",
		[]any{
			pathParam("document_uuid"),
			bodyParam(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"attributes": map[string]any{"type": "object"},
				},
			}),
		},
		map[string]any{"$ref": "#/definitions/v1GetDocumentResponse"},
	)

	pathAdminRoles, _ := paths["/gateway/api/v1/admin/roles"].(map[string]any)
	if pathAdminRoles == nil {
		pathAdminRoles = map[string]any{}
		paths["/gateway/api/v1/admin/roles"] = pathAdminRoles
	}
	pathAdminRoles["get"] = jsonOperation(
		"ListRoles возвращает роли IAM через gateway",
		"GatewayAdmin_ListRoles",
		"GatewayAdmin",
		nil,
		map[string]any{"type": "object"},
	)

	pathAdminUser, _ := paths["/gateway/api/v1/admin/users/{user_id}"].(map[string]any)
	if pathAdminUser == nil {
		pathAdminUser = map[string]any{}
		paths["/gateway/api/v1/admin/users/{user_id}"] = pathAdminUser
	}
	pathAdminUser["patch"] = jsonOperation(
		"UpdateUser обновляет пользователя IAM через gateway",
		"GatewayAdmin_UpdateUser",
		"GatewayAdmin",
		[]any{
			pathParam("user_id"),
			bodyParam(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"login":     map[string]any{"type": "string"},
					"password":  map[string]any{"type": "string"},
					"is_active": map[string]any{"type": "boolean"},
				},
			}),
		},
		map[string]any{"type": "object"},
	)

	pathAdminUserAttrs, _ := paths["/gateway/api/v1/admin/users/{user_id}/attributes"].(map[string]any)
	if pathAdminUserAttrs == nil {
		pathAdminUserAttrs = map[string]any{}
		paths["/gateway/api/v1/admin/users/{user_id}/attributes"] = pathAdminUserAttrs
	}
	pathAdminUserAttrs["put"] = jsonOperation(
		"ReplaceUserAttributes заменяет attributes пользователя IAM через gateway",
		"GatewayAdmin_ReplaceUserAttributes",
		"GatewayAdmin",
		[]any{
			pathParam("user_id"),
			bodyParam(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"attributes": map[string]any{"type": "object"},
				},
			}),
		},
		map[string]any{"type": "object"},
	)

	pathAdminUserRoles, _ := paths["/gateway/api/v1/admin/users/{user_id}/roles"].(map[string]any)
	if pathAdminUserRoles == nil {
		pathAdminUserRoles = map[string]any{}
		paths["/gateway/api/v1/admin/users/{user_id}/roles"] = pathAdminUserRoles
	}
	pathAdminUserRoles["put"] = jsonOperation(
		"SetUserRoles полностью заменяет роли пользователя IAM через gateway",
		"GatewayAdmin_SetUserRoles",
		"GatewayAdmin",
		[]any{
			pathParam("user_id"),
			bodyParam(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"role_codes": map[string]any{
						"type":  "array",
						"items": map[string]any{"type": "string"},
					},
				},
			}),
		},
		map[string]any{"type": "object"},
	)

	out, err := json.Marshal(root)
	if err != nil {
		return nil, fmt.Errorf("marshal swagger spec: %w", err)
	}

	return out, nil
}

func jsonOperation(summary string, operationID string, tag string, params []any, responseSchema map[string]any) map[string]any {
	if params == nil {
		params = []any{}
	}
	return map[string]any{
		"summary":     summary,
		"operationId": operationID,
		"tags":        []any{tag},
		"security":    []any{map[string]any{"BearerAuth": []any{}}},
		"consumes":    []any{"application/json"},
		"produces":    []any{"application/json"},
		"parameters":  params,
		"responses": map[string]any{
			"200": map[string]any{"description": "OK", "schema": responseSchema},
			"400": map[string]any{"description": "Bad Request", "schema": map[string]any{"$ref": "#/definitions/rpcStatus"}},
			"401": map[string]any{"description": "Unauthorized", "schema": map[string]any{"$ref": "#/definitions/rpcStatus"}},
			"403": map[string]any{"description": "Forbidden", "schema": map[string]any{"$ref": "#/definitions/rpcStatus"}},
			"404": map[string]any{"description": "Not Found", "schema": map[string]any{"$ref": "#/definitions/rpcStatus"}},
		},
	}
}

func pathParam(name string) map[string]any {
	return map[string]any{"name": name, "in": "path", "required": true, "type": "string"}
}

func bodyParam(schema map[string]any) map[string]any {
	return map[string]any{"name": "body", "in": "body", "required": true, "schema": schema}
}

// RegisterAt регистрирует Swagger UI по заданному пути.
func RegisterAt(mux *http.ServeMux, serviceName string, docsPath string) {
	specJSON, err := loadSpecJSON()
	if err != nil {
		slog.Warn("Swagger UI недоступен",
			"component", "gateway.docs",
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
