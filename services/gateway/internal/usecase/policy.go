package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	iamv1 "secure-rag-platform/services/iam/gen/v1"
)

type PolicyAuthorizer interface {
	AllowDocument(ctx context.Context, subject *iamv1.SubjectContext, documentAttributes map[string]any) (bool, error)
}

type OPAAuthorizer struct {
	endpoint   string
	httpClient *http.Client
}

func NewOPAAuthorizer(baseURL string) *OPAAuthorizer {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return nil
	}
	return &OPAAuthorizer{
		endpoint:   baseURL + "/v1/data/secure_rag/document/allow",
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (a *OPAAuthorizer) AllowDocument(
	ctx context.Context,
	subject *iamv1.SubjectContext,
	documentAttributes map[string]any,
) (bool, error) {
	if a == nil {
		return allowedByAttributes(documentAttributes, subject), nil
	}

	subjectInput := map[string]any{
		"user_id":    "",
		"login":      "",
		"is_active":  false,
		"roles":      []string{},
		"attributes": map[string]any{},
	}
	if subject != nil {
		subjectInput["user_id"] = subject.GetUserId()
		subjectInput["login"] = subject.GetLogin()
		subjectInput["is_active"] = subject.GetIsActive()
		subjectInput["roles"] = subject.GetRoles()
		if subject.GetAttributes() != nil {
			subjectInput["attributes"] = subject.GetAttributes().AsMap()
		}
	}

	payload := map[string]any{
		"input": map[string]any{
			"subject": subjectInput,
			"document": map[string]any{
				"attributes": documentAttributes,
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Errorf("marshal opa input: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.endpoint, bytes.NewReader(body))
	if err != nil {
		return false, fmt.Errorf("create opa request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("call opa: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return false, fmt.Errorf("opa status: %d", resp.StatusCode)
	}

	var out struct {
		Result bool `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return false, fmt.Errorf("decode opa response: %w", err)
	}
	return out.Result, nil
}
