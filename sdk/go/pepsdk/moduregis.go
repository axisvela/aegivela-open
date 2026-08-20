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

// ModuregisAuthorizeRequest carries the control-plane context required for MODUREGIS
// to issue a short-lived authorization for a capability invocation.
type ModuregisAuthorizeRequest struct {
	Action              string   `json:"action"`
	ResourceKind        string   `json:"resource_kind"`
	ResourceReference   string   `json:"resource_reference"`
	BearerToken         string   `json:"bearer_token"`
	Scope               []string `json:"scope"`
	ApprovalArtifact    string   `json:"approval_artifact"`
	AdapterID           string   `json:"adapter_id"`
	AdapterVersion      string   `json:"adapter_version"`
	ExecutionID         string   `json:"execution_id"`
	TraceID             string   `json:"trace_id"`
	ToolID              string   `json:"tool_id"`
	SkillHash           string   `json:"skill_hash"`
	ImplementationDigest string  `json:"implementation_digest"`
}

// ModuregisAuthorizeDecision is the signed authorization returned by MODUREGIS.
type ModuregisAuthorizeDecision struct {
	DecisionID        string    `json:"decision_id"`
	Outcome           string    `json:"outcome"`
	PolicyVersion     string    `json:"policy_version"`
	TenantID          string    `json:"tenant_id"`
	ActorID           string    `json:"actor_id"`
	AgentID           string    `json:"agent_id"`
	MasterID          string    `json:"master_id"`
	WorkloadID        string    `json:"workload_id"`
	SubjectRef        string    `json:"subject_ref"`
	EvidenceRefs      []string  `json:"evidence_refs"`
	ExpiresAt         time.Time `json:"expires_at"`
	GrantToken        string    `json:"grant_token"`
}

// ModuregisAuthorize requests a short-lived authorization from MODUREGIS for a single
// capability invocation. internalToken is the PEP's own internal bootstrap token, sent as
// the X-AEGIVELA-PEP header; it is distinct from the caller BearerToken in the request body.
func (c *Client) ModuregisAuthorize(ctx context.Context, internalToken string, request ModuregisAuthorizeRequest) (*ModuregisAuthorizeDecision, error) {
	if internalToken == "" {
		return nil, ErrInvalidInput
	}
	if request.TraceID == "" {
		request.TraceID = generateTraceID()
	}
	if request.Action == "" || request.ResourceKind == "" || request.ResourceReference == "" {
		return nil, ErrInvalidInput
	}

	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/moduregis/authorize", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-AEGIVELA-PEP", internalToken)

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
	default:
		return nil, ErrUnavailable
	}

	var decision ModuregisAuthorizeDecision
	decoder := json.NewDecoder(io.LimitReader(resp.Body, maxResponseBytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&decision); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}

	if decision.Outcome != "allow" {
		return nil, ErrDenied
	}
	return &decision, nil
}
