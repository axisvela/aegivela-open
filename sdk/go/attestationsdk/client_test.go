package attestationsdk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIssueVerifyAndJWKS(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/attestations/issue":
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"signed_artifact":"att","jti":"j1"}`))
		case "/v1/attestations/verify":
			_, _ = w.Write([]byte(`{"valid":true}`))
		case "/v1/attestations/jwks":
			_, _ = w.Write([]byte(`{"keys":[]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	c := New(srv.URL, srv.Client())
	res, err := c.IssueAttestation(context.Background(), "itok", AttestationRequest{
		BearerToken: "b", ArtifactDigest: "d", ArtifactRef: json.RawMessage(`{}`),
		SignerIdentity: "s", SignerPublicKey: "pk", Audience: "au", PolicyVersion: "v1",
	})
	if err != nil || res.SignedArtifact != "att" {
		t.Fatalf("res=%+v err=%v", res, err)
	}
	if _, err := c.VerifyAttestation(context.Background(), "itok", AttestationVerifyRequest{SignedArtifact: "a", Issuer: "i", Audience: "au"}); err != nil {
		t.Fatalf("VerifyAttestation: %v", err)
	}
	jwks, err := c.AttestationJWKS(context.Background())
	if err != nil || !strings.Contains(string(jwks), "keys") {
		t.Fatalf("jwks=%s err=%v", jwks, err)
	}
}
