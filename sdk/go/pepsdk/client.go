package pepsdk

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const maxResponseBytes = 1 << 20

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string, httpClient *http.Client) *Client {
	return &Client{baseURL: baseURL, httpClient: httpClient}
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
		apiVersion = "aegivela.io/v1alpha1"
	}

	reqBody := evaluateRequest{
		APIVersion:           apiVersion,
		AuthorizationMode:    input.Mode,
		Principal:            minimalPrincipal{},
		BearerToken:          input.BearerToken,
		WorkloadAssertion:    input.WorkloadAssertion,
		ParentExecutionGrant: input.ParentExecutionGrant,
		Action:               route.Action,
		Resource: Resource{
			Kind:      route.Kind,
			Reference: route.Reference,
		},
		RequestedScope:     route.Scope,
		Audience:           route.Audience,
		RequestedExpiresAt: time.Now().Add(5 * time.Minute),
		TaskBinding:        input.TaskBinding,
		TraceID:            generateTraceID(),
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}

	url := c.baseURL + "/v1/policy/decisions/evaluate"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer resp.Body.Close()

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

	limitedBody := io.LimitReader(resp.Body, maxResponseBytes)

	var evalResp evaluateResponse
	decoder := json.NewDecoder(limitedBody)
	if err := decoder.Decode(&evalResp); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return nil, ErrUnavailable
	}
	if bodyLeft, _ := io.ReadAll(resp.Body); len(bodyLeft) > 0 {
		return nil, ErrUnavailable
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
		DecisionID:    evalResp.DecisionID,
		PolicyVersion: evalResp.PolicyVersion,
		EvidenceRefs:  evalResp.EvidenceRefs,
		Outcome:       evalResp.Outcome,
	}
	return result, nil
}

func generateTraceID() string {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "fallback-trace-id"
	}
	return hex.EncodeToString(buf[:])
}
