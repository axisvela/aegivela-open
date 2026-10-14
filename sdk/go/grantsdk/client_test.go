package grantsdk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIssueAndVerifyGrant(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/grants/issue":
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"grant":"grant-jwt"}`))
		case "/v1/grants/verify":
			_, _ = w.Write([]byte(`{"apiVersion":"aegivela.io/v2.0","grantId":"g1","valid":true}`))
		case "/v1/grants/jwks":
			_, _ = w.Write([]byte(`{"keys":[]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	c := New(server.URL, server.Client())

	grant, err := c.IssueGrant(context.Background(), "itok", GrantIssueRequest{BearerToken: "b", SignedAllowDecision: "s"})
	if err != nil || grant != "grant-jwt" {
		t.Fatalf("grant=%q err=%v", grant, err)
	}
	res, err := c.VerifyGrant(context.Background(), "itok", GrantVerifyRequest{SignedGrant: "g", Audience: "au", Scope: []string{"read"}})
	if err != nil || !res.Valid || res.GrantID != "g1" {
		t.Fatalf("res=%+v err=%v", res, err)
	}
	jwks, err := c.GrantJWKS(context.Background())
	if err != nil || !strings.Contains(string(jwks), "keys") {
		t.Fatalf("jwks=%s err=%v", jwks, err)
	}
	_ = json.Marshal
}
