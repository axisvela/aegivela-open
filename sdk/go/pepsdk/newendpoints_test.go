package pepsdk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSimulateDecision(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/policy/decisions/simulate" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"decision_id":"d1","outcome":"allow","policy_version":"v1","evidence_refs":["e"]}`))
	}))
	defer srv.Close()
	c := NewClient(srv.URL, srv.Client())
	res, err := c.SimulateDecision(context.Background(), "itok",
		AuthorizationInput{Mode: ModeHumanWeb, BearerToken: "bt"},
		Route{Audience: "au", Action: "tool:invoke", Scope: []string{"read"}, Kind: "tool", Reference: "t1"})
	if err != nil || res.Outcome != "allow" || res.DecisionID != "d1" {
		t.Fatalf("res=%+v err=%v", res, err)
	}
}

func TestCreateWebResourceWrite(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/web-resources" || r.Method != http.MethodPost {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"resource_id":"r1"}`))
	}))
	defer srv.Close()
	c := NewClient(srv.URL, srv.Client())
	raw, err := c.CreateWebResource(context.Background(), "itok", CreateWebResourceRequest{
		ProductID: "moduregis", ResourceID: "r1", Kind: "resource", DisplayName: "R1",
	})
	if err != nil || !strings.Contains(string(raw), "r1") {
		t.Fatalf("raw=%s err=%v", raw, err)
	}
	if body["api_version"] != APIVersionV1Alpha1 {
		t.Errorf("body = %v", body)
	}
}
