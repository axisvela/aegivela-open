package pepsdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ToolAttribution binds a tool invocation to its four-hop authority chain.
type ToolAttribution struct {
	HumanRef             string    `json:"human_ref"`
	AgentRef             string    `json:"agent_ref"`
	ToolID               string    `json:"tool_id"`
	SkillHash            string    `json:"skill_hash"`
	ImplementationDigest string    `json:"implementation_digest"`
	SignedAt             time.Time `json:"signed_at,omitempty"`
	Signature            string    `json:"signature,omitempty"`
}

// GatewayConnectRequest mirrors the gateway connect request contract.
type GatewayConnectRequest struct {
	Action                 string
	AuthorizationMode      Mode
	AgentID                string
	AgentClass             string
	WorkloadAssertion      string
	TenantID               string
	TargetHost             string
	TargetScheme           string
	TargetPath             string
	RequestedScope         []string
	Audience               string
	ToolAttribution        *ToolAttribution
	ArgumentDigest         string
	ArgumentClassification string
	TraceID                string
}

func (r GatewayConnectRequest) valid() bool {
	if r.WorkloadAssertion == "" || r.Action == "" || r.TargetHost == "" ||
		(r.TargetScheme != "http" && r.TargetScheme != "https") ||
		len(r.RequestedScope) == 0 || r.Audience == "" {
		return false
	}
	if r.Action == "tool:invoke" {
		if r.ToolAttribution == nil || r.ToolAttribution.AgentRef == "" ||
			r.ArgumentDigest == "" || r.ArgumentClassification == "" {
			return false
		}
	}
	return r.validMode()
}

// validMode rejects a cross-mode substitution before the request leaves the
// PEP: the declared authorization mode must agree with the agent fields, and
// the agent fields must agree with each other when the mode is omitted.
func (r GatewayConnectRequest) validMode() bool {
	switch r.AuthorizationMode {
	case "":
		return (r.AgentID == "") == (r.AgentClass == "")
	case ModeSystemAPI:
		return r.AgentID == "" && r.AgentClass == ""
	case ModeServiceAgentAPI:
		return r.AgentID != "" && r.AgentClass == "service"
	case ModeTwinAgentAPI:
		return r.AgentID != "" && r.AgentClass == "twin"
	default:
		return false
	}
}

// GatewayConnectDecision is the typed connect decision result.
type GatewayConnectDecision struct {
	DecisionID         string
	Outcome            string
	PolicyVersion      string
	ExpiresAt          time.Time
	TTLSeconds         int
	OfflineEligible    bool
	Scope              []string
	Obligations        []string
	EvidenceRefs       []string
	CredentialRef      string
	CredentialClass    string
	MitmRequired       bool
	MitmScope          []string
	AllowedTargetPaths []string
}

// GatewayConnect obtains a short-lived connect decision from the Server Gateway endpoint.
func (c *Client) GatewayConnect(ctx context.Context, internalToken string, request GatewayConnectRequest) (*GatewayConnectDecision, error) {
	if internalToken == "" || !request.valid() {
		return nil, ErrInvalidInput
	}
	traceID := request.TraceID
	if traceID == "" {
		traceID = generateTraceID()
	}
	body := map[string]any{
		"api_version":        "aegivela.io/v1alpha1",
		"action":             request.Action,
		"agent_id":           request.AgentID,
		"agent_class":        request.AgentClass,
		"workload_assertion": request.WorkloadAssertion,
		"tenant_id":          request.TenantID,
		"target_host":        request.TargetHost,
		"target_scheme":      request.TargetScheme,
		"target_path":        request.TargetPath,
		"requested_scope":    request.RequestedScope,
		"audience":           request.Audience,
		"trace_id":           traceID,
	}
	if request.AuthorizationMode != "" {
		body["authorization_mode"] = request.AuthorizationMode
	}
	if request.ToolAttribution != nil {
		body["attribution"] = request.ToolAttribution
		body["argument_digest"] = request.ArgumentDigest
		body["argument_classification"] = request.ArgumentClassification
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/gateway/connect", bytes.NewReader(payload))
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
	case http.StatusOK:
	case http.StatusUnauthorized:
		return nil, ErrUnauthenticated
	case http.StatusForbidden:
		return nil, ErrDenied
	case http.StatusBadRequest:
		return nil, ErrInvalidInput
	default:
		return nil, ErrUnavailable
	}
	limited := io.LimitReader(resp.Body, maxResponseBytes)
	var decision struct {
		APIVersion         string    `json:"api_version"`
		DecisionID         string    `json:"decision_id"`
		Outcome            string    `json:"outcome"`
		PolicyVersion      string    `json:"policy_version"`
		ExpiresAt          time.Time `json:"expires_at"`
		TTLSeconds         int       `json:"ttl_seconds"`
		OfflineEligible    bool      `json:"offline_eligible"`
		Scope              []string  `json:"scope"`
		Obligations        []string  `json:"obligations"`
		EvidenceRefs       []string  `json:"evidence_refs"`
		CredentialRef      string    `json:"credential_ref"`
		CredentialClass    string    `json:"credential_class"`
		MitmRequired       bool      `json:"mitm_required"`
		MitmScope          []string  `json:"mitm_scope"`
		AllowedTargetPaths []string  `json:"allowed_target_paths"`
	}
	decoder := json.NewDecoder(limited)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&decision); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return nil, ErrUnavailable
	}
	if decision.DecisionID == "" || decision.PolicyVersion == "" {
		return nil, ErrUnavailable
	}
	if decision.Outcome != "allow" && decision.Outcome != "approval_required" {
		return nil, ErrDenied
	}
	return &GatewayConnectDecision{
		DecisionID:         decision.DecisionID,
		Outcome:            decision.Outcome,
		PolicyVersion:      decision.PolicyVersion,
		ExpiresAt:          decision.ExpiresAt,
		TTLSeconds:         decision.TTLSeconds,
		OfflineEligible:    decision.OfflineEligible,
		Scope:              decision.Scope,
		Obligations:        decision.Obligations,
		EvidenceRefs:       decision.EvidenceRefs,
		CredentialRef:      decision.CredentialRef,
		CredentialClass:    decision.CredentialClass,
		MitmRequired:       decision.MitmRequired,
		MitmScope:          decision.MitmScope,
		AllowedTargetPaths: decision.AllowedTargetPaths,
	}, nil
}
