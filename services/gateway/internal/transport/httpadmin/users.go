package httpadmin

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"secure-rag-platform/services/gateway/internal/usecase"
)

type UserHandlers struct {
	uc     *usecase.Service
	logger *slog.Logger
}

type createUserRequest struct {
	Login      string         `json:"login"`
	Password   string         `json:"password"`
	IsActive   *bool          `json:"is_active"`
	RoleCodes  []string       `json:"role_codes"`
	Attributes map[string]any `json:"attributes"`
}

type createUserResponse struct {
	User *usecase.User `json:"user"`
}

type updateUserRequest struct {
	Login    *string `json:"login"`
	Password *string `json:"password"`
	IsActive *bool   `json:"is_active"`
}

type rolesResponse struct {
	Roles []usecase.Role `json:"roles"`
}

type setUserRolesRequest struct {
	RoleCodes []string `json:"role_codes"`
}

type userRolesResponse struct {
	UserID string         `json:"user_id"`
	Roles  []usecase.Role `json:"roles"`
	CtxVer int64          `json:"ctx_ver"`
}

type replaceUserAttributesRequest struct {
	Attributes map[string]any `json:"attributes"`
}

type userAttributesResponse struct {
	UserID     string         `json:"user_id"`
	Attributes map[string]any `json:"attributes"`
	CtxVer     int64          `json:"ctx_ver"`
}

func NewUserHandlers(uc *usecase.Service, logger *slog.Logger) *UserHandlers {
	if logger == nil {
		logger = slog.Default()
	}
	return &UserHandlers{uc: uc, logger: logger}
}

// CreateUser создаёт пользователя через gateway.
func (h *UserHandlers) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if h.uc == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "service not configured")
		return
	}

	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}

	user, err := h.uc.CreateUser(r.Context(), usecase.CreateUserRequest{
		Login:      req.Login,
		Password:   req.Password,
		IsActive:   req.IsActive,
		RoleCodes:  req.RoleCodes,
		Attributes: req.Attributes,
	}, extractAccessToken(r))
	if err != nil {
		h.logger.WarnContext(r.Context(), "не удалось создать пользователя",
			"component", "gateway.http-admin",
			"login", req.Login,
			"error", err,
		)
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, createUserResponse{User: user})
}

// ListRoles возвращает роли IAM через gateway.
func (h *UserHandlers) ListRoles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if h.uc == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "service not configured")
		return
	}

	roles, err := h.uc.ListRoles(r.Context(), extractAccessToken(r))
	if err != nil {
		h.logger.WarnContext(r.Context(), "не удалось получить роли",
			"component", "gateway.http-admin",
			"error", err,
		)
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, rolesResponse{Roles: roles})
}

// Users обрабатывает административные операции над существующими пользователями.
func (h *UserHandlers) Users(w http.ResponseWriter, r *http.Request) {
	if h.uc == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "service not configured")
		return
	}

	userID, tail, ok := parseAdminUserPath(r.URL.Path)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}

	switch {
	case r.Method == http.MethodPatch && tail == "":
		h.updateUser(w, r, userID)
	case r.Method == http.MethodPut && tail == "roles":
		h.setUserRoles(w, r, userID)
	case r.Method == http.MethodPut && tail == "attributes":
		h.replaceUserAttributes(w, r, userID)
	default:
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *UserHandlers) updateUser(w http.ResponseWriter, r *http.Request, userID string) {
	var req updateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}

	user, err := h.uc.UpdateUser(r.Context(), usecase.UpdateUserRequest{
		UserID:   userID,
		Login:    req.Login,
		Password: req.Password,
		IsActive: req.IsActive,
	}, extractAccessToken(r))
	if err != nil {
		h.logger.WarnContext(r.Context(), "не удалось обновить пользователя",
			"component", "gateway.http-admin",
			"user_id", userID,
			"error", err,
		)
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, createUserResponse{User: user})
}

func (h *UserHandlers) setUserRoles(w http.ResponseWriter, r *http.Request, userID string) {
	var req setUserRolesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}

	result, err := h.uc.SetUserRoles(r.Context(), userID, req.RoleCodes, extractAccessToken(r))
	if err != nil {
		h.logger.WarnContext(r.Context(), "не удалось назначить роли пользователя",
			"component", "gateway.http-admin",
			"user_id", userID,
			"error", err,
		)
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, userRolesResponse{
		UserID: result.UserID,
		Roles:  result.Roles,
		CtxVer: result.CtxVer,
	})
}

func (h *UserHandlers) replaceUserAttributes(w http.ResponseWriter, r *http.Request, userID string) {
	var req replaceUserAttributesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}

	result, err := h.uc.ReplaceUserAttributes(r.Context(), userID, req.Attributes, extractAccessToken(r))
	if err != nil {
		h.logger.WarnContext(r.Context(), "не удалось обновить атрибуты пользователя",
			"component", "gateway.http-admin",
			"user_id", userID,
			"error", err,
		)
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, userAttributesResponse{
		UserID:     result.UserID,
		Attributes: result.Attributes,
		CtxVer:     result.CtxVer,
	})
}

func parseAdminUserPath(path string) (string, string, bool) {
	const prefix = "/gateway/api/v1/admin/users/"
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

func extractAccessToken(r *http.Request) string {
	value := strings.TrimSpace(r.Header.Get("Authorization"))
	if value == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(value), "bearer ") {
		return strings.TrimSpace(value[7:])
	}
	return value
}

func writeUsecaseError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, usecase.ErrNotConfigured):
		writeJSONError(w, http.StatusServiceUnavailable, err.Error())
	case errors.Is(err, usecase.ErrInvalidRequest):
		writeJSONError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, usecase.ErrUnauthorized):
		writeJSONError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, usecase.ErrForbidden):
		writeJSONError(w, http.StatusForbidden, err.Error())
	case errors.Is(err, usecase.ErrNotFound):
		writeJSONError(w, http.StatusNotFound, err.Error())
	default:
		writeJSONError(w, http.StatusInternalServerError, "internal error")
	}
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
