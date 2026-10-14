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

// GrantIssueRequest is the input to IssueGrant (v1alpha1 contract). A grant is
// minted from a signed allow decision plus the caller's bearer token.
type GrantIssueRequest struct {
	BearerToken         string
	SignedAllowDecision string
	Scope               []string
	TraceID             string
	ApprovalJTI         string
	PreAuthorizationJTI string
}

type grantIssueRequestWire struct {
	BearerToken         string   `json:"bearer_token"`
	SignedDecision      string   `json:"signed_allow_decision"`
	Scope               []string `json:"scope,omitempty"`
	TraceID             string   `json:"trace_id"`
	ApprovalJTI         string   `json:"approval_jti,omitempty"`
	PreAuthorizationJTI string   `json:"pre_authorization_jti,omitempty"`
}

type grantIssueResponse struct {
	Grant string `json:"grant"`
}

// IssueGrant mints an execution grant from a signed allow decision. The caller
// presents the AEGIVELA internal token.
func (c *Client) IssueGrant(ctx context.Context, internalToken string, req GrantIssueRequest) (string, error) {
	if strings.TrimSpace(internalToken) == "" ||
		strings.TrimSpace(req.BearerToken) == "" ||
		strings.TrimSpace(req.SignedAllowDecision) == "" {
		return "", ErrInvalidInput
	}
	traceID := req.TraceID
	if traceID == "" {
		traceID = generateTraceID()
	}
	body, err := json.Marshal(grantIssueRequestWire{
		BearerToken:         req.BearerToken,
		SignedDecision:      req.SignedAllowDecision,
		Scope:               req.Scope,
		TraceID:             traceID,
		ApprovalJTI:         req.ApprovalJTI,
		PreAuthorizationJTI: req.PreAuthorizationJTI,
	})
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/grants/issue", bytes.NewReader(body))
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
	case http.StatusForbidden:
		return "", ErrDenied
	case http.StatusBadRequest:
		return "", ErrInvalidInput
	case http.StatusServiceUnavailable:
		return "", ErrUnavailable
	case http.StatusOK, http.StatusCreated:
	default:
		return "", ErrUnavailable
	}
	var wire grantIssueResponse
	if err := decodeBoundedResponse(resp.Body, &wire, false); err != nil {
		return "", fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	if strings.TrimSpace(wire.Grant) == "" {
		return "", fmt.Errorf("%w: empty grant", ErrUnavailable)
	}
	return wire.Grant, nil
}

// GrantVerifyRequest is the input to VerifyGrant (v2 contract).
type GrantVerifyRequest struct {
	SignedGrant string
	Audience    string
	Scope       []string
	TaskBinding string
	TraceID     string
}

// VerifyGrantResult is the verified grant state.
type VerifyGrantResult struct {
	APIVersion string
	GrantID    string
	Valid      bool
	Scope      []string
	Audience   string
	ExpiresAt  time.Time
	Expired    bool
	Revoked    bool
}

type grantVerifyBindingV2 struct {
	Audience    string   `json:"audience"`
	Scope       []string `json:"scope"`
	TaskBinding string   `json:"taskBinding,omitempty"`
	TraceID     string   `json:"traceId,omitempty"`
}

type grantVerifyRequestV2 struct {
	APIVersion       string               `json:"apiVersion"`
	SignedGrant      string               `json:"signedGrant"`
	ExpectedBindings grantVerifyBindingV2 `json:"expectedBindings"`
}

type grantVerifyResponseV2 struct {
	APIVersion string    `json:"apiVersion"`
	GrantID    string    `json:"grantId"`
	Valid      bool      `json:"valid"`
	Scope      []string  `json:"scope"`
	Audience   string    `json:"audience"`
	ExpiresAt  time.Time `json:"expiresAt"`
	Expired    bool      `json:"expired"`
	Revoked    bool      `json:"revoked"`
}

// VerifyGrant verifies a signed execution grant against expected bindings
// (v2 contract). The caller presents the AEGIVELA internal or PEP token.
func (c *Client) VerifyGrant(ctx context.Context, internalToken string, req GrantVerifyRequest) (*VerifyGrantResult, error) {
	if strings.TrimSpace(internalToken) == "" ||
		strings.TrimSpace(req.SignedGrant) == "" ||
		strings.TrimSpace(req.Audience) == "" || len(req.Scope) == 0 {
		return nil, ErrInvalidInput
	}
	body, err := json.Marshal(grantVerifyRequestV2{
		APIVersion:  APIVersionV2,
		SignedGrant: req.SignedGrant,
		ExpectedBindings: grantVerifyBindingV2{
			Audience: req.Audience, Scope: req.Scope, TaskBinding: req.TaskBinding, TraceID: req.TraceID,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/grants/verify", bytes.NewReader(body))
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
	case http.StatusForbidden:
		return nil, ErrDenied
	case http.StatusBadRequest:
		return nil, ErrInvalidInput
	case http.StatusServiceUnavailable:
		return nil, ErrUnavailable
	case http.StatusOK:
	default:
		return nil, ErrUnavailable
	}
	var wire grantVerifyResponseV2
	if err := decodeBoundedResponse(resp.Body, &wire, false); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return &VerifyGrantResult{
		APIVersion: wire.APIVersion, GrantID: wire.GrantID, Valid: wire.Valid,
		Scope: wire.Scope, Audience: wire.Audience, ExpiresAt: wire.ExpiresAt,
		Expired: wire.Expired, Revoked: wire.Revoked,
	}, nil
}

// GrantJWKS returns the raw JSON Web Key Set for grant verification.
func (c *Client) GrantJWKS(ctx context.Context) (json.RawMessage, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/grants/jwks", nil)
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
