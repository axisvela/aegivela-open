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

// ToolAttribution binds a gateway connection to the human owner, agent, tool, and the
// immutable skill/implementation digests that establish provenance.
type ToolAttribution struct {
	HumanRef            string `json:"human_ref"`
	AgentRef            string `json:"agent_ref"`
	ToolID              string `json:"tool_id"`
	SkillHash           string `json:"skill_hash"`
	ImplementationDigest string `json:"implementation_digest"`
	SignedAt            time.Time `json:"signed_at"`
	Signature           string `json:"signature"`
}

// GatewayConnectRequest requests a gateway connect authorization for an agent workload to
// reach an external egress target.
type GatewayConnectRequest struct {
	Action                string          `json:"action"`
	AgentID               string          `json:"agent_id"`
	AgentClass            string          `json:"agent_class"`
	WorkloadAssertion     string          `json:"workload_assertion"`
	TenantID              string          `json:"tenant_id"`
	TargetHost            string          `json:"target_host"`
	TargetScheme          string          `json:"target_scheme"`
	TargetPath            string          `json:"target_path"`
	RequestedScope        []string        `json:"requested_scope"`
	Audience              string          `json:"audience"`
	ToolAttribution       *ToolAttribution `json:"tool_attribution"`
	ArgumentDigest        string          `json:"argument_digest"`
	ArgumentClassification string         `json:"argument_classification"`
	TraceID               string          `json:"trace_id"`
}

// GatewayConnectDecision is the signed gateway token returned by MODUREGIS.
type GatewayConnectDecision struct {
	DecisionID    string    `json:"decision_id"`
	Outcome       string    `json:"outcome"`
	GatewayToken  string    `json:"gateway_token"`
	PolicyVersion string    `json:"policy_version"`
	EvidenceRefs  []string  `json:"evidence_refs"`
	ExpiresAt     time.Time `json:"expires_at"`
	ResolvedHost  string    `json:"resolved_host"`
	ResolvedScheme string   `json:"resolved_scheme"`
	ResolvedPath  string    `json:"resolved_path"`
}

// GatewayConnect requests a gateway connect authorization from MODUREGIS for an agent
// workload to reach an external egress target. internalToken is the PEP's bootstrap token
// sent as the X-AEGIVELA-PEP header; it is distinct from the WorkloadAssertion in the body.
func (c *Client) GatewayConnect(ctx context.Context, internalToken string, request GatewayConnectRequest) (*GatewayConnectDecision, error) {
	if internalToken == "" {
		return nil, ErrInvalidInput
	}
	if request.TraceID == "" {
		request.TraceID = generateTraceID()
	}
	if request.Action == "" || request.WorkloadAssertion == "" || request.TargetHost == "" {
		return nil, ErrInvalidInput
	}

	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/gateway/connect", bytes.NewReader(body))
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

	var decision GatewayConnectDecision
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
