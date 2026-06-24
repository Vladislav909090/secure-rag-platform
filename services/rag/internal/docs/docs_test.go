package docs

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveSpecPathPrefersWorkspacePath(t *testing.T) {
	tmp := t.TempDir()
	t.Chdir(tmp)
	specDir := filepath.Join(tmp, "services", "rag", "gen", "openapiv2", "v1")
	require.NoError(t, os.MkdirAll(specDir, 0o755))
	specPath := filepath.Join(specDir, defaultSpecFileName)
	require.NoError(t, os.WriteFile(specPath, []byte(`{"openapi":"3.0.0"}`), 0o644))

	assert.Equal(t, filepath.Join("services", "rag", "gen", "openapiv2", "v1", defaultSpecFileName), resolveSpecPath())
}

func TestLoadSpecJSONReturnsErrorWhenMissing(t *testing.T) {
	t.Chdir(t.TempDir())

	got, err := loadSpecJSON()
	require.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "swagger spec not loaded")
}

func TestRegisterAtServesSwaggerHTML(t *testing.T) {
	tmp := t.TempDir()
	t.Chdir(tmp)
	specDir := filepath.Join(tmp, "gen", "openapiv2", "v1")
	require.NoError(t, os.MkdirAll(specDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(specDir, defaultSpecFileName), []byte(`{"openapi":"3.0.0"}`), 0o644))

	mux := http.NewServeMux()
	RegisterAt(mux, "RAG", "/docs")

	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "text/html")
	assert.Contains(t, rec.Body.String(), "SwaggerUIBundle")
	assert.Contains(t, rec.Body.String(), "RAG")
}
