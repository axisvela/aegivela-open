package pepsdk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// SimulateDecision runs the internal-token-gated policy evaluation
// (POST /v1/policy/decisions/simulate) for a route + authorization input.
func (c *Client) SimulateDecision(ctx context.Context, internalToken string, input AuthorizationInput, route Route) (*Result, error) {
	if strings.TrimSpace(internalToken) == "" {
		return nil, ErrInvalidInput
	}
	if err := input.Validate(); err != nil {
		return nil, err
	}
	if err := route.Validate(); err != nil {
		return nil, err
	}
	body, err := json.Marshal(evaluateRequest{
		APIVersion: APIVersionV1Alpha1, AuthorizationMode: input.Mode,
		Principal: minimalPrincipal{}, BearerToken: input.BearerToken,
		WorkloadAssertion: input.WorkloadAssertion, ParentExecutionGrant: input.ParentExecutionGrant,
		Action: route.Action, Resource: Resource{Kind: route.Kind, Reference: route.Reference},
		RequestedScope: route.Scope, Audience: route.Audience,
		RequestedExpiresAt: time.Now().Add(5 * time.Minute), TaskBinding: input.TaskBinding,
		TraceID: generateTraceID(),
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	resp, err := c.postInternal(ctx, "/v1/policy/decisions/simulate", internalToken, body)
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
	var wire evaluateResponse
	if err := decodeBoundedResponse(resp.Body, &wire, false); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return &Result{
		DecisionID: wire.DecisionID, PolicyVersion: wire.PolicyVersion,
		EvidenceRefs: wire.EvidenceRefs, Outcome: wire.Outcome,
	}, nil
}
