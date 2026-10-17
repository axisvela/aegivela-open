package grantsdk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/axisvela/aegivela-open/sdk/go/internal/httpx"
	"github.com/axisvela/aegivela-open/sdk/go/pepsdk"
)

// ResourceRef identifies a resource by kind + id.
type ResourceRef struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
}

// ApprovalRequest is the input to IssueApproval (v1alpha1).
type ApprovalRequest struct {
	BearerToken   string
	Action        string
	Resource      ResourceRef
	Scope         []string
	Expiry        string
	Reason        string
	PolicyVersion string
}

type approvalIssueWire struct {
	APIVersion    string      `json:"api_version"`
	BearerToken   string      `json:"bearer_token"`
	Action        string      `json:"action"`
	Resource      ResourceRef `json:"resource"`
	Scope         []string    `json:"scope"`
	Expiry        string      `json:"expiry"`
	Reason        string      `json:"reason"`
	PolicyVersion string      `json:"policy_version"`
}

// IssueApproval issues an approval artifact; the artifact JSON is returned
// verbatim. The caller presents the AEGIVELA internal token.
func (c *Client) IssueApproval(ctx context.Context, internalToken string, req ApprovalRequest) (json.RawMessage, error) {
	if strings.TrimSpace(internalToken) == "" || strings.TrimSpace(req.BearerToken) == "" ||
		strings.TrimSpace(req.Action) == "" || strings.TrimSpace(req.Resource.Kind) == "" {
		return nil, pepsdk.ErrInvalidInput
	}
	body, err := json.Marshal(approvalIssueWire{
		APIVersion: pepsdk.APIVersionV1Alpha1, BearerToken: req.BearerToken, Action: req.Action,
		Resource: req.Resource, Scope: req.Scope, Expiry: req.Expiry, Reason: req.Reason,
		PolicyVersion: req.PolicyVersion,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", pepsdk.ErrUnavailable, err)
	}
	return c.authorityJSON(ctx, "/v1/approvals/issue", internalToken, body)
}

// PreAuthorizationRequest is the input to IssuePreAuthorization.
type PreAuthorizationRequest struct {
	BearerToken   string
	Action        string
	Resource      ResourceRef
	Scope         []string
	Expiry        string
	PolicyVersion string
	Audience      string
	TraceID       string
}

type preAuthIssueWire struct {
	BearerToken   string      `json:"bearer_token"`
	Action        string      `json:"action"`
	Resource      ResourceRef `json:"resource_ref"`
	Scope         []string    `json:"scope"`
	Expiry        string      `json:"expiry"`
	PolicyVersion string      `json:"policy_version"`
	Audience      string      `json:"audience"`
	TraceID       string      `json:"trace_id"`
}

// IssuePreAuthorization issues a pre-authorization artifact; the artifact JSON
// is returned verbatim.
func (c *Client) IssuePreAuthorization(ctx context.Context, internalToken string, req PreAuthorizationRequest) (json.RawMessage, error) {
	if strings.TrimSpace(internalToken) == "" || strings.TrimSpace(req.BearerToken) == "" ||
		strings.TrimSpace(req.Action) == "" || strings.TrimSpace(req.Resource.Kind) == "" {
		return nil, pepsdk.ErrInvalidInput
	}
	traceID := req.TraceID
	if traceID == "" {
		traceID = httpx.GenerateTraceID()
	}
	body, err := json.Marshal(preAuthIssueWire{
		BearerToken: req.BearerToken, Action: req.Action, Resource: req.Resource, Scope: req.Scope,
		Expiry: req.Expiry, PolicyVersion: req.PolicyVersion, Audience: req.Audience, TraceID: traceID,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", pepsdk.ErrUnavailable, err)
	}
	return c.authorityJSON(ctx, "/v1/pre-authorizations/issue", internalToken, body)
}

// TokenExchangeRequest is the input to ExchangeToken (v1alpha1).
type TokenExchangeRequest struct {
	SubjectToken     string
	SubjectTokenType string
	Scope            []string
	Audience         string
	Action           string
	Resource         ResourceRef
	SignedDecision   string
	TraceID          string
}

type exchangeWire struct {
	SubjectToken     string      `json:"subject_token"`
	SubjectTokenType string      `json:"subject_token_type"`
	Scope            []string    `json:"scope"`
	Audience         string      `json:"audience"`
	Action           string      `json:"action"`
	Resource         ResourceRef `json:"resource_ref"`
	SignedDecision   string      `json:"signed_decision"`
	TraceID          string      `json:"trace_id"`
}

type exchangeResponse struct {
	Grant string `json:"grant"`
}

// ExchangeToken exchanges a subject token for a grant token
// (POST /v1/token-exchange). The caller presents the AEGIVELA internal token.
func (c *Client) ExchangeToken(ctx context.Context, internalToken string, req TokenExchangeRequest) (string, error) {
	if strings.TrimSpace(internalToken) == "" || strings.TrimSpace(req.SubjectToken) == "" ||
		strings.TrimSpace(req.SubjectTokenType) == "" || len(req.Scope) == 0 ||
		strings.TrimSpace(req.Audience) == "" || strings.TrimSpace(req.Action) == "" ||
		strings.TrimSpace(req.Resource.Kind) == "" || strings.TrimSpace(req.SignedDecision) == "" {
		return "", pepsdk.ErrInvalidInput
	}
	traceID := req.TraceID
	if traceID == "" {
		traceID = httpx.GenerateTraceID()
	}
	body, err := json.Marshal(exchangeWire{
		SubjectToken: req.SubjectToken, SubjectTokenType: req.SubjectTokenType, Scope: req.Scope,
		Audience: req.Audience, Action: req.Action, Resource: req.Resource,
		SignedDecision: req.SignedDecision, TraceID: traceID,
	})
	if err != nil {
		return "", fmt.Errorf("%w: %v", pepsdk.ErrUnavailable, err)
	}
	raw, err := c.authorityJSON(ctx, "/v1/token-exchange", internalToken, body)
	if err != nil {
		return "", err
	}
	var out exchangeResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", fmt.Errorf("%w: %v", pepsdk.ErrUnavailable, err)
	}
	if strings.TrimSpace(out.Grant) == "" {
		return "", fmt.Errorf("%w: empty grant", pepsdk.ErrUnavailable)
	}
	return out.Grant, nil
}

func (c *Client) authorityJSON(ctx context.Context, path, internalToken string, body []byte) (json.RawMessage, error) {
	resp, err := c.HC.Request(ctx, http.MethodPost, path, internalHeader(internalToken), body)
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
