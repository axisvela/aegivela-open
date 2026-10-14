package pepsdk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// RiskSignalRequest is the input to SubmitRiskSignal (v1alpha1 contract).
type RiskSignalRequest struct {
	SubjectRef string
	SessionRef string
	RiskLevel  string
	RiskRef    string
	Reason     string
	TraceID    string
}

type riskSignalWire struct {
	APIVersion string `json:"api_version"`
	SubjectRef string `json:"subject_ref"`
	SessionRef string `json:"session_ref,omitempty"`
	RiskLevel  string `json:"risk_level"`
	RiskRef    string `json:"risk_ref"`
	Reason     string `json:"reason,omitempty"`
	TraceID    string `json:"trace_id"`
}

// SubmitRiskSignal posts a risk signal that may drive revocation. The caller
// presents the AEGIVELA internal token.
func (c *Client) SubmitRiskSignal(ctx context.Context, internalToken string, req RiskSignalRequest) error {
	if strings.TrimSpace(internalToken) == "" ||
		strings.TrimSpace(req.SubjectRef) == "" || strings.TrimSpace(req.RiskRef) == "" {
		return ErrInvalidInput
	}
	traceID := req.TraceID
	if traceID == "" {
		traceID = generateTraceID()
	}
	body, err := json.Marshal(riskSignalWire{
		APIVersion: APIVersionV1Alpha1, SubjectRef: req.SubjectRef, SessionRef: req.SessionRef,
		RiskLevel: req.RiskLevel, RiskRef: req.RiskRef, Reason: req.Reason, TraceID: traceID,
	})
	if err != nil {
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	resp, err := c.postInternal(ctx, "/v1/risk/signals", internalToken, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return ErrUnauthenticated
	case http.StatusBadRequest:
		return ErrInvalidInput
	case http.StatusServiceUnavailable:
		return ErrUnavailable
	}
	if resp.StatusCode >= 300 {
		return ErrUnavailable
	}
	return nil
}
