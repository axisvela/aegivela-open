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

// ModuregisAuthorizeRequest mirrors the MODUREGIS adapter request contract.
// The v1alpha1 wire contract carries no namespace field; namespace isolation
// is enforced by the MODUREGIS registry against the decision's tenant.
type ModuregisAuthorizeRequest struct {
	Action               string
	ResourceKind         string
	ResourceReference    string
	BearerToken          string
	Scope                []string
	ApprovalArtifact     string
	AdapterID            string
	AdapterVersion       string
	ExecutionID          string
	TraceID              string
	ToolID               string
	SkillHash            string
	ImplementationDigest string
}

func (r ModuregisAuthorizeRequest) valid() bool {
	if r.BearerToken == "" || r.Action == "" || r.ResourceKind == "" || r.ResourceReference == "" {
		return false
	}
	switch r.Action {
	case "capability:read":
	case "capability:publish":
		if len(r.Scope) == 0 || r.ApprovalArtifact == "" {
			return false
		}
	case "adapter:activate":
		if r.AdapterID == "" || r.AdapterVersion == "" {
			return false
		}
	case "capability:invoke":
		if len(r.Scope) == 0 || r.ExecutionID == "" {
			return false
		}
	case "tool:invoke":
		if len(r.Scope) == 0 || r.ExecutionID == "" || r.ToolID == "" || r.SkillHash == "" || r.ImplementationDigest == "" {
			return false
		}
	default:
		return false
	}
	return true
}

// ModuregisAuthorizeDecision is the typed adapter decision result. The
// principal envelope follows the published v1alpha1 response schema.
type ModuregisAuthorizeDecision struct {
	DecisionID    string
	Outcome       string
	PolicyVersion string
	TenantID      string
	ActorID       string
	AgentID       string
	MasterID      string
	WorkloadID    string
	SubjectRef    string
	EvidenceRefs  []string
	ExpiresAt     time.Time
	GrantToken    string
}

// ModuregisAuthorize obtains an adapter authorization decision from the MODUREGIS endpoint.
func (c *Client) ModuregisAuthorize(ctx context.Context, internalToken string, request ModuregisAuthorizeRequest) (*ModuregisAuthorizeDecision, error) {
	if internalToken == "" || !request.valid() {
		return nil, ErrInvalidInput
	}
	traceID := request.TraceID
	if traceID == "" {
		traceID = generateTraceID()
	}
	body := map[string]any{
		"api_version":  "aegivela.io/v1alpha1",
		"bearer_token": request.BearerToken,
		"action":       request.Action,
		"resource": map[string]string{
			"kind":      request.ResourceKind,
			"reference": request.ResourceReference,
		},
		"trace_id": traceID,
	}
	if len(request.Scope) > 0 {
		body["scope"] = request.Scope
	}
	if request.ApprovalArtifact != "" {
		body["approval_artifact"] = request.ApprovalArtifact
	}
	if request.AdapterID != "" {
		body["adapter_id"] = request.AdapterID
	}
	if request.AdapterVersion != "" {
		body["adapter_version"] = request.AdapterVersion
	}
	if request.ExecutionID != "" {
		body["execution_id"] = request.ExecutionID
	}
	if request.ToolID != "" {
		body["tool_id"] = request.ToolID
	}
	if request.SkillHash != "" {
		body["skill_hash"] = request.SkillHash
	}
	if request.ImplementationDigest != "" {
		body["implementation_digest"] = request.ImplementationDigest
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/moduregis/authorize", bytes.NewReader(payload))
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
		APIVersion    string    `json:"api_version"`
		DecisionID    string    `json:"decision_id"`
		Outcome       string    `json:"outcome"`
		PolicyVersion string    `json:"policy_version"`
		TenantID      string    `json:"tenant_id"`
		ActorID       string    `json:"actor_id"`
		AgentID       string    `json:"agent_id"`
		MasterID      string    `json:"master_id"`
		WorkloadID    string    `json:"workload_id"`
		SubjectRef    string    `json:"subject_ref"`
		EvidenceRefs  []string  `json:"evidence_refs"`
		ExpiresAt     time.Time `json:"expires_at"`
		GrantToken    string    `json:"grant_token"`
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
	return &ModuregisAuthorizeDecision{
		DecisionID:    decision.DecisionID,
		Outcome:       decision.Outcome,
		PolicyVersion: decision.PolicyVersion,
		TenantID:      decision.TenantID,
		ActorID:       decision.ActorID,
		AgentID:       decision.AgentID,
		MasterID:      decision.MasterID,
		WorkloadID:    decision.WorkloadID,
		SubjectRef:    decision.SubjectRef,
		EvidenceRefs:  decision.EvidenceRefs,
		ExpiresAt:     decision.ExpiresAt,
		GrantToken:    decision.GrantToken,
	}, nil
}
