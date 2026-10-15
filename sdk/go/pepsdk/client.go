package pepsdk

import (
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

// Option configures a client before it is shared between requests.
type Option func(*Client)

// WithHeaders copies default request headers. Endpoint-specific headers win.
// Cross-origin redirects are refused when defaults are configured.
func WithHeaders(headers map[string]string) Option {
	defaults := make(map[string]string, len(headers))
	for key, value := range headers {
		defaults[key] = value
	}
	return func(c *Client) { c.hc.WithHeaders(defaults) }
}

// NewClient creates an AEGIVELA client for baseURL.
func NewClient(baseURL string, httpClient *http.Client, opts ...Option) *Client {
	c := &Client{hc: httpx.New(baseURL, httpClient)}
	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}
	return c
}

// WithTenant sets the tenant whose revocation scope the PEP rechecks. The
// revocation-check contract requires it; requests cannot widen this scope.
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
	APIVersion           string   `json:"api_version"`
	DecisionID           string   `json:"decision_id"`
	Outcome              string   `json:"outcome"`
	PolicyVersion        string   `json:"policy_version"`
	EvidenceRefs         []string `json:"evidence_refs"`
	SignedDecision       string   `json:"signedDecision,omitempty"`
	SignedDecisionLegacy string   `json:"signed_decision,omitempty"`
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

type riskContextV2 struct {
	Level string `json:"riskLevel,omitempty"`
	Ref   string `json:"riskRef,omitempty"`
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
	SignedDecision           string    `json:"signedDecision,omitempty"`
	SignedDecisionLegacy     string    `json:"signed_decision,omitempty"`
}

type errorEnvelopeV2 struct {
	APIVersion string `json:"apiVersion,omitempty"`
	Code       string `json:"code,omitempty"`
	ErrorCode  string `json:"errorCode,omitempty"`
}

// Authorize evaluates the request and rejects every outcome other than allow.
// SignedDecision is exposed but is not cryptographically verified here.
func (c *Client) Authorize(ctx context.Context, route Route, input AuthorizationInput) (*Result, error) {
	evaluated, err := c.Evaluate(ctx, route, input)
	if err != nil {
		return nil, err
	}
	if evaluated.Outcome != "allow" {
		return nil, ErrDenied
	}
	return &Result{
		DecisionID: evaluated.DecisionID, PolicyVersion: evaluated.PolicyVersion,
		EvidenceRefs: evaluated.EvidenceRefs, Outcome: evaluated.Outcome,
		SignedDecision: evaluated.SignedDecision,
	}, nil
}

func buildV2Principal(input AuthorizationInput) json.RawMessage {
	if len(input.Principal) > 0 {
		return input.Principal
	}
	b, _ := json.Marshal(minimalPrincipalV2{Namespace: input.Namespace, ActorRef: input.ActorRef})
	return b
}

func buildRiskContext(input AuthorizationInput) *riskContextV2 {
	if input.RiskLevel == "" && input.RiskRef == "" {
		return nil
	}
	return &riskContextV2{Level: input.RiskLevel, Ref: input.RiskRef}
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

func decodeV2Evaluate(resp *http.Response) (*EvaluateResult, error) {
	var wire evaluateResponseV2
	if err := decodeStrictResponse(resp.Body, &wire); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrMalformed, err)
	}
	if wire.APIVersion != APIVersionV2 || wire.DecisionID == "" || len(wire.EvidenceRefs) == 0 {
		return nil, ErrMalformed
	}
	if (wire.Result == "") == (wire.Outcome == "") {
		return nil, ErrMalformed
	}
	outcome := wire.Result
	if outcome == "" {
		outcome = wire.Outcome
		if wire.PolicyVersion == "" {
			return nil, ErrMalformed
		}
	}
	return &EvaluateResult{
		DecisionID: wire.DecisionID, PolicyVersion: wire.PolicyVersion, EvidenceRefs: wire.EvidenceRefs,
		Outcome: outcome, SignedDecision: firstNonEmpty(wire.SignedDecision, wire.SignedDecisionLegacy),
		ExpiresAt: wire.ExpiresAt, Obligations: wire.Obligations,
	}, nil
}

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
