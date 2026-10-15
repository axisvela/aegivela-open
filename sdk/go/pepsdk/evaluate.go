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
// (allow | deny | approval_required | ...). A valid deny response is a normal
// result; invalid input, HTTP failures and malformed responses are errors.
func (c *Client) Evaluate(ctx context.Context, route Route, input AuthorizationInput) (*EvaluateResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
	if err := route.Validate(); err != nil {
		return nil, err
	}
	if len(route.Modes) > 0 {
		allowed := false
		for _, mode := range route.Modes {
			if mode == input.Mode {
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
		return nil, fmt.Errorf("%w: %w", ErrUnavailable, err)
	}
	defer resp.Body.Close()
	if apiVersion == APIVersionV2 && resp.StatusCode != http.StatusOK {
		return nil, &APIError{StatusCode: resp.StatusCode, Err: decodeV2Error(resp)}
	}
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return nil, &APIError{StatusCode: resp.StatusCode, Err: ErrUnauthenticated}
	case http.StatusForbidden:
		return nil, &APIError{StatusCode: resp.StatusCode, Err: ErrDenied}
	case http.StatusOK:
	default:
		return nil, &APIError{StatusCode: resp.StatusCode, Err: ErrUnavailable}
	}

	if apiVersion == APIVersionV2 {
		return decodeV2Evaluate(resp)
	}

	var wire evaluateResponse
	if err := decodeBoundedResponse(resp.Body, &wire, false); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrMalformed, err)
	}
	if wire.DecisionID == "" || wire.PolicyVersion == "" {
		return nil, ErrMalformed
	}
	return &EvaluateResult{
		DecisionID: wire.DecisionID, Outcome: wire.Outcome, PolicyVersion: wire.PolicyVersion,
		EvidenceRefs:   wire.EvidenceRefs,
		SignedDecision: firstNonEmpty(wire.SignedDecision, wire.SignedDecisionLegacy),
	}, nil
}
