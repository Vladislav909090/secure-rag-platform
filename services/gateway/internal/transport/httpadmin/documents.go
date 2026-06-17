package httpadmin

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"secure-rag-platform/services/gateway/internal/usecase"
)

type DocumentHandlers struct {
	uc     *usecase.Service
	logger *slog.Logger
}

type updateDocumentRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
}

type updateDocumentAttributesRequest struct {
	Attributes map[string]any `json:"attributes"`
}

type documentResponse struct {
	Document usecase.Document `json:"document"`
}

func NewDocumentHandlers(uc *usecase.Service, logger *slog.Logger) *DocumentHandlers {
	if logger == nil {
		logger = slog.Default()
	}

	return &DocumentHandlers{uc: uc, logger: logger}
}

// Documents перехватывает административные PATCH-операции над документами
func (h *DocumentHandlers) Documents(gateway http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		documentUUID, tail, ok := parseDocumentPath(r.URL.Path)
		if !ok || r.Method != http.MethodPatch {
			gateway.ServeHTTP(w, r)
			return
		}
		if h.uc == nil {
			writeJSONError(w, http.StatusServiceUnavailable, "service not configured")
			return
		}

		if tail == "" {
			h.updateDocument(w, r, documentUUID)
			return
		}
		if tail == "attributes" {
			h.updateDocumentAttributes(w, r, documentUUID)
			return
		}

		gateway.ServeHTTP(w, r)
	}
}

func (h *DocumentHandlers) updateDocument(w http.ResponseWriter, r *http.Request, documentUUID string) {
	var req updateDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}

	item, err := h.uc.UpdateDocument(r.Context(), usecase.UpdateDocumentRequest{
		DocumentUUID: documentUUID,
		Title:        req.Title,
		Description:  req.Description,
	}, extractAccessToken(r))
	if err != nil {
		h.logger.WarnContext(r.Context(), "не удалось обновить документ",
			"component", "gateway.http-admin",
			"document_uuid", documentUUID,
			"error", err,
		)
		writeUsecaseError(w, err)

		return
	}

	writeJSON(w, http.StatusOK, documentResponse{Document: item.Document})
}

func (h *DocumentHandlers) updateDocumentAttributes(w http.ResponseWriter, r *http.Request, documentUUID string) {
	var req updateDocumentAttributesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}

	item, err := h.uc.UpdateDocumentAttributes(r.Context(), usecase.UpdateDocumentAttributesRequest{
		DocumentUUID: documentUUID,
		Attributes:   req.Attributes,
	}, extractAccessToken(r))
	if err != nil {
		h.logger.WarnContext(r.Context(), "не удалось обновить атрибуты документа",
			"component", "gateway.http-admin",
			"document_uuid", documentUUID,
			"error", err,
		)
		writeUsecaseError(w, err)

		return
	}

	writeJSON(w, http.StatusOK, documentResponse{Document: item.Document})
}

func parseDocumentPath(path string) (string, string, bool) {
	const prefix = "/gateway/api/v1/documents/"
	if !strings.HasPrefix(path, prefix) {
		return "", "", false
	}
	trimmed := strings.Trim(strings.TrimPrefix(path, prefix), "/")
	if trimmed == "" {
		return "", "", false
	}

	parts := strings.SplitN(trimmed, "/", 2)
	if parts[0] == "" {
		return "", "", false
	}
	if len(parts) == 1 {
		return parts[0], "", true
	}

	return parts[0], strings.Trim(parts[1], "/"), true
}
