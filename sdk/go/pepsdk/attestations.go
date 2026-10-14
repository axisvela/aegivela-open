package pepsdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// AttestationRequest is the input to IssueAttestation (v1alpha1 contract).
type AttestationRequest struct {
	BearerToken       string
	ArtifactDigest    string
	ArtifactRef       json.RawMessage
	SignerIdentity    string
	SignerPublicKey   string
	WorkloadAssertion string
	EnvironmentClaims json.RawMessage
	Audience          string
	PolicyVersion     string
	TraceID           string
}

// AttestationResult is the issued attestation artifact.
type AttestationResult struct {
	SignedArtifact string
	JTI            string
}

type attestationIssueRequestWire struct {
	BearerToken       string          `json:"bearer_token"`
	ArtifactDigest    string          `json:"artifact_digest"`
	ArtifactRef       json.RawMessage `json:"artifact_ref"`
	SignerIdentity    string          `json:"signer_identity"`
	SignerPublicKey   string          `json:"signer_public_key"`
	WorkloadAssertion string          `json:"workload_assertion,omitempty"`
	EnvironmentClaims json.RawMessage `json:"environment_claims,omitempty"`
	Audience          string          `json:"audience"`
	PolicyVersion     string          `json:"policy_version"`
	TraceID           string          `json:"trace_id"`
}

type attestationIssueResponse struct {
	SignedArtifact string `json:"signed_artifact"`
	JTI            string `json:"jti"`
}

// IssueAttestation issues a signed attestation artifact. The caller presents the
// AEGIVELA internal token.
func (c *Client) IssueAttestation(ctx context.Context, internalToken string, req AttestationRequest) (*AttestationResult, error) {
	if strings.TrimSpace(internalToken) == "" ||
		strings.TrimSpace(req.BearerToken) == "" ||
		strings.TrimSpace(req.ArtifactDigest) == "" ||
		len(req.ArtifactRef) == 0 ||
		strings.TrimSpace(req.SignerIdentity) == "" ||
		strings.TrimSpace(req.SignerPublicKey) == "" ||
		strings.TrimSpace(req.Audience) == "" ||
		strings.TrimSpace(req.PolicyVersion) == "" {
		return nil, ErrInvalidInput
	}
	traceID := req.TraceID
	if traceID == "" {
		traceID = generateTraceID()
	}
	body, err := json.Marshal(attestationIssueRequestWire{
		BearerToken: req.BearerToken, ArtifactDigest: req.ArtifactDigest, ArtifactRef: req.ArtifactRef,
		SignerIdentity: req.SignerIdentity, SignerPublicKey: req.SignerPublicKey,
		WorkloadAssertion: req.WorkloadAssertion, EnvironmentClaims: req.EnvironmentClaims,
		Audience: req.Audience, PolicyVersion: req.PolicyVersion, TraceID: traceID,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/attestations/issue", bytes.NewReader(body))
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
	case http.StatusUnauthorized:
		return nil, ErrUnauthenticated
	case http.StatusBadRequest:
		return nil, ErrInvalidInput
	case http.StatusServiceUnavailable:
		return nil, ErrUnavailable
	case http.StatusOK, http.StatusCreated:
	default:
		return nil, ErrUnavailable
	}
	var wire attestationIssueResponse
	if err := decodeBoundedResponse(resp.Body, &wire, false); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	if strings.TrimSpace(wire.SignedArtifact) == "" {
		return nil, fmt.Errorf("%w: empty signed artifact", ErrUnavailable)
	}
	return &AttestationResult{SignedArtifact: wire.SignedArtifact, JTI: wire.JTI}, nil
}

// AttestationVerifyRequest is the input to VerifyAttestation.
type AttestationVerifyRequest struct {
	SignedArtifact string
	Issuer         string
	Audience       string
}

// VerifyAttestation verifies an attestation artifact; the result JSON is
// returned verbatim (its shape is AEGIVELA-owned).
func (c *Client) VerifyAttestation(ctx context.Context, internalToken string, req AttestationVerifyRequest) (json.RawMessage, error) {
	if strings.TrimSpace(internalToken) == "" ||
		strings.TrimSpace(req.SignedArtifact) == "" ||
		strings.TrimSpace(req.Issuer) == "" || strings.TrimSpace(req.Audience) == "" {
		return nil, ErrInvalidInput
	}
	body, err := json.Marshal(map[string]string{
		"signed_artifact": req.SignedArtifact, "issuer": req.Issuer, "audience": req.Audience,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/attestations/verify", bytes.NewReader(body))
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
	var raw json.RawMessage
	if err := decodeBoundedResponse(resp.Body, &raw, false); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return raw, nil
}

// AttestationJWKS returns the raw JSON Web Key Set for attestation verification.
func (c *Client) AttestationJWKS(ctx context.Context) (json.RawMessage, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/attestations/jwks", nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, ErrUnavailable
	}
	var raw json.RawMessage
	if err := decodeBoundedResponse(resp.Body, &raw, false); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return raw, nil
}
