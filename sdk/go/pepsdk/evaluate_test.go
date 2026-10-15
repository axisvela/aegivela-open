package pepsdk

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEvaluateRawOutcome(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/policy/decisions/evaluate" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"api_version":"aegivela.io/v1alpha1","decision_id":"d1","outcome":"approval_required","policy_version":"v1","evidence_refs":["e"],"signed_decision":"jws"}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, srv.Client())
	res, err := c.Evaluate(context.Background(),
		Route{Audience: "au", Action: "tool:invoke", Scope: []string{"read"}, Kind: "tool", Reference: "t1"},
		AuthorizationInput{Mode: ModeHumanWeb, BearerToken: "bt"})
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if res.Outcome != "approval_required" {
		t.Fatalf("Outcome = %q, want approval_required", res.Outcome)
	}
	if res.SignedDecision != "jws" {
		t.Fatalf("SignedDecision = %q", res.SignedDecision)
	}
}
