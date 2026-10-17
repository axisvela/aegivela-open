// Package governancesdk is the Apache-2.0 AEGIVELA governance client:
// policy definitions, web resources + permission resolution, and risk signals.
// It wraps the shared internal transport; runtime enforcement stays in pepsdk.
package governancesdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/axisvela/aegivela-open/sdk/go/v3/internal/httpx"
	"github.com/axisvela/aegivela-open/sdk/go/v3/pepsdk"
)

// Client is the governance client.
type Client struct {
	HC *httpx.Client
}

// New creates a governance client.
func New(baseURL string, httpClient *http.Client) *Client {
	return &Client{HC: httpx.New(baseURL, httpClient)}
}

func internalHeader(token string) map[string]string {
	return map[string]string{"X-Aegivela-Internal-Token": token}
}

// postInternal POSTs a JSON body with the internal-token header.
func (c *Client) postInternal(ctx context.Context, path, internalToken string, body []byte) (*http.Response, error) {
	resp, err := c.HC.Request(ctx, http.MethodPost, path, internalHeader(internalToken), body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", pepsdk.ErrUnavailable, err)
	}
	return resp, nil
}

// ---- Policies ----

// Policy is a policy definition (rules are raw JSON).
type Policy struct {
	PolicyID     string          `json:"policy_id"`
	Version      string          `json:"version"`
	TenantID     string          `json:"tenant_id"`
	Rules        json.RawMessage `json:"rules"`
	ActivatedAt  time.Time       `json:"activated_at"`
	ExpiredAt    *time.Time      `json:"expired_at,omitempty"`
	DeprecatedAt *time.Time      `json:"deprecated_at,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
}

// CreatePolicyRequest is the input to CreatePolicy.
type CreatePolicyRequest struct {
	PolicyID     string
	Version      string
	Rules        json.RawMessage
	ActivatedAt  time.Time
	ExpiredAt    *time.Time
	DeprecatedAt *time.Time
}

type createPolicyWire struct {
	PolicyID     string          `json:"policy_id"`
	Version      string          `json:"version"`
	Rules        json.RawMessage `json:"rules"`
	ActivatedAt  time.Time       `json:"activated_at"`
	ExpiredAt    *time.Time      `json:"expired_at,omitempty"`
	DeprecatedAt *time.Time      `json:"deprecated_at,omitempty"`
}

// CreatePolicy creates a policy version.
func (c *Client) CreatePolicy(ctx context.Context, internalToken string, req CreatePolicyRequest) (*Policy, error) {
	if strings.TrimSpace(internalToken) == "" || strings.TrimSpace(req.PolicyID) == "" ||
		strings.TrimSpace(req.Version) == "" || len(req.Rules) == 0 {
		return nil, pepsdk.ErrInvalidInput
	}
	body, err := json.Marshal(createPolicyWire{
		PolicyID: req.PolicyID, Version: req.Version, Rules: req.Rules,
		ActivatedAt: req.ActivatedAt, ExpiredAt: req.ExpiredAt, DeprecatedAt: req.DeprecatedAt,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", pepsdk.ErrUnavailable, err)
	}
	resp, err := c.postInternal(ctx, "/v1/policies", internalToken, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := policyStatus(resp); err != nil {
		return nil, err
	}
	return decodePolicy(resp)
}

// GetPolicy fetches the active policy version.
func (c *Client) GetPolicy(ctx context.Context, internalToken, policyID string) (*Policy, error) {
	if strings.TrimSpace(internalToken) == "" || strings.TrimSpace(policyID) == "" {
		return nil, pepsdk.ErrInvalidInput
	}
	return c.getPolicy(ctx, internalToken, "/v1/policies/"+url.PathEscape(policyID))
}

// GetPolicyVersion fetches a specific policy version.
func (c *Client) GetPolicyVersion(ctx context.Context, internalToken, policyID, version string) (*Policy, error) {
	if strings.TrimSpace(internalToken) == "" || strings.TrimSpace(policyID) == "" || strings.TrimSpace(version) == "" {
		return nil, pepsdk.ErrInvalidInput
	}
	path := "/v1/policies/" + url.PathEscape(policyID) + "/versions/" + url.PathEscape(version)
	return c.getPolicy(ctx, internalToken, path)
}

func (c *Client) getPolicy(ctx context.Context, internalToken, path string) (*Policy, error) {
	resp, err := c.HC.Request(ctx, http.MethodGet, path, internalHeader(internalToken), nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", pepsdk.ErrUnavailable, err)
	}
	defer resp.Body.Close()
	if err := policyStatus(resp); err != nil {
		return nil, err
	}
	return decodePolicy(resp)
}

func policyStatus(resp *http.Response) error {
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return pepsdk.ErrUnauthenticated
	case http.StatusBadRequest:
		return pepsdk.ErrInvalidInput
	case http.StatusNotFound:
		return pepsdk.ErrDenied
	case http.StatusServiceUnavailable:
		return pepsdk.ErrUnavailable
	}
	if resp.StatusCode >= 300 {
		return pepsdk.ErrUnavailable
	}
	return nil
}

func decodePolicy(resp *http.Response) (*Policy, error) {
	var out Policy
	if err := httpx.DecodeBounded(resp.Body, &out, false); err != nil {
		return nil, fmt.Errorf("%w: %v", pepsdk.ErrUnavailable, err)
	}
	return &out, nil
}

// ---- Risk signals ----

// RiskSignalRequest is the input to SubmitRiskSignal (v1alpha1).
type RiskSignalRequest struct {
	SubjectRef string
	SessionRef string
	RiskLevel  string
	RiskRef    string
	Reason     string
	TraceID    string
}

type riskSignalWire struct {
	APIVersion string `json:"api_version"`
	SubjectRef string `json:"subject_ref"`
	SessionRef string `json:"session_ref,omitempty"`
	RiskLevel  string `json:"risk_level"`
	RiskRef    string `json:"risk_ref"`
	Reason     string `json:"reason,omitempty"`
	TraceID    string `json:"trace_id"`
}

// SubmitRiskSignal posts a risk signal that may drive revocation.
func (c *Client) SubmitRiskSignal(ctx context.Context, internalToken string, req RiskSignalRequest) error {
	if strings.TrimSpace(internalToken) == "" || strings.TrimSpace(req.SubjectRef) == "" || strings.TrimSpace(req.RiskRef) == "" {
		return pepsdk.ErrInvalidInput
	}
	traceID := req.TraceID
	if traceID == "" {
		traceID = httpx.GenerateTraceID()
	}
	body, err := json.Marshal(riskSignalWire{
		APIVersion: pepsdk.APIVersionV1Alpha1, SubjectRef: req.SubjectRef, SessionRef: req.SessionRef,
		RiskLevel: req.RiskLevel, RiskRef: req.RiskRef, Reason: req.Reason, TraceID: traceID,
	})
	if err != nil {
		return fmt.Errorf("%w: %v", pepsdk.ErrUnavailable, err)
	}
	resp, err := c.postInternal(ctx, "/v1/risk/signals", internalToken, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return pepsdk.ErrUnauthenticated
	case http.StatusBadRequest:
		return pepsdk.ErrInvalidInput
	case http.StatusServiceUnavailable:
		return pepsdk.ErrUnavailable
	}
	if resp.StatusCode >= 300 {
		return pepsdk.ErrUnavailable
	}
	return nil
}

// ---- Web resources ----

// ResolveWebPermissions resolves the caller's permission set over a product's
// root resource (POST /v1/web-permissions:resolve); returned verbatim.
func (c *Client) ResolveWebPermissions(ctx context.Context, internalToken, productID, rootResourceID string) (json.RawMessage, error) {
	if strings.TrimSpace(internalToken) == "" || strings.TrimSpace(productID) == "" {
		return nil, pepsdk.ErrInvalidInput
	}
	body, err := json.Marshal(struct {
		APIVersion     string `json:"api_version"`
		ProductID      string `json:"product_id"`
		RootResourceID string `json:"root_resource_id,omitempty"`
	}{APIVersion: pepsdk.APIVersionV1Alpha1, ProductID: productID, RootResourceID: rootResourceID})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", pepsdk.ErrUnavailable, err)
	}
	return c.webJSON(ctx, http.MethodPost, "/v1/web-permissions:resolve", internalToken, body)
}

// ListWebResources lists active web resources for a product; returned verbatim.
func (c *Client) ListWebResources(ctx context.Context, internalToken, productID, parentResourceID string) (json.RawMessage, error) {
	if strings.TrimSpace(internalToken) == "" || strings.TrimSpace(productID) == "" {
		return nil, pepsdk.ErrInvalidInput
	}
	path := "/v1/web-resources/" + url.PathEscape(productID)
	if parentResourceID != "" {
		path += "?parent_resource_id=" + url.QueryEscape(parentResourceID)
	}
	return c.webJSON(ctx, http.MethodGet, path, internalToken, nil)
}

// GetWebResource fetches one active web resource; returned verbatim.
func (c *Client) GetWebResource(ctx context.Context, internalToken, productID, resourceID string) (json.RawMessage, error) {
	if strings.TrimSpace(internalToken) == "" || strings.TrimSpace(productID) == "" || strings.TrimSpace(resourceID) == "" {
		return nil, pepsdk.ErrInvalidInput
	}
	path := "/v1/web-resources/" + url.PathEscape(productID) + "/" + url.PathEscape(resourceID)
	return c.webJSON(ctx, http.MethodGet, path, internalToken, nil)
}

// CreateWebResourceRequest is the input to CreateWebResource.
type CreateWebResourceRequest struct {
	ProductID         string
	ResourceID        string
	ParentResourceID  string
	Kind              string
	DisplayName       string
	ProjectionActions json.RawMessage
}

// CreateWebResource registers a web resource; returned verbatim.
func (c *Client) CreateWebResource(ctx context.Context, internalToken string, req CreateWebResourceRequest) (json.RawMessage, error) {
	if strings.TrimSpace(internalToken) == "" || strings.TrimSpace(req.ProductID) == "" ||
		strings.TrimSpace(req.ResourceID) == "" || strings.TrimSpace(req.DisplayName) == "" {
		return nil, pepsdk.ErrInvalidInput
	}
	body, err := json.Marshal(struct {
		APIVersion        string          `json:"api_version"`
		ProductID         string          `json:"product_id"`
		ResourceID        string          `json:"resource_id"`
		ParentResourceID  string          `json:"parent_resource_id"`
		Kind              string          `json:"kind"`
		DisplayName       string          `json:"display_name"`
		ProjectionActions json.RawMessage `json:"projection_actions"`
	}{
		APIVersion: pepsdk.APIVersionV1Alpha1, ProductID: req.ProductID, ResourceID: req.ResourceID,
		ParentResourceID: req.ParentResourceID, Kind: req.Kind, DisplayName: req.DisplayName,
		ProjectionActions: req.ProjectionActions,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", pepsdk.ErrUnavailable, err)
	}
	return c.webJSON(ctx, http.MethodPost, "/v1/web-resources", internalToken, body)
}

// DeprecateWebResource deprecates a web resource; returned verbatim.
func (c *Client) DeprecateWebResource(ctx context.Context, internalToken, productID, resourceID string) (json.RawMessage, error) {
	if strings.TrimSpace(internalToken) == "" || strings.TrimSpace(productID) == "" || strings.TrimSpace(resourceID) == "" {
		return nil, pepsdk.ErrInvalidInput
	}
	path := "/v1/web-resources/" + url.PathEscape(productID) + "/" + url.PathEscape(resourceID) + ":deprecate"
	return c.webJSON(ctx, http.MethodPost, path, internalToken, nil)
}

// CreateWebResourceRuleRequest is the input to CreateWebResourceRule.
type CreateWebResourceRuleRequest struct {
	ProductID       string
	RuleID          string
	ResourceID      string
	Descendants     bool
	SubjectSelector json.RawMessage
	Actions         json.RawMessage
	Effect          string
}

// CreateWebResourceRule creates a web-resource policy rule.
func (c *Client) CreateWebResourceRule(ctx context.Context, internalToken string, req CreateWebResourceRuleRequest) error {
	if strings.TrimSpace(internalToken) == "" || strings.TrimSpace(req.ProductID) == "" ||
		strings.TrimSpace(req.RuleID) == "" || strings.TrimSpace(req.ResourceID) == "" ||
		len(req.SubjectSelector) == 0 || len(req.Actions) == 0 || strings.TrimSpace(req.Effect) == "" {
		return pepsdk.ErrInvalidInput
	}
	body, err := json.Marshal(struct {
		APIVersion      string          `json:"api_version"`
		ProductID       string          `json:"product_id"`
		RuleID          string          `json:"rule_id"`
		ResourceID      string          `json:"resource_id"`
		Descendants     bool            `json:"descendants"`
		SubjectSelector json.RawMessage `json:"subject_selector"`
		Actions         json.RawMessage `json:"actions"`
		Effect          string          `json:"effect"`
	}{
		APIVersion: pepsdk.APIVersionV1Alpha1, ProductID: req.ProductID, RuleID: req.RuleID,
		ResourceID: req.ResourceID, Descendants: req.Descendants, SubjectSelector: req.SubjectSelector,
		Actions: req.Actions, Effect: req.Effect,
	})
	if err != nil {
		return fmt.Errorf("%w: %v", pepsdk.ErrUnavailable, err)
	}
	_, err = c.webJSON(ctx, http.MethodPost, "/v1/web-resource-policies", internalToken, body)
	return err
}

func (c *Client) webJSON(ctx context.Context, method, path, internalToken string, body []byte) (json.RawMessage, error) {
	var reader *bytes.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	var httpReq *http.Request
	var err error
	if reader != nil {
		httpReq, err = http.NewRequestWithContext(ctx, method, c.HC.BaseURL+path, reader)
	} else {
		httpReq, err = http.NewRequestWithContext(ctx, method, c.HC.BaseURL+path, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %v", pepsdk.ErrUnavailable, err)
	}
	if body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	httpReq.Header.Set("X-Aegivela-Internal-Token", internalToken)
	resp, err := c.HC.HTTP.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", pepsdk.ErrUnavailable, err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return nil, pepsdk.ErrUnauthenticated
	case http.StatusForbidden, http.StatusNotFound:
		return nil, pepsdk.ErrDenied
	case http.StatusBadRequest:
		return nil, pepsdk.ErrInvalidInput
	case http.StatusServiceUnavailable:
		return nil, pepsdk.ErrUnavailable
	}
	if resp.StatusCode >= 300 {
		return nil, pepsdk.ErrUnavailable
	}
	var raw json.RawMessage
	if err := httpx.DecodeBounded(resp.Body, &raw, false); err != nil {
		return nil, fmt.Errorf("%w: %v", pepsdk.ErrUnavailable, err)
	}
	return raw, nil
}
