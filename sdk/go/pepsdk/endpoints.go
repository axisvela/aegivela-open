package pepsdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ToolInvokeRequest is the input to ToolInvoke (v1alpha1 contract).
type ToolInvokeRequest struct {
	BearerToken          string
	Resource             Resource
	Scope                []string
	ApprovalArtifact     string
	ExecutionID          string
	ToolID               string
	SkillHash            string
	ImplementationDigest string
	TraceID              string
}

// ToolInvokeResult is a tool:invoke decision plus the issued grant token.
type ToolInvokeResult struct {
	APIVersion    string
	DecisionID    string
	Outcome       string
	PolicyVersion string
	TenantID      string
	ActorID       string
	ToolID        string
	GrantToken    string
	EvidenceRefs  []string
}

type toolInvokeWire struct {
	APIVersion           string   `json:"api_version"`
	BearerToken          string   `json:"bearer_token"`
	Action               string   `json:"action"`
	Resource             Resource `json:"resource"`
	Scope                []string `json:"scope,omitempty"`
	ApprovalArtifact     string   `json:"approval_artifact,omitempty"`
	ExecutionID          string   `json:"execution_id,omitempty"`
	ToolID               string   `json:"tool_id,omitempty"`
	SkillHash            string   `json:"skill_hash,omitempty"`
	ImplementationDigest string   `json:"implementation_digest,omitempty"`
	TraceID              string   `json:"trace_id"`
}

type toolInvokeResponse struct {
	APIVersion    string   `json:"api_version"`
	DecisionID    string   `json:"decision_id"`
	Outcome       string   `json:"outcome"`
	PolicyVersion string   `json:"policy_version"`
	TenantID      string   `json:"tenant_id"`
	ActorID       string   `json:"actor_id"`
	ToolID        string   `json:"tool_id"`
	GrantToken    string   `json:"grant_token"`
	EvidenceRefs  []string `json:"evidence_refs"`
}

// ToolInvoke obtains a tool:invoke authorization decision and mints the
// execution grant. The caller presents the AEGIVELA internal token.
func (c *Client) ToolInvoke(ctx context.Context, internalToken string, req ToolInvokeRequest) (*ToolInvokeResult, error) {
	if strings.TrimSpace(internalToken) == "" ||
		strings.TrimSpace(req.BearerToken) == "" ||
		strings.TrimSpace(req.Resource.Kind) == "" || strings.TrimSpace(req.Resource.Reference) == "" ||
		len(req.Scope) == 0 || strings.TrimSpace(req.ExecutionID) == "" ||
		strings.TrimSpace(req.ToolID) == "" || strings.TrimSpace(req.SkillHash) == "" ||
		strings.TrimSpace(req.ImplementationDigest) == "" {
		return nil, ErrInvalidInput
	}
	traceID := req.TraceID
	if traceID == "" {
		traceID = generateTraceID()
	}
	body, err := json.Marshal(toolInvokeWire{
		APIVersion: APIVersionV1Alpha1, BearerToken: req.BearerToken, Action: "tool:invoke",
		Resource: req.Resource, Scope: req.Scope, ApprovalArtifact: req.ApprovalArtifact,
		ExecutionID: req.ExecutionID, ToolID: req.ToolID, SkillHash: req.SkillHash,
		ImplementationDigest: req.ImplementationDigest, TraceID: traceID,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	resp, err := c.postInternal(ctx, "/v1/tool/invoke", internalToken, body)
	if err != nil {
		return nil, err
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
	var wire toolInvokeResponse
	if err := decodeBoundedResponse(resp.Body, &wire, false); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return &ToolInvokeResult{
		APIVersion: wire.APIVersion, DecisionID: wire.DecisionID, Outcome: wire.Outcome,
		PolicyVersion: wire.PolicyVersion, TenantID: wire.TenantID, ActorID: wire.ActorID,
		ToolID: wire.ToolID, GrantToken: wire.GrantToken, EvidenceRefs: wire.EvidenceRefs,
	}, nil
}

// DecisionVerifyRequest is the input to VerifyDecision (v2 contract).
type DecisionVerifyRequest struct {
	SignedDecision string
	// Principal is the expected principal binding (raw JSON: shape is
	// AEGIVELA-owned, e.g. {"namespace":..,"actorRef":..}).
	Principal json.RawMessage
	Action    string
	Resource  Resource
	Scope     []string
	Audience  string
	TraceID   string
}

// VerifiedDecision is a verified signed decision.
type VerifiedDecision struct {
	APIVersion    string
	DecisionID    string
	Valid         bool
	Outcome       string
	PolicyVersion string
	ExpiresAt     time.Time
	Obligations   []string
	EvidenceRefs  []string
}

type decisionVerifyBindingV2 struct {
	Principal json.RawMessage `json:"principal,omitempty"`
	Action    string          `json:"action"`
	Resource  Resource        `json:"resource"`
	Scope     []string        `json:"scope,omitempty"`
	Audience  string          `json:"audience,omitempty"`
	TraceID   string          `json:"traceId,omitempty"`
}

type decisionVerifyRequestV2 struct {
	APIVersion       string                  `json:"apiVersion"`
	SignedDecision   string                  `json:"signedDecision"`
	ExpectedBindings decisionVerifyBindingV2 `json:"expectedBindings"`
}

type decisionVerifyResponseV2 struct {
	APIVersion    string    `json:"apiVersion"`
	DecisionID    string    `json:"decisionId"`
	Valid         bool      `json:"valid"`
	Outcome       string    `json:"outcome"`
	PolicyVersion string    `json:"policyVersion"`
	ExpiresAt     time.Time `json:"expiresAt"`
	Obligations   []string  `json:"obligations,omitempty"`
	EvidenceRefs  []string  `json:"evidenceRefs"`
}

// VerifyDecision verifies a signed policy decision against expected bindings
// (v2 contract). The caller presents the AEGIVELA internal token.
func (c *Client) VerifyDecision(ctx context.Context, internalToken string, req DecisionVerifyRequest) (*VerifiedDecision, error) {
	if strings.TrimSpace(internalToken) == "" || strings.TrimSpace(req.SignedDecision) == "" ||
		strings.TrimSpace(req.Action) == "" || strings.TrimSpace(req.Resource.Kind) == "" ||
		strings.TrimSpace(req.Resource.Reference) == "" {
		return nil, ErrInvalidInput
	}
	body, err := json.Marshal(decisionVerifyRequestV2{
		APIVersion: APIVersionV2, SignedDecision: req.SignedDecision,
		ExpectedBindings: decisionVerifyBindingV2{
			Principal: req.Principal, Action: req.Action, Resource: req.Resource,
			Scope: req.Scope, Audience: req.Audience, TraceID: req.TraceID,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	resp, err := c.postInternal(ctx, "/v1/policy/decisions/verify", internalToken, body)
	if err != nil {
		return nil, err
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
	var wire decisionVerifyResponseV2
	if err := decodeBoundedResponse(resp.Body, &wire, false); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return &VerifiedDecision{
		APIVersion: wire.APIVersion, DecisionID: wire.DecisionID, Valid: wire.Valid,
		Outcome: wire.Outcome, PolicyVersion: wire.PolicyVersion, ExpiresAt: wire.ExpiresAt,
		Obligations: wire.Obligations, EvidenceRefs: wire.EvidenceRefs,
	}, nil
}

// DecisionJWKS returns the raw JWKS used to verify signed decisions.
func (c *Client) DecisionJWKS(ctx context.Context) (json.RawMessage, error) {
	return c.getJWKS(ctx, "/v1/policy/decisions/jwks")
}

// RevocationList returns the policy revocation events published since `since`
// (GET /v1/revocations). The events are returned verbatim.
func (c *Client) RevocationList(ctx context.Context, internalToken string, since time.Time) ([]json.RawMessage, error) {
	if strings.TrimSpace(internalToken) == "" {
		return nil, ErrInvalidInput
	}
	path := "/v1/revocations"
	if !since.IsZero() {
		path += "?since=" + url.QueryEscape(since.UTC().Format(time.RFC3339))
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.hc.BaseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	httpReq.Header.Set("X-Aegivela-Internal-Token", internalToken)
	resp, err := c.hc.HTTP.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return nil, ErrUnauthenticated
	case http.StatusOK:
	default:
		return nil, ErrUnavailable
	}
	var out struct {
		Events []json.RawMessage `json:"events"`
	}
	if err := decodeBoundedResponse(resp.Body, &out, false); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return out.Events, nil
}

// postInternal POSTs a JSON body with the internal-token header.
func (c *Client) postInternal(ctx context.Context, path, internalToken string, body []byte) (*http.Response, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.hc.BaseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Aegivela-Internal-Token", internalToken)
	resp, err := c.hc.HTTP.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return resp, nil
}

// getJWKS fetches a raw JWKS document (unauthenticated GET).
func (c *Client) getJWKS(ctx context.Context, path string) (json.RawMessage, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.hc.BaseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	resp, err := c.hc.HTTP.Do(httpReq)
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
