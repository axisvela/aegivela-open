package grantsdk

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestWithRetryRetriesGrantJWKS(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) < 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte(`{"keys":[]}`))
	}))
	defer srv.Close()

	c := New(srv.URL, srv.Client(), WithRetry(3))
	if _, err := c.GrantJWKS(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Fatalf("calls = %d, want 2", got)
	}
}

func TestClientHeadersAreCopiedAndEndpointCredentialsWin(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Aegivela-PEP"); got != "configured" {
			t.Errorf("PEP header = %q", got)
		}
		switch r.URL.Path {
		case "/v1/grants/issue":
			if got := r.Header.Get("X-Aegivela-Internal-Token"); got != "per-call" {
				t.Errorf("endpoint credential = %q", got)
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"grant":"g"}`))
		case "/v1/grants/verify":
			if got := r.Header.Get("X-Aegivela-Internal-Token"); got != "per-call" {
				t.Errorf("endpoint credential = %q", got)
			}
			_, _ = w.Write([]byte(`{"apiVersion":"aegivela.io/v2.0","valid":true}`))
		case "/v1/grants/jwks":
			_, _ = w.Write([]byte(`{"keys":[]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	headers := map[string]string{"X-Aegivela-PEP": "configured", "X-Aegivela-Internal-Token": "default"}
	c := New(srv.URL, srv.Client(), WithHeaders(headers))
	headers["X-Aegivela-PEP"] = "mutated"

	if _, err := c.IssueGrant(context.Background(), "per-call", GrantIssueRequest{BearerToken: "b", SignedAllowDecision: "s"}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.VerifyGrant(context.Background(), "per-call", GrantVerifyRequest{SignedGrant: "g", Audience: "au", Scope: []string{"read"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.GrantJWKS(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestClientHeadersRefuseCrossOriginRedirect(t *testing.T) {
	upstreamCalls := 0
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamCalls++
	}))
	defer upstream.Close()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, upstream.URL, http.StatusTemporaryRedirect)
	}))
	defer srv.Close()

	c := New(srv.URL, srv.Client(), WithHeaders(map[string]string{"X-Aegivela-Internal-Token": "secret"}))
	if _, err := c.GrantJWKS(context.Background()); err == nil {
		t.Fatal("cross-origin redirect accepted")
	}
	if upstreamCalls != 0 {
		t.Fatal("configured credentials reached a different origin")
	}
}
