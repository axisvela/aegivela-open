package pepsdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/axisrobo/aegivela-open/sdk/go/internal/httpx"
)

const maxResponseBytes = httpx.MaxResponseBytes

type Client struct {
	hc *httpx.Client
}

// NewClient creates an AEGIVELA client for baseURL.
func NewClient(baseURL string, httpClient *http.Client) *Client {
	return &Client{hc: httpx.New(baseURL, httpClient)}
}

// WithTenant sets the tenant whose revocation scope the PEP rechecks. The
// revocation-check contract requires it, and a PEP must not be able to widen
// its scope through request fields.
func (c *Client) WithTenant(tenantID string) *Client {
	c.hc.WithTenant(tenantID)
	return c
}

type evaluateRequest struct {
	APIVersion           string           `json:"api_version"`
	AuthorizationMode    Mode             `json:"authorization_mode"`
	Principal            minimalPrincipal `json:"principal"`
	BearerToken          string           `json:"bearer_token,omitempty"`
	WorkloadAssertion    string           `json:"workload_assertion,omitempty"`
	ParentExecutionGrant string           `json:"parent_execution_grant,omitempty"`
	Action               string           `json:"action"`
	Resource             Resource         `json:"resource"`
	RequestedScope       []string         `json:"requested_scope"`
	Audience             string           `json:"audience"`
	RequestedExpiresAt   time.Time        `json:"requested_expires_at"`
	TaskBinding          string           `json:"task_binding"`
	TraceID              string           `json:"trace_id"`
}

type minimalPrincipal struct {
	TenantID string `json:"tenant_id"`
	ActorRef string `json:"actor_ref"`
}

type evaluateResponse struct {
	APIVersion    string   `json:"api_version"`
	DecisionID    string   `json:"decision_id"`
	Outcome       string   `json:"outcome"`
	PolicyVersion string   `json:"policy_version"`
	EvidenceRefs  []string `json:"evidence_refs"`
	// SignedDecision carries the signed-decision artifact (either spelling).
	SignedDecision       string `json:"signedDecision,omitempty"`
	SignedDecisionLegacy string `json:"signed_decision,omitempty"`
}

type evaluateRequestV2 struct {
	APIVersion           string          `json:"apiVersion"`
	AuthorizationMode    string          `json:"authorizationMode"`
	Principal            json.RawMessage `json:"principal"`
	BearerToken          string          `json:"bearerToken,omitempty"`
	WorkloadAssertion    string          `json:"workloadAssertion,omitempty"`
	ParentExecutionGrant string          `json:"parentExecutionGrant,omitempty"`
	Action               string          `json:"action"`
	Resource             Resource        `json:"resource"`
	RequestedScope       []string        `json:"requestedScope"`
	Audience             string          `json:"audience"`
	RequestedExpiresAt   time.Time       `json:"requestedExpiresAt"`
	TaskBinding          string          `json:"taskBinding"`
	TraceID              string          `json:"traceId"`
	RiskContext          *riskContextV2  `json:"riskContext,omitempty"`
}

// riskContextV2 is the optional risk hint carried on a v2 evaluate request.
type riskContextV2 struct {
	Level string `json:"riskLevel,omitempty"`
	Ref   string `json:"riskRef,omitempty"`
}

type authorizationInputV2 struct {
	APIVersion           string   `json:"apiVersion"`
	Mode                 string   `json:"mode"`
	BearerToken          string   `json:"bearerToken,omitempty"`
	WorkloadAssertion    string   `json:"workloadAssertion,omitempty"`
	ParentExecutionGrant string   `json:"parentExecutionGrant,omitempty"`
	RequestedScope       []string `json:"requestedScope"`
	Audience             string   `json:"audience"`
	TraceID              string   `json:"traceId"`
}

type minimalPrincipalV2 struct {
	Namespace string `json:"namespace"`
	ActorRef  string `json:"actorRef"`
}

type evaluateResponseV2 struct {
	APIVersion               string    `json:"apiVersion"`
	StructuredResourceDigest string    `json:"structuredResourceDigest,omitempty"`
	DecisionID               string    `json:"decisionId"`
	Result                   string    `json:"result,omitempty"`
	Outcome                  string    `json:"outcome,omitempty"`
	PolicyVersion            string    `json:"policyVersion,omitempty"`
	ExpiresAt                time.Time `json:"expiresAt,omitempty"`
	Obligations              []string  `json:"obligations,omitempty"`
	EvidenceRefs             []string  `json:"evidenceRefs"`
	// SignedDecision carries the signed-decision artifact (either spelling).
	SignedDecision       string `json:"signedDecision,omitempty"`
	SignedDecisionLegacy string `json:"signed_decision,omitempty"`
}

type errorEnvelopeV2 struct {
	APIVersion string `json:"apiVersion,omitempty"`
	Code       string `json:"code,omitempty"`
	ErrorCode  string `json:"errorCode,omitempty"`
}

func (c *Client) Authorize(ctx context.Context, route Route, input AuthorizationInput) (*Result, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
	if err := route.Validate(); err != nil {
		return nil, err
	}
	if len(route.Modes) > 0 {
		allowed := false
		for _, m := range route.Modes {
			if m == input.Mode {
				allowed = true
				break
			}
		}
		if !allowed {
			return nil, ErrInvalidInput
		}
	}

	apiVersion := input.APIVersion
	if apiVersion == "" {
		apiVersion = APIVersionV1Alpha1
	}
	// The v2.0 evaluate contract cannot express the agent principal binding and
	// parent authority that service/twin agent modes require, so those modes are
	// rejected up front rather than sent as requests that can never succeed.
	if apiVersion == APIVersionV2 && (input.Mode == ModeServiceAgentAPI || input.Mode == ModeTwinAgentAPI) {
		return nil, ErrInvalidInput
	}

	var reqBody any
	if apiVersion == APIVersionV2 {
		pepInput := newAuthorizationInputV2(route, input, generateTraceID())
		reqBody = evaluateRequestV2{
			APIVersion: pepInput.APIVersion, AuthorizationMode: pepInput.Mode, Principal: buildV2Principal(input),
			BearerToken: pepInput.BearerToken, WorkloadAssertion: pepInput.WorkloadAssertion,
			ParentExecutionGrant: pepInput.ParentExecutionGrant, Action: route.Action,
			Resource: Resource{Kind: route.Kind, Reference: route.Reference}, RequestedScope: pepInput.RequestedScope,
			Audience: pepInput.Audience, RequestedExpiresAt: time.Now().Add(5 * time.Minute),
			TaskBinding: input.TaskBinding, TraceID: pepInput.TraceID, RiskContext: buildRiskContext(input),
		}
	} else {
		reqBody = evaluateRequest{
			APIVersion: apiVersion, AuthorizationMode: input.Mode, Principal: minimalPrincipal{},
			BearerToken: input.BearerToken, WorkloadAssertion: input.WorkloadAssertion,
			ParentExecutionGrant: input.ParentExecutionGrant, Action: route.Action,
			Resource: Resource{Kind: route.Kind, Reference: route.Reference}, RequestedScope: route.Scope,
			Audience: route.Audience, RequestedExpiresAt: time.Now().Add(5 * time.Minute),
			TaskBinding: input.TaskBinding, TraceID: generateTraceID(),
		}
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}

	url := c.hc.BaseURL + "/v1/policy/decisions/evaluate"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.hc.HTTP.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer resp.Body.Close()

	if apiVersion == APIVersionV2 && resp.StatusCode != http.StatusOK {
		return nil, decodeV2Error(resp)
	}

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return nil, ErrUnauthenticated
	case http.StatusForbidden:
		return nil, ErrDenied
	case http.StatusOK:
	default:
		if resp.StatusCode == http.StatusServiceUnavailable {
			return nil, ErrUnavailable
		}
		return nil, ErrUnavailable
	}

	if apiVersion == APIVersionV2 {
		return decodeV2Result(resp)
	}

	var evalResp evaluateResponse
	if err := decodeBoundedResponse(resp.Body, &evalResp, false); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}

	if evalResp.DecisionID == "" {
		return nil, fmt.Errorf("%w: empty decision ID", ErrUnavailable)
	}
	if evalResp.PolicyVersion == "" {
		return nil, fmt.Errorf("%w: empty policy version", ErrUnavailable)
	}

	if evalResp.Outcome != "allow" {
		return nil, ErrDenied
	}

	result := &Result{
		DecisionID:     evalResp.DecisionID,
		PolicyVersion:  evalResp.PolicyVersion,
		EvidenceRefs:   evalResp.EvidenceRefs,
		Outcome:        evalResp.Outcome,
		SignedDecision: firstNonEmpty(evalResp.SignedDecision, evalResp.SignedDecisionLegacy),
	}
	return result, nil
}

// buildV2Principal returns the wire principal for a v2 evaluate request: the
// caller-supplied raw principal when set, else a minimal one.
func buildV2Principal(input AuthorizationInput) json.RawMessage {
	if len(input.Principal) > 0 {
		return input.Principal
	}
	b, _ := json.Marshal(minimalPrincipalV2{Namespace: input.Namespace, ActorRef: input.ActorRef})
	return b
}

// buildRiskContext returns the optional risk context for a v2 evaluate request.
func buildRiskContext(input AuthorizationInput) *riskContextV2 {
	if input.RiskLevel == "" && input.RiskRef == "" {
		return nil
	}
	return &riskContextV2{Level: input.RiskLevel, Ref: input.RiskRef}
}

func newAuthorizationInputV2(route Route, input AuthorizationInput, traceID string) authorizationInputV2 {
	return authorizationInputV2{
		APIVersion: input.APIVersion, Mode: modeV2(input.Mode), BearerToken: input.BearerToken,
		WorkloadAssertion: input.WorkloadAssertion, ParentExecutionGrant: input.ParentExecutionGrant,
		RequestedScope: route.Scope, Audience: route.Audience, TraceID: traceID,
	}
}

func modeV2(mode Mode) string {
	switch mode {
	case ModeHumanWeb:
		return "humanWeb"
	case ModeSystemAPI:
		return "systemApi"
	case ModeDelegatedAPI:
		return "delegatedApi"
	case ModeServiceAgentAPI:
		return "serviceAgentApi"
	case ModeTwinAgentAPI:
		return "twinAgentApi"
	default:
		return string(mode)
	}
}

func decodeV2Result(resp *http.Response) (*Result, error) {
	var wire evaluateResponseV2
	if err := decodeStrictResponse(resp.Body, &wire); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	if wire.APIVersion != APIVersionV2 || wire.DecisionID == "" || len(wire.EvidenceRefs) == 0 {
		return nil, ErrUnavailable
	}
	if (wire.Result == "") == (wire.Outcome == "") {
		return nil, ErrUnavailable
	}
	outcome := wire.Result
	if outcome == "" {
		outcome = wire.Outcome
		if wire.PolicyVersion == "" {
			return nil, ErrUnavailable
		}
	}
	if outcome != "allow" {
		return nil, ErrDenied
	}
	return &Result{DecisionID: wire.DecisionID, PolicyVersion: wire.PolicyVersion, EvidenceRefs: wire.EvidenceRefs, Outcome: outcome, SignedDecision: firstNonEmpty(wire.SignedDecision, wire.SignedDecisionLegacy)}, nil
}

// firstNonEmpty returns the first non-empty string.
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func decodeV2Error(resp *http.Response) error {
	var wire errorEnvelopeV2
	if err := decodeStrictResponse(resp.Body, &wire); err != nil {
		return ErrUnavailable
	}
	if wire.Code != "" {
		if wire.APIVersion != APIVersionV2 || wire.ErrorCode != "" {
			return ErrUnavailable
		}
		switch wire.Code {
		case "unauthenticated":
			if resp.StatusCode == http.StatusUnauthorized {
				return ErrUnauthenticated
			}
		case "invalidPepRequest":
			if resp.StatusCode == http.StatusBadRequest {
				return ErrDenied
			}
		case "authorizationDenied":
			if resp.StatusCode == http.StatusForbidden {
				return ErrDenied
			}
		case "authorizationUnavailable":
			if resp.StatusCode == http.StatusServiceUnavailable {
				return ErrUnavailable
			}
		}
		return ErrUnavailable
	}
	if wire.ErrorCode == "" || wire.APIVersion != "" {
		return ErrUnavailable
	}
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		if wire.ErrorCode == "unauthenticated" || strings.HasPrefix(wire.ErrorCode, "invalid") {
			return ErrUnauthenticated
		}
	case http.StatusBadRequest, http.StatusForbidden:
		if wire.ErrorCode == "authorizationDenied" || strings.HasSuffix(wire.ErrorCode, "Denied") || strings.HasSuffix(wire.ErrorCode, "Mismatch") {
			return ErrDenied
		}
	case http.StatusServiceUnavailable:
		if strings.HasSuffix(wire.ErrorCode, "Unavailable") {
			return ErrUnavailable
		}
	}
	return ErrUnavailable
}

func decodeStrictResponse(body io.Reader, target any) error {
	return httpx.DecodeStrict(body, target)
}

func decodeBoundedResponse(body io.Reader, target any, strict bool) error {
	return httpx.DecodeBounded(body, target, strict)
}

func generateTraceID() string { return httpx.GenerateTraceID() }
