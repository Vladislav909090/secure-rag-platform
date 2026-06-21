package httpadmin

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"secure-rag-platform/services/gateway/internal/usecase"
)

func TestParseAdminUserPath(t *testing.T) {
	userID, tail, ok := parseAdminUserPath("/gateway/api/v1/admin/users/user-1/roles")
	if !ok || userID != "user-1" || tail != "roles" {
		t.Fatalf("parseAdminUserPath() = %q, %q, %v", userID, tail, ok)
	}

	_, _, ok = parseAdminUserPath("/gateway/api/v1/admin/users/")
	if ok {
		t.Fatalf("empty admin user path should not match")
	}
}

func TestParseDocumentPath(t *testing.T) {
	docID, tail, ok := parseDocumentPath("/gateway/api/v1/documents/doc-1/attributes")
	if !ok || docID != "doc-1" || tail != "attributes" {
		t.Fatalf("parseDocumentPath() = %q, %q, %v", docID, tail, ok)
	}

	_, _, ok = parseDocumentPath("/gateway/api/v1/other/doc-1")
	if ok {
		t.Fatalf("unexpected document path match")
	}
}

func TestWriteUsecaseError(t *testing.T) {
	tests := []struct {
		err    error
		status int
	}{
		{usecase.ErrNotConfigured, http.StatusServiceUnavailable},
		{usecase.ErrInvalidRequest, http.StatusBadRequest},
		{usecase.ErrUnauthorized, http.StatusUnauthorized},
		{usecase.ErrForbidden, http.StatusForbidden},
		{usecase.ErrNotFound, http.StatusNotFound},
		{errors.New("boom"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		rec := httptest.NewRecorder()
		writeUsecaseError(rec, tt.err)
		if rec.Code != tt.status {
			t.Fatalf("writeUsecaseError(%v) status = %d, want %d", tt.err, rec.Code, tt.status)
		}
	}
}
