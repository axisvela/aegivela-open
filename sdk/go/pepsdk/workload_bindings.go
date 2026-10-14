package pepsdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// CreateWorkloadBindingRequest is the input to CreateWorkloadBinding
// (v1alpha1 contract).
type CreateWorkloadBindingRequest struct {
	APIVersion       string
	SubjectRef       string
	OrganizationRoot string
	ClientID         string
	WorkloadRef      string
	EvidenceRef      string
	TraceID          string
}

type createWorkloadBindingWire struct {
	APIVersion       string `json:"api_version"`
	SubjectRef       string `json:"subject_ref"`
	OrganizationRoot string `json:"organization_root"`
	ClientID         string `json:"client_id"`
	WorkloadRef      string `json:"workload_ref"`
	EvidenceRef      string `json:"evidence_ref"`
	TraceID          string `json:"trace_id"`
}

// CreateWorkloadBinding registers a workload binding with AEGIVELA. The caller
// presents the AEGIVELA internal token; the created binding JSON is returned
// verbatim.
func (c *Client) CreateWorkloadBinding(ctx context.Context, internalToken string, req CreateWorkloadBindingRequest) (json.RawMessage, error) {
	if strings.TrimSpace(internalToken) == "" ||
		strings.TrimSpace(req.SubjectRef) == "" ||
		strings.TrimSpace(req.OrganizationRoot) == "" ||
		strings.TrimSpace(req.ClientID) == "" ||
		strings.TrimSpace(req.WorkloadRef) == "" {
		return nil, ErrInvalidInput
	}
	apiVersion := req.APIVersion
	if apiVersion == "" {
		apiVersion = APIVersionV1Alpha1
	}
	traceID := req.TraceID
	if traceID == "" {
		traceID = generateTraceID()
	}
	body, err := json.Marshal(createWorkloadBindingWire{
		APIVersion: apiVersion, SubjectRef: req.SubjectRef, OrganizationRoot: req.OrganizationRoot,
		ClientID: req.ClientID, WorkloadRef: req.WorkloadRef, EvidenceRef: req.EvidenceRef, TraceID: traceID,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/workload-bindings", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Aegivela-Internal-Token", internalToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return nil, ErrUnauthenticated
	case http.StatusBadRequest:
		return nil, ErrInvalidInput
	case http.StatusServiceUnavailable:
		return nil, ErrUnavailable
	case http.StatusOK, http.StatusCreated:
	default:
		return nil, ErrUnavailable
	}
	var raw json.RawMessage
	if err := decodeBoundedResponse(resp.Body, &raw, false); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return raw, nil
}

// WorkloadAssertionJWKS returns the raw JWKS used to verify workload
// assertions.
func (c *Client) WorkloadAssertionJWKS(ctx context.Context) (json.RawMessage, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/workload-assertions/jwks", nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, ErrUnavailable
	}
	var raw json.RawMessage
	if err := decodeBoundedResponse(resp.Body, &raw, false); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return raw, nil
}
