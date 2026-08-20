package pepsdk

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGatewayConnectSuccess(t *testing.T) {
	var gotReq GatewayConnectRequest
	var gotToken string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotToken = r.Header.Get("X-AEGIVELA-PEP")
		if r.URL.Path != "/v1/gateway/connect" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		b, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(b, &gotReq); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(GatewayConnectDecision{
			DecisionID:    "gdec-1",
			Outcome:       "allow",
			GatewayToken:  "gw-1",
			PolicyVersion: "pv-1",
			ResolvedHost:  "api.example.com",
			ResolvedScheme: "https",
			ResolvedPath:  "/v1/orders",
			ExpiresAt:     time.Now().Add(time.Hour),
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, srv.Client())
	req := GatewayConnectRequest{
		Action:            "gateway.connect",
		WorkloadAssertion: "wl-assert",
		TargetHost:        "api.example.com",
		ToolAttribution: &ToolAttribution{
			HumanRef:  "human-1",
			AgentRef:  "agent-1",
			ToolID:    "tool-1",
			SkillHash: "sha256:abc",
		},
	}
	dec, err := c.GatewayConnect(context.Background(), "pep-internal", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dec.DecisionID != "gdec-1" || dec.GatewayToken != "gw-1" || dec.ResolvedHost != "api.example.com" {
		t.Errorf("unexpected decision: %+v", dec)
	}
	if gotToken != "pep-internal" {
		t.Errorf("expected X-AEGIVELA-PEP=pep-internal, got %q", gotToken)
	}
	if gotReq.TargetHost != "api.example.com" || gotReq.TraceID == "" || gotReq.ToolAttribution == nil {
		t.Errorf("request not marshaled correctly: %+v", gotReq)
	}
}

func TestGatewayConnectDenied(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(GatewayConnectDecision{Outcome: "deny"})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, srv.Client())
	_, err := c.GatewayConnect(context.Background(), "pep", GatewayConnectRequest{
		Action:            "gateway.connect",
		WorkloadAssertion: "wl",
		TargetHost:        "h",
	})
	if !errors.Is(err, ErrDenied) {
		t.Fatalf("expected ErrDenied, got %v", err)
	}
}

func TestGatewayConnectStatusMapping(t *testing.T) {
	cases := []struct {
		status int
		want   error
	}{
		{http.StatusUnauthorized, ErrUnauthenticated},
		{http.StatusForbidden, ErrDenied},
		{http.StatusInternalServerError, ErrUnavailable},
	}
	for _, tc := range cases {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(tc.status)
		}))
		c := NewClient(srv.URL, srv.Client())
		_, err := c.GatewayConnect(context.Background(), "pep", GatewayConnectRequest{
			Action: "gateway.connect", WorkloadAssertion: "wl", TargetHost: "h",
		})
		srv.Close()
		if !errors.Is(err, tc.want) {
			t.Fatalf("status %d: expected %v, got %v", tc.status, tc.want, err)
		}
	}
}

func TestGatewayConnectInvalidInput(t *testing.T) {
	c := NewClient("http://example.invalid", nil)
	if _, err := c.GatewayConnect(context.Background(), "", GatewayConnectRequest{
		Action: "gateway.connect", WorkloadAssertion: "wl", TargetHost: "h",
	}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for empty token, got %v", err)
	}
	if _, err := c.GatewayConnect(context.Background(), "pep", GatewayConnectRequest{
		WorkloadAssertion: "wl", TargetHost: "h",
	}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for empty action, got %v", err)
	}
	if _, err := c.GatewayConnect(context.Background(), "pep", GatewayConnectRequest{
		Action: "gateway.connect", TargetHost: "h",
	}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for empty workload, got %v", err)
	}
}
