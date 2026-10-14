package pepsdk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Policy is a policy definition (rules are returned verbatim as raw JSON).
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

// CreatePolicy creates a policy version. The caller presents the AEGIVELA
// internal token.
func (c *Client) CreatePolicy(ctx context.Context, internalToken string, req CreatePolicyRequest) (*Policy, error) {
	if strings.TrimSpace(internalToken) == "" ||
		strings.TrimSpace(req.PolicyID) == "" || strings.TrimSpace(req.Version) == "" || len(req.Rules) == 0 {
		return nil, ErrInvalidInput
	}
	body, err := json.Marshal(createPolicyWire{
		PolicyID: req.PolicyID, Version: req.Version, Rules: req.Rules,
		ActivatedAt: req.ActivatedAt, ExpiredAt: req.ExpiredAt, DeprecatedAt: req.DeprecatedAt,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
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

// GetPolicy fetches the active policy version for policyID.
func (c *Client) GetPolicy(ctx context.Context, internalToken, policyID string) (*Policy, error) {
	if strings.TrimSpace(internalToken) == "" || strings.TrimSpace(policyID) == "" {
		return nil, ErrInvalidInput
	}
	return c.getPolicy(ctx, internalToken, "/v1/policies/"+url.PathEscape(policyID))
}

// GetPolicyVersion fetches a specific policy version.
func (c *Client) GetPolicyVersion(ctx context.Context, internalToken, policyID, version string) (*Policy, error) {
	if strings.TrimSpace(internalToken) == "" || strings.TrimSpace(policyID) == "" || strings.TrimSpace(version) == "" {
		return nil, ErrInvalidInput
	}
	path := "/v1/policies/" + url.PathEscape(policyID) + "/versions/" + url.PathEscape(version)
	return c.getPolicy(ctx, internalToken, path)
}

func (c *Client) getPolicy(ctx context.Context, internalToken, path string) (*Policy, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	httpReq.Header.Set("X-Aegivela-Internal-Token", internalToken)
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
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
		return ErrUnauthenticated
	case http.StatusBadRequest:
		return ErrInvalidInput
	case http.StatusNotFound:
		return ErrDenied
	case http.StatusServiceUnavailable:
		return ErrUnavailable
	}
	if resp.StatusCode >= 300 {
		return ErrUnavailable
	}
	return nil
}

func decodePolicy(resp *http.Response) (*Policy, error) {
	var out Policy
	if err := decodeBoundedResponse(resp.Body, &out, false); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return &out, nil
}
