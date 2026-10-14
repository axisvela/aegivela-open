package pepsdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// WorkloadAssertionRequest is the input to IssueWorkloadAssertion.
type WorkloadAssertionRequest struct {
	// APIVersion defaults to APIVersionV1Alpha1.
	APIVersion       string
	SubjectRef       string
	OrganizationRoot string
	BindingID        string
	ClientID         string
	WorkloadRef      string
	// Audience is the target service the assertion is minted for.
	Audience string
	// ExpiresAt defaults to 10 minutes from now when zero.
	ExpiresAt time.Time
	TraceID   string
}

type workloadAssertionRequest struct {
	APIVersion       string    `json:"api_version"`
	SubjectRef       string    `json:"subject_ref"`
	OrganizationRoot string    `json:"organization_root"`
	BindingID        string    `json:"binding_id"`
	ClientID         string    `json:"client_id"`
	WorkloadRef      string    `json:"workload_ref"`
	Audience         string    `json:"audience"`
	ExpiresAt        time.Time `json:"expires_at"`
	TraceID          string    `json:"trace_id"`
}

type workloadAssertionResponse struct {
	WorkloadAssertion string `json:"workload_assertion"`
}

// IssueWorkloadAssertion mints a short-lived workload assertion JWT for
// service-to-service authentication. The caller presents the AEGIVELA internal
// token (X-Aegivela-Internal-Token); the assertion audience is typically the
// target service.
func (c *Client) IssueWorkloadAssertion(ctx context.Context, internalToken string, req WorkloadAssertionRequest) (string, error) {
	if strings.TrimSpace(internalToken) == "" ||
		strings.TrimSpace(req.BindingID) == "" ||
		strings.TrimSpace(req.ClientID) == "" ||
		strings.TrimSpace(req.WorkloadRef) == "" ||
		strings.TrimSpace(req.Audience) == "" {
		return "", ErrInvalidInput
	}
	apiVersion := req.APIVersion
	if apiVersion == "" {
		apiVersion = APIVersionV1Alpha1
	}
	expiresAt := req.ExpiresAt
	if expiresAt.IsZero() {
		expiresAt = time.Now().Add(10 * time.Minute)
	}
	traceID := req.TraceID
	if traceID == "" {
		traceID = generateTraceID()
	}
	body, err := json.Marshal(workloadAssertionRequest{
		APIVersion:       apiVersion,
		SubjectRef:       req.SubjectRef,
		OrganizationRoot: req.OrganizationRoot,
		BindingID:        req.BindingID,
		ClientID:         req.ClientID,
		WorkloadRef:      req.WorkloadRef,
		Audience:         req.Audience,
		ExpiresAt:        expiresAt.UTC(),
		TraceID:          traceID,
	})
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/workload-assertions/issue", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Aegivela-Internal-Token", internalToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return "", ErrUnauthenticated
	case http.StatusBadRequest:
		return "", ErrInvalidInput
	case http.StatusServiceUnavailable:
		return "", ErrUnavailable
	case http.StatusOK:
	default:
		return "", ErrUnavailable
	}

	var wire workloadAssertionResponse
	if err := decodeBoundedResponse(resp.Body, &wire, false); err != nil {
		return "", fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	if strings.TrimSpace(wire.WorkloadAssertion) == "" {
		return "", fmt.Errorf("%w: empty workload assertion", ErrUnavailable)
	}
	return wire.WorkloadAssertion, nil
}
