package pepsdk

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestClientHeadersAreCopiedAndEndpointCredentialsWin(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Aegivela-PEP"); got != "configured" {
			t.Errorf("PEP header = %q", got)
		}
		if r.URL.Path == "/v1/revocations" {
			if got := r.Header.Get("X-Aegivela-Internal-Token"); got != "per-call" {
				t.Errorf("endpoint credential = %q", got)
			}
			_, _ = w.Write([]byte(`{"events":[]}`))
			return
		}
		if r.URL.Path == "/v1/policy/decisions/jwks" {
			_, _ = w.Write([]byte(`{"keys":[]}`))
			return
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("content type = %q", got)
		}
		_, _ = w.Write([]byte(`{"decision_id":"d1","outcome":"allow","policy_version":"v1"}`))
	}))
	defer srv.Close()
	headers := map[string]string{"X-Aegivela-PEP": "configured", "X-Aegivela-Internal-Token": "default", "Content-Type": "text/plain"}
	c := NewClient(srv.URL, srv.Client(), WithHeaders(headers))
	headers["X-Aegivela-PEP"] = "mutated"
	route := Route{Audience: "au", Action: "read", Scope: []string{"read"}, Kind: "tool", Reference: "t1"}
	input := AuthorizationInput{Mode: ModeHumanWeb, BearerToken: "bt"}
	if _, err := c.Evaluate(context.Background(), route, input); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Authorize(context.Background(), route, input); err != nil {
		t.Fatal(err)
	}
	if _, err := c.DecisionJWKS(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := c.RevocationList(context.Background(), "per-call", time.Time{}); err != nil {
		t.Fatal(err)
	}
}

func TestClientHeadersRefuseCrossOriginRedirect(t *testing.T) {
	var upstreamCalls atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamCalls.Add(1)
	}))
	defer upstream.Close()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, upstream.URL, http.StatusTemporaryRedirect)
	}))
	defer srv.Close()
	hc := srv.Client()
	c := NewClient(srv.URL, hc, WithHeaders(map[string]string{"X-Aegivela-Internal-Token": "secret"}))
	if _, err := c.DecisionJWKS(context.Background()); err == nil {
		t.Fatal("cross-origin redirect accepted")
	}
	if upstreamCalls.Load() != 0 {
		t.Fatal("configured credentials reached a different origin")
	}
	if hc.CheckRedirect != nil {
		t.Fatal("caller HTTP client was mutated")
	}
}

func TestClientHeadersPreserveSameOriginRedirectPolicy(t *testing.T) {
	var checks atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/final" {
			http.Redirect(w, r, "/final", http.StatusTemporaryRedirect)
			return
		}
		if r.Header.Get("X-Aegivela-PEP") != "configured" {
			t.Error("same-origin redirect lost headers")
		}
		_, _ = w.Write([]byte(`{"keys":[]}`))
	}))
	defer srv.Close()
	hc := srv.Client()
	hc.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		checks.Add(1)
		return nil
	}
	if _, err := NewClient(srv.URL, hc, WithHeaders(map[string]string{"X-Aegivela-PEP": "configured"})).DecisionJWKS(context.Background()); err != nil {
		t.Fatal(err)
	}
	if checks.Load() != 1 {
		t.Fatalf("redirect checks = %d", checks.Load())
	}
}
