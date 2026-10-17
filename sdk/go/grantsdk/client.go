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

	"github.com/axisvela/aegivela-open/sdk/go/v3/internal/httpx"
	"github.com/axisvela/aegivela-open/sdk/go/v3/pepsdk"
)

// Client is the execution-grant client.
type Client struct {
	HC *httpx.Client
}

// Option configures a client before it is shared between requests.
type Option func(*Client)

// WithHeaders copies default request headers (for example the configured
// Authorization / X-AEGIVELA-* credentials). Endpoint-specific headers win, so
// the per-call X-Aegivela-Internal-Token is never overridden. Cross-origin
// redirects are refused when defaults are configured.
func WithHeaders(headers map[string]string) Option {
	defaults := make(map[string]string, len(headers))
	for key, value := range headers {
		defaults[key] = value
	}
	return func(c *Client) { c.HC.WithHeaders(defaults) }
}

// WithRetry bounds retries for idempotent requests (transport failures and 5xx
// responses). attempts < 1 disables retries; non-idempotent methods are never
// retried.
func WithRetry(attempts int) Option {
	return func(c *Client) { c.HC.SetRetry(attempts) }
}

// New creates an execution-grant client.
func New(baseURL string, httpClient *http.Client, opts ...Option) *Client {
	c := &Client{HC: httpx.New(baseURL, httpClient)}
	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}
	return c
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
	// Grant is the v1-style envelope field.
	Grant string `json:"grant"`
	// SignedGrant is the v2 structured-response field carrying the same compact
	// signed grant artifact.
	SignedGrant string `json:"signedGrant"`
}

func (r grantIssueResponse) token() string {
	if strings.TrimSpace(r.Grant) != "" {
		return r.Grant
	}
	return r.SignedGrant
}

// IssueGrant mints an execution grant from a signed allow decision.
func (c *Client) IssueGrant(ctx context.Context, internalToken string, req GrantIssueRequest) (string, error) {
	if strings.TrimSpace(internalToken) == "" || strings.TrimSpace(req.BearerToken) == "" || strings.TrimSpace(req.SignedAllowDecision) == "" {
		return "", pepsdk.ErrInvalidInput
	}
	traceID := req.TraceID
	if traceID == "" {
		traceID = httpx.GenerateTraceID()
	}
	body, err := json.Marshal(grantIssueRequestWire{
		BearerToken: req.BearerToken, SignedDecision: req.SignedAllowDecision, Scope: req.Scope,
		TraceID: traceID, ApprovalJTI: req.ApprovalJTI, PreAuthorizationJTI: req.PreAuthorizationJTI,
	})
	if err != nil {
		return "", fmt.Errorf("%w: %v", pepsdk.ErrUnavailable, err)
	}
	resp, err := c.HC.Request(ctx, http.MethodPost, "/v1/grants/issue", internalHeader(internalToken), body)
	if err != nil {
		return "", fmt.Errorf("%w: %v", pepsdk.ErrUnavailable, err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return "", &pepsdk.APIError{StatusCode: resp.StatusCode, Err: pepsdk.ErrUnauthenticated}
	case http.StatusForbidden:
		return "", &pepsdk.APIError{StatusCode: resp.StatusCode, Err: pepsdk.ErrDenied}
	case http.StatusBadRequest:
		return "", &pepsdk.APIError{StatusCode: resp.StatusCode, Err: pepsdk.ErrInvalidInput}
	case http.StatusServiceUnavailable:
		return "", &pepsdk.APIError{StatusCode: resp.StatusCode, Err: pepsdk.ErrUnavailable}
	case http.StatusOK, http.StatusCreated:
	default:
		return "", &pepsdk.APIError{StatusCode: resp.StatusCode, Err: pepsdk.ErrUnavailable}
	}
	var wire grantIssueResponse
	if err := httpx.DecodeBounded(resp.Body, &wire, false); err != nil {
		return "", fmt.Errorf("%w: %v", pepsdk.ErrUnavailable, err)
	}
	if strings.TrimSpace(wire.token()) == "" {
		return "", fmt.Errorf("%w: empty grant", pepsdk.ErrUnavailable)
	}
	return wire.token(), nil
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
	resp, err := c.HC.Request(ctx, http.MethodPost, "/v1/grants/verify", internalHeader(internalToken), body)
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
	if err := httpx.DecodeBounded(resp.Body, &wire, false); err != nil {
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
	resp, err := c.HC.Request(ctx, http.MethodGet, "/v1/grants/jwks", nil, nil)
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
