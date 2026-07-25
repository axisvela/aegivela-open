package pepsdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// SLOClass is the ADR-0008 revocation freshness class for a recheck.
type SLOClass string

const (
	ClassPreDispatch  SLOClass = "pre_dispatch"
	ClassContinuation SLOClass = "continuation"
	ClassConnection   SLOClass = "connection"
)

// RevocationSelector identifies one tenant-scoped revocation dimension.
type RevocationSelector struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

// CheckRevocation performs a revocation recheck at a declared boundary. pre_dispatch and
// continuation classes are authoritative cache-free checks; any error means fail closed.
func (c *Client) CheckRevocation(ctx context.Context, internalToken string, class SLOClass, selectors []RevocationSelector, lifecycleEpoch *int64) error {
	if class != ClassPreDispatch && class != ClassContinuation && class != ClassConnection {
		return ErrInvalidInput
	}
	if len(selectors) == 0 || len(selectors) > 16 || internalToken == "" {
		return ErrInvalidInput
	}
	bodyMap := map[string]any{
		"api_version": "aegivela.io/v1alpha1",
		"class":       string(class),
		"selectors":   selectors,
		"trace_id":    generateTraceID(),
	}
	if lifecycleEpoch != nil {
		bodyMap["lifecycle_epoch"] = *lifecycleEpoch
	}
	body, err := json.Marshal(bodyMap)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/revocations/check", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-AEGIVELA-PEP", internalToken)
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusUnauthorized:
		return ErrUnauthenticated
	default:
		return ErrUnavailable
	}
	var checkResp struct {
		Outcome string `json:"outcome"`
	}
	decoder := json.NewDecoder(io.LimitReader(resp.Body, maxResponseBytes))
	if err := decoder.Decode(&checkResp); err != nil {
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	if checkResp.Outcome != "clear" {
		return ErrDenied
	}
	return nil
}
