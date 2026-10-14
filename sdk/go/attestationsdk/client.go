// Package attestationsdk is the Apache-2.0 AEGIVELA attestation client: issue,
// verify, and the attestation JWKS. It wraps the shared pepsdk client.
package attestationsdk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/axisrobo/aegivela-open/sdk/go/pepsdk"
)

// Client is the attestation client.
type Client struct {
	PE *pepsdk.Client
}

// New creates an attestation client.
func New(baseURL string, httpClient *http.Client) *Client {
	return &Client{PE: pepsdk.NewClient(baseURL, httpClient)}
}

func internalHeader(token string) map[string]string {
	return map[string]string{"X-Aegivela-Internal-Token": token}
}

// AttestationRequest is the input to IssueAttestation (v1alpha1).
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

// IssueAttestation issues a signed attestation artifact.
func (c *Client) IssueAttestation(ctx context.Context, internalToken string, req AttestationRequest) (*AttestationResult, error) {
	if strings.TrimSpace(internalToken) == "" || strings.TrimSpace(req.BearerToken) == "" ||
		strings.TrimSpace(req.ArtifactDigest) == "" || len(req.ArtifactRef) == 0 ||
		strings.TrimSpace(req.SignerIdentity) == "" || strings.TrimSpace(req.SignerPublicKey) == "" ||
		strings.TrimSpace(req.Audience) == "" || strings.TrimSpace(req.PolicyVersion) == "" {
		return nil, pepsdk.ErrInvalidInput
	}
	traceID := req.TraceID
	if traceID == "" {
		traceID = pepsdk.GenerateTraceID()
	}
	body, err := json.Marshal(attestationIssueRequestWire{
		BearerToken: req.BearerToken, ArtifactDigest: req.ArtifactDigest, ArtifactRef: req.ArtifactRef,
		SignerIdentity: req.SignerIdentity, SignerPublicKey: req.SignerPublicKey,
		WorkloadAssertion: req.WorkloadAssertion, EnvironmentClaims: req.EnvironmentClaims,
		Audience: req.Audience, PolicyVersion: req.PolicyVersion, TraceID: traceID,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", pepsdk.ErrUnavailable, err)
	}
	resp, err := c.PE.Request(ctx, http.MethodPost, "/v1/attestations/issue", internalHeader(internalToken), body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", pepsdk.ErrUnavailable, err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return nil, pepsdk.ErrUnauthenticated
	case http.StatusBadRequest:
		return nil, pepsdk.ErrInvalidInput
	case http.StatusServiceUnavailable:
		return nil, pepsdk.ErrUnavailable
	case http.StatusOK, http.StatusCreated:
	default:
		return nil, pepsdk.ErrUnavailable
	}
	var wire attestationIssueResponse
	if err := pepsdk.DecodeJSON(resp.Body, &wire, false); err != nil {
		return nil, fmt.Errorf("%w: %v", pepsdk.ErrUnavailable, err)
	}
	if strings.TrimSpace(wire.SignedArtifact) == "" {
		return nil, fmt.Errorf("%w: empty signed artifact", pepsdk.ErrUnavailable)
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
// returned verbatim.
func (c *Client) VerifyAttestation(ctx context.Context, internalToken string, req AttestationVerifyRequest) (json.RawMessage, error) {
	if strings.TrimSpace(internalToken) == "" || strings.TrimSpace(req.SignedArtifact) == "" ||
		strings.TrimSpace(req.Issuer) == "" || strings.TrimSpace(req.Audience) == "" {
		return nil, pepsdk.ErrInvalidInput
	}
	body, err := json.Marshal(map[string]string{
		"signed_artifact": req.SignedArtifact, "issuer": req.Issuer, "audience": req.Audience,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", pepsdk.ErrUnavailable, err)
	}
	resp, err := c.PE.Request(ctx, http.MethodPost, "/v1/attestations/verify", internalHeader(internalToken), body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", pepsdk.ErrUnavailable, err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return nil, pepsdk.ErrUnauthenticated
	case http.StatusForbidden:
		return nil, pepsdk.ErrDenied
	case http.StatusBadRequest:
		return nil, pepsdk.ErrInvalidInput
	case http.StatusServiceUnavailable:
		return nil, pepsdk.ErrUnavailable
	case http.StatusOK:
	default:
		return nil, pepsdk.ErrUnavailable
	}
	var raw json.RawMessage
	if err := pepsdk.DecodeJSON(resp.Body, &raw, false); err != nil {
		return nil, fmt.Errorf("%w: %v", pepsdk.ErrUnavailable, err)
	}
	return raw, nil
}

// AttestationJWKS returns the raw JWKS for attestation verification.
func (c *Client) AttestationJWKS(ctx context.Context) (json.RawMessage, error) {
	resp, err := c.PE.Request(ctx, http.MethodGet, "/v1/attestations/jwks", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", pepsdk.ErrUnavailable, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, pepsdk.ErrUnavailable
	}
	var raw json.RawMessage
	if err := pepsdk.DecodeJSON(resp.Body, &raw, false); err != nil {
		return nil, fmt.Errorf("%w: %v", pepsdk.ErrUnavailable, err)
	}
	return raw, nil
}
