package pepsdk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestToolInvoke(t *testing.T) {
	var mu sync.Mutex
	var path, token string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		path, token = r.URL.Path, r.Header.Get("X-Aegivela-Internal-Token")
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"api_version":"aegivela.io/v1alpha1","decision_id":"d1","outcome":"allow","policy_version":"v1","tool_id":"t1","grant_token":"gt"}`))
	}))
	defer srv.Close()
	c := NewClient(srv.URL, srv.Client())
	res, err := c.ToolInvoke(context.Background(), "itok", ToolInvokeRequest{
		BearerToken: "btok", Resource: Resource{Kind: "tool", Reference: "t1"},
		Scope: []string{"read"}, ExecutionID: "e1", ToolID: "t1", SkillHash: "sh", ImplementationDigest: "id",
	})
	if err != nil {
		t.Fatalf("ToolInvoke: %v", err)
	}
	if res.GrantToken != "gt" || res.DecisionID != "d1" {
		t.Errorf("res = %+v", res)
	}
	mu.Lock()
	defer mu.Unlock()
	if path != "/v1/tool/invoke" || token != "itok" {
		t.Errorf("path=%q token=%q", path, token)
	}
}

func TestVerifyDecision(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"apiVersion":"aegivela.io/v2.0","decisionId":"d1","valid":true,"outcome":"allow","policyVersion":"v1","evidenceRefs":["e"]}`))
	}))
	defer srv.Close()
	c := NewClient(srv.URL, srv.Client())
	res, err := c.VerifyDecision(context.Background(), "itok", DecisionVerifyRequest{
		SignedDecision: "sd", Action: "tool:invoke", Resource: Resource{Kind: "tool", Reference: "t1"},
	})
	if err != nil {
		t.Fatalf("VerifyDecision: %v", err)
	}
	if !res.Valid || res.DecisionID != "d1" {
		t.Errorf("res = %+v", res)
	}
	if body["apiVersion"] != APIVersionV2 || body["signedDecision"] != "sd" {
		t.Errorf("body = %v", body)
	}
}

func TestDecisionJWKSAndRevocationList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/policy/decisions/jwks":
			_, _ = w.Write([]byte(`{"keys":[]}`))
		case "/v1/revocations":
			_, _ = w.Write([]byte(`{"events":[{"id":"r1"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	c := NewClient(srv.URL, srv.Client())
	if raw, err := c.DecisionJWKS(context.Background()); err != nil || !strings.Contains(string(raw), "keys") {
		t.Fatalf("DecisionJWKS raw=%s err=%v", raw, err)
	}
	events, err := c.RevocationList(context.Background(), "itok", time.Time{})
	if err != nil || len(events) != 1 {
		t.Fatalf("RevocationList events=%v err=%v", events, err)
	}
}
