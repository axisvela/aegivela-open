// Package grantsdk is the Apache-2.0 AEGIVELA execution-grant client: issue,
// verify, and the grant JWKS. It wraps the shared pepsdk client for transport.
package grantsdk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/axisrobo/aegivela-open/sdk/go/pepsdk"
)

// Client is the execution-grant client.
type Client struct {
	PE *pepsdk.Client
}

// New creates an execution-grant client.
func New(baseURL string, httpClient *http.Client) *Client {
	return &Client{PE: pepsdk.NewClient(baseURL, httpClient)}
}

func internalHeader(token string) map[string]string {
	return map[string]string{"X-Aegivela-Internal-Token": token}
}

// GrantIssueRequest is the input to IssueGrant (v1alpha1).
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

// IssueGrant mints an execution grant from a signed allow decision.
func (c *Client) IssueGrant(ctx context.Context, internalToken string, req GrantIssueRequest) (string, error) {
	if strings.TrimSpace(internalToken) == "" || strings.TrimSpace(req.BearerToken) == "" || strings.TrimSpace(req.SignedAllowDecision) == "" {
		return "", pepsdk.ErrInvalidInput
	}
	traceID := req.TraceID
	if traceID == "" {
		traceID = pepsdk.GenerateTraceID()
	}
	body, err := json.Marshal(grantIssueRequestWire{
		BearerToken: req.BearerToken, SignedDecision: req.SignedAllowDecision, Scope: req.Scope,
		TraceID: traceID, ApprovalJTI: req.ApprovalJTI, PreAuthorizationJTI: req.PreAuthorizationJTI,
	})
	if err != nil {
		return "", fmt.Errorf("%w: %v", pepsdk.ErrUnavailable, err)
	}
	resp, err := c.PE.Request(ctx, http.MethodPost, "/v1/grants/issue", internalHeader(internalToken), body)
	if err != nil {
		return "", fmt.Errorf("%w: %v", pepsdk.ErrUnavailable, err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return "", pepsdk.ErrUnauthenticated
	case http.StatusForbidden:
		return "", pepsdk.ErrDenied
	case http.StatusBadRequest:
		return "", pepsdk.ErrInvalidInput
	case http.StatusServiceUnavailable:
		return "", pepsdk.ErrUnavailable
	case http.StatusOK, http.StatusCreated:
	default:
		return "", pepsdk.ErrUnavailable
	}
	var wire grantIssueResponse
	if err := pepsdk.DecodeJSON(resp.Body, &wire, false); err != nil {
		return "", fmt.Errorf("%w: %v", pepsdk.ErrUnavailable, err)
	}
	if strings.TrimSpace(wire.Grant) == "" {
		return "", fmt.Errorf("%w: empty grant", pepsdk.ErrUnavailable)
	}
	return wire.Grant, nil
}

// GrantVerifyRequest is the input to VerifyGrant (v2).
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

// VerifyGrant verifies a signed execution grant against expected bindings (v2).
func (c *Client) VerifyGrant(ctx context.Context, internalToken string, req GrantVerifyRequest) (*VerifyGrantResult, error) {
	if strings.TrimSpace(internalToken) == "" || strings.TrimSpace(req.SignedGrant) == "" ||
		strings.TrimSpace(req.Audience) == "" || len(req.Scope) == 0 {
		return nil, pepsdk.ErrInvalidInput
	}
	body, err := json.Marshal(grantVerifyRequestV2{
		APIVersion: pepsdk.APIVersionV2, SignedGrant: req.SignedGrant,
		ExpectedBindings: grantVerifyBindingV2{Audience: req.Audience, Scope: req.Scope, TaskBinding: req.TaskBinding, TraceID: req.TraceID},
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", pepsdk.ErrUnavailable, err)
	}
	resp, err := c.PE.Request(ctx, http.MethodPost, "/v1/grants/verify", internalHeader(internalToken), body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", pepsdk.ErrUnavailable, err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return nil, pepsdk.ErrUnauthenticated
	case http.StatusForbidden:
		return nil, pepsdk.ErrDenied
	case http.StatusBadRequest:
		return nil, pepsdk.ErrInvalidInput
	case http.StatusServiceUnavailable:
		return nil, pepsdk.ErrUnavailable
	case http.StatusOK:
	default:
		return nil, pepsdk.ErrUnavailable
	}
	var wire grantVerifyResponseV2
	if err := pepsdk.DecodeJSON(resp.Body, &wire, false); err != nil {
		return nil, fmt.Errorf("%w: %v", pepsdk.ErrUnavailable, err)
	}
	return &VerifyGrantResult{
		APIVersion: wire.APIVersion, GrantID: wire.GrantID, Valid: wire.Valid,
		Scope: wire.Scope, Audience: wire.Audience, ExpiresAt: wire.ExpiresAt,
		Expired: wire.Expired, Revoked: wire.Revoked,
	}, nil
}

// GrantJWKS returns the raw JWKS for grant verification.
func (c *Client) GrantJWKS(ctx context.Context) (json.RawMessage, error) {
	resp, err := c.PE.Request(ctx, http.MethodGet, "/v1/grants/jwks", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", pepsdk.ErrUnavailable, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, pepsdk.ErrUnavailable
	}
	var raw json.RawMessage
	if err := pepsdk.DecodeJSON(resp.Body, &raw, false); err != nil {
		return nil, fmt.Errorf("%w: %v", pepsdk.ErrUnavailable, err)
	}
	return raw, nil
}
