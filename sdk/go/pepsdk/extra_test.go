package pepsdk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestIssueAttestation(t *testing.T) {
	var mu sync.Mutex
	var path, token string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		path, token = r.URL.Path, r.Header.Get("X-Aegivela-Internal-Token")
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"signed_artifact":"att-jwt","jti":"j1"}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, srv.Client())
	res, err := c.IssueAttestation(context.Background(), "itok", AttestationRequest{
		BearerToken: "btok", ArtifactDigest: "d1", ArtifactRef: json.RawMessage(`{"kind":"wasm"}`),
		SignerIdentity: "axond", SignerPublicKey: "pk", Audience: "rheovela", PolicyVersion: "v1",
	})
	if err != nil {
		t.Fatalf("IssueAttestation: %v", err)
	}
	if res.SignedArtifact != "att-jwt" || res.JTI != "j1" {
		t.Errorf("result = %+v", res)
	}
	mu.Lock()
	defer mu.Unlock()
	if path != "/v1/attestations/issue" || token != "itok" {
		t.Errorf("path=%q token=%q", path, token)
	}
}

func TestAttestationJWKSAndVerify(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`{"keys":[]}`))
			return
		}
		_, _ = w.Write([]byte(`{"valid":true}`))
	}))
	defer srv.Close()
	c := NewClient(srv.URL, srv.Client())
	if raw, err := c.AttestationJWKS(context.Background()); err != nil || !strings.Contains(string(raw), "keys") {
		t.Fatalf("AttestationJWKS raw=%s err=%v", raw, err)
	}
	if raw, err := c.VerifyAttestation(context.Background(), "itok", AttestationVerifyRequest{SignedArtifact: "a", Issuer: "i", Audience: "au"}); err != nil || !strings.Contains(string(raw), "valid") {
		t.Fatalf("VerifyAttestation raw=%s err=%v", raw, err)
	}
}

func TestCreateWorkloadBindingAndJWKS(t *testing.T) {
	var path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`{"keys":[]}`))
			return
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"BindingID":"b1"}`))
	}))
	defer srv.Close()
	c := NewClient(srv.URL, srv.Client())
	raw, err := c.CreateWorkloadBinding(context.Background(), "itok", CreateWorkloadBindingRequest{
		SubjectRef: "s", OrganizationRoot: "o", ClientID: "c", WorkloadRef: "w",
	})
	if err != nil {
		t.Fatalf("CreateWorkloadBinding: %v", err)
	}
	if !strings.Contains(string(raw), "b1") || path != "/v1/workload-bindings" {
		t.Errorf("raw=%s path=%s", raw, path)
	}
	if _, err := c.WorkloadAssertionJWKS(context.Background()); err != nil {
		t.Fatalf("WorkloadAssertionJWKS: %v", err)
	}
}
