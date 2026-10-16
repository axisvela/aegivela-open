// Package identitysdk is the Apache-2.0 AEGIVELA workload-identity client:
// workload assertions and bindings, plus the assertion JWKS. It wraps the
// shared pepsdk client for transport.
package identitysdk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/axisrobo/aegivela-open/sdk/go/internal/httpx"
	"github.com/axisrobo/aegivela-open/sdk/go/pepsdk"
)

// Client is the workload-identity client.
type Client struct {
	HC *httpx.Client
	// Authority is the caller-asserted trusted authority sent alongside the
	// internal token. AEGIVELA's internal resolver requires the tenant, actor and
	// authority-root headers (its binding/assertion handlers compare the request's
	// subject/organization against the resolved actor/authority), so a client that
	// sends only the internal token is answered 401. Zero values are omitted, so a
	// client that does not set them behaves exactly as before.
	Authority TrustedAuthority
}

// TrustedAuthority is the caller-asserted authority for internal-token calls.
type TrustedAuthority struct {
	TenantID      string
	ActorID       string
	AuthorityRoot string
}

// New creates a workload-identity client.
func New(baseURL string, httpClient *http.Client) *Client {
	return &Client{HC: httpx.New(baseURL, httpClient)}
}

// WithTrustedAuthority sets the authority headers sent with internal-token
// calls and returns the client for chaining.
func (c *Client) WithTrustedAuthority(a TrustedAuthority) *Client {
	c.Authority = a
	return c
}

func internalHeader(token string) map[string]string {
	return map[string]string{"X-Aegivela-Internal-Token": token}
}

// headers builds the request headers: the internal token plus any configured
// trusted-authority headers.
func (c *Client) headers(token string) map[string]string {
	h := internalHeader(token)
	if v := strings.TrimSpace(c.Authority.TenantID); v != "" {
		h["X-Aegivela-Tenant-ID"] = v
	}
	if v := strings.TrimSpace(c.Authority.ActorID); v != "" {
		h["X-Aegivela-Actor-ID"] = v
	}
	if v := strings.TrimSpace(c.Authority.AuthorityRoot); v != "" {
		h["X-Aegivela-Authority-Root"] = v
	}
	return h
}

// WorkloadAssertionRequest is the input to IssueWorkloadAssertion.
type WorkloadAssertionRequest struct {
	APIVersion       string
	SubjectRef       string
	OrganizationRoot string
	BindingID        string
	ClientID         string
	WorkloadRef      string
	Audience         string
	ExpiresAt        time.Time
	TraceID          string
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

// IssueWorkloadAssertion mints a short-lived workload assertion JWT.
func (c *Client) IssueWorkloadAssertion(ctx context.Context, internalToken string, req WorkloadAssertionRequest) (string, error) {
	if strings.TrimSpace(internalToken) == "" ||
		strings.TrimSpace(req.BindingID) == "" || strings.TrimSpace(req.ClientID) == "" ||
		strings.TrimSpace(req.WorkloadRef) == "" || strings.TrimSpace(req.Audience) == "" {
		return "", pepsdk.ErrInvalidInput
	}
	apiVersion := req.APIVersion
	if apiVersion == "" {
		apiVersion = pepsdk.APIVersionV1Alpha1
	}
	expiresAt := req.ExpiresAt
	if expiresAt.IsZero() {
		// The server bounds an assertion's lifetime (AEGIVELA maxAssertionTTL =
		// 5m) and rejects a request beyond it, so the default must sit inside the
		// contract rather than at a value the server always refuses.
		expiresAt = time.Now().Add(5 * time.Minute)
	}
	traceID := req.TraceID
	if traceID == "" {
		traceID = httpx.GenerateTraceID()
	}
	body, err := json.Marshal(workloadAssertionRequest{
		APIVersion: apiVersion, SubjectRef: req.SubjectRef, OrganizationRoot: req.OrganizationRoot,
		BindingID: req.BindingID, ClientID: req.ClientID, WorkloadRef: req.WorkloadRef,
		Audience: req.Audience, ExpiresAt: expiresAt.UTC(), TraceID: traceID,
	})
	if err != nil {
		return "", fmt.Errorf("%w: %v", pepsdk.ErrUnavailable, err)
	}
	resp, err := c.HC.Request(ctx, http.MethodPost, "/v1/workload-assertions/issue", c.headers(internalToken), body)
	if err != nil {
		return "", fmt.Errorf("%w: %v", pepsdk.ErrUnavailable, err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return "", pepsdk.ErrUnauthenticated
	case http.StatusBadRequest:
		return "", pepsdk.ErrInvalidInput
	case http.StatusServiceUnavailable:
		return "", pepsdk.ErrUnavailable
	case http.StatusOK:
	default:
		return "", pepsdk.ErrUnavailable
	}
	var wire workloadAssertionResponse
	if err := httpx.DecodeBounded(resp.Body, &wire, false); err != nil {
		return "", fmt.Errorf("%w: %v", pepsdk.ErrUnavailable, err)
	}
	if strings.TrimSpace(wire.WorkloadAssertion) == "" {
		return "", fmt.Errorf("%w: empty workload assertion", pepsdk.ErrUnavailable)
	}
	return wire.WorkloadAssertion, nil
}

// CreateWorkloadBindingRequest is the input to CreateWorkloadBinding.
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

// CreateWorkloadBinding registers a workload binding; the binding JSON is
// returned verbatim.
func (c *Client) CreateWorkloadBinding(ctx context.Context, internalToken string, req CreateWorkloadBindingRequest) (json.RawMessage, error) {
	if strings.TrimSpace(internalToken) == "" ||
		strings.TrimSpace(req.SubjectRef) == "" || strings.TrimSpace(req.OrganizationRoot) == "" ||
		strings.TrimSpace(req.ClientID) == "" || strings.TrimSpace(req.WorkloadRef) == "" {
		return nil, pepsdk.ErrInvalidInput
	}
	apiVersion := req.APIVersion
	if apiVersion == "" {
		apiVersion = pepsdk.APIVersionV1Alpha1
	}
	traceID := req.TraceID
	if traceID == "" {
		traceID = httpx.GenerateTraceID()
	}
	body, err := json.Marshal(createWorkloadBindingWire{
		APIVersion: apiVersion, SubjectRef: req.SubjectRef, OrganizationRoot: req.OrganizationRoot,
		ClientID: req.ClientID, WorkloadRef: req.WorkloadRef, EvidenceRef: req.EvidenceRef, TraceID: traceID,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", pepsdk.ErrUnavailable, err)
	}
	resp, err := c.HC.Request(ctx, http.MethodPost, "/v1/workload-bindings", c.headers(internalToken), body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", pepsdk.ErrUnavailable, err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return nil, pepsdk.ErrUnauthenticated
	case http.StatusBadRequest:
		return nil, pepsdk.ErrInvalidInput
	case http.StatusServiceUnavailable:
		return nil, pepsdk.ErrUnavailable
	case http.StatusOK, http.StatusCreated:
	default:
		return nil, pepsdk.ErrUnavailable
	}
	var raw json.RawMessage
	if err := httpx.DecodeBounded(resp.Body, &raw, false); err != nil {
		return nil, fmt.Errorf("%w: %v", pepsdk.ErrUnavailable, err)
	}
	return raw, nil
}

// WorkloadAssertionJWKS returns the raw JWKS used to verify workload assertions.
func (c *Client) WorkloadAssertionJWKS(ctx context.Context) (json.RawMessage, error) {
	resp, err := c.HC.Request(ctx, http.MethodGet, "/v1/workload-assertions/jwks", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", pepsdk.ErrUnavailable, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, pepsdk.ErrUnavailable
	}
	var raw json.RawMessage
	if err := httpx.DecodeBounded(resp.Body, &raw, false); err != nil {
		return nil, fmt.Errorf("%w: %v", pepsdk.ErrUnavailable, err)
	}
	return raw, nil
}
