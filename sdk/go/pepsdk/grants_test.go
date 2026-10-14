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

func TestIssueGrant(t *testing.T) {
	var mu sync.Mutex
	var path, token string
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var b map[string]any
		_ = json.NewDecoder(r.Body).Decode(&b)
		mu.Lock()
		path, token, body = r.URL.Path, r.Header.Get("X-Aegivela-Internal-Token"), b
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]string{"grant": "grant-jwt"})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, srv.Client())
	grant, err := c.IssueGrant(context.Background(), "itok", GrantIssueRequest{
		BearerToken: "btok", SignedAllowDecision: "signed", Scope: []string{"read"},
	})
	if err != nil {
		t.Fatalf("IssueGrant: %v", err)
	}
	if grant != "grant-jwt" {
		t.Errorf("grant = %q", grant)
	}
	mu.Lock()
	defer mu.Unlock()
	if path != "/v1/grants/issue" || token != "itok" {
		t.Errorf("path=%q token=%q", path, token)
	}
	if body["signed_allow_decision"] != "signed" {
		t.Errorf("body = %v", body)
	}
}

func TestVerifyGrant(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"apiVersion":"aegivela.io/v2","grantId":"g1","valid":true,"scope":["read"],"audience":"rheovela","expired":false,"revoked":false}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, srv.Client())
	res, err := c.VerifyGrant(context.Background(), "itok", GrantVerifyRequest{SignedGrant: "g", Audience: "rheovela", Scope: []string{"read"}})
	if err != nil {
		t.Fatalf("VerifyGrant: %v", err)
	}
	if res.GrantID != "g1" || !res.Valid {
		t.Errorf("result = %+v", res)
	}
	if body["apiVersion"] != APIVersionV2 || body["signedGrant"] != "g" {
		t.Errorf("body = %v", body)
	}
}

func TestGrantJWKS(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/grants/jwks" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"keys":[{"kty":"OKP"}]}`))
	}))
	defer srv.Close()
	c := NewClient(srv.URL, srv.Client())
	raw, err := c.GrantJWKS(context.Background())
	if err != nil {
		t.Fatalf("GrantJWKS: %v", err)
	}
	if !strings.Contains(string(raw), "keys") {
		t.Errorf("jwks = %s", raw)
	}
}
