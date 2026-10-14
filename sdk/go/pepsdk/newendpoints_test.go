package pepsdk

import (
	"context"
	"net/http"
	"net/http/httptest"
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
