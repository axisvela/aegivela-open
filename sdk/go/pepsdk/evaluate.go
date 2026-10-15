package pepsdk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// EvaluateResult is the raw evaluate outcome. Unlike Authorize, a deny is
// returned in Outcome rather than as an error, and the signed-decision artifact
// is exposed for verification.
type EvaluateResult struct {
	DecisionID     string
	Outcome        string
	PolicyVersion  string
	EvidenceRefs   []string
	ExpiresAt      time.Time
	Obligations    []string
	SignedDecision string
}

// Evaluate posts the policy evaluate request and returns the raw outcome
// (allow | deny | approval_required | …). Only transport/unavailable failures
// are errors; a deny is a normal result.
func (c *Client) Evaluate(ctx context.Context, route Route, input AuthorizationInput) (*EvaluateResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
	if err := route.Validate(); err != nil {
		return nil, err
	}
	apiVersion := input.APIVersion
	if apiVersion == "" {
		apiVersion = APIVersionV1Alpha1
	}
	if apiVersion == APIVersionV2 && (input.Mode == ModeServiceAgentAPI || input.Mode == ModeTwinAgentAPI) {
		return nil, ErrInvalidInput
	}
	traceID := generateTraceID()
	requestedExpiresAt := time.Now().Add(5 * time.Minute)

	var reqBody any
	if apiVersion == APIVersionV2 {
		reqBody = evaluateRequestV2{
			APIVersion: apiVersion, AuthorizationMode: modeV2(input.Mode), Principal: buildV2Principal(input),
			BearerToken: input.BearerToken, WorkloadAssertion: input.WorkloadAssertion,
			ParentExecutionGrant: input.ParentExecutionGrant, Action: route.Action,
			Resource: Resource{Kind: route.Kind, Reference: route.Reference}, RequestedScope: route.Scope,
			Audience: route.Audience, RequestedExpiresAt: requestedExpiresAt,
			TaskBinding: input.TaskBinding, TraceID: traceID, RiskContext: buildRiskContext(input),
		}
	} else {
		reqBody = evaluateRequest{
			APIVersion: apiVersion, AuthorizationMode: input.Mode, Principal: minimalPrincipal{},
			BearerToken: input.BearerToken, WorkloadAssertion: input.WorkloadAssertion,
			ParentExecutionGrant: input.ParentExecutionGrant, Action: route.Action,
			Resource: Resource{Kind: route.Kind, Reference: route.Reference}, RequestedScope: route.Scope,
			Audience: route.Audience, RequestedExpiresAt: requestedExpiresAt,
			TaskBinding: input.TaskBinding, TraceID: traceID,
		}
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	resp, err := c.hc.Request(ctx, http.MethodPost, "/v1/policy/decisions/evaluate", nil, body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return nil, ErrUnauthenticated
	case http.StatusServiceUnavailable:
		return nil, ErrUnavailable
	case http.StatusOK:
	default:
		return nil, ErrUnavailable
	}

	if apiVersion == APIVersionV2 {
		var wire evaluateResponseV2
		if err := decodeBoundedResponse(resp.Body, &wire, false); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
		}
		outcome := wire.Result
		if outcome == "" {
			outcome = wire.Outcome
		}
		return &EvaluateResult{
			DecisionID: wire.DecisionID, Outcome: outcome, PolicyVersion: wire.PolicyVersion,
			EvidenceRefs: wire.EvidenceRefs, ExpiresAt: wire.ExpiresAt, Obligations: wire.Obligations,
			SignedDecision: firstNonEmpty(wire.SignedDecision, wire.SignedDecisionLegacy),
		}, nil
	}

	var wire evaluateResponse
	if err := decodeBoundedResponse(resp.Body, &wire, false); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return &EvaluateResult{
		DecisionID: wire.DecisionID, Outcome: wire.Outcome, PolicyVersion: wire.PolicyVersion,
		EvidenceRefs: wire.EvidenceRefs,
		SignedDecision: firstNonEmpty(wire.SignedDecision, wire.SignedDecisionLegacy),
	}, nil
}
